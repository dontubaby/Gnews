package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"Skillfactory/36-GoNews/pkg/api"
	"Skillfactory/36-GoNews/pkg/rss"

	DB "Skillfactory/36-GoNews/pkg/storage"
	"Skillfactory/36-GoNews/pkg/storage/models"
	"Skillfactory/36-GoNews/pkg/storage/postgress"

	kfk "github.com/dontubaby/kafka_wrapper"
	middleware "github.com/dontubaby/mware"
)

// Объект с настройками приложения
type Config struct {
	RSSsources []string `json:"source"`
	Interval   int      `json:"interval"`
	Brokers    []string `json:"brokers"`
	Topic      []string `json:"topic"`
}

// Функция - конвертер JSON файла с настройками в объект с настройками приложения.
func ParseConfigFile(filename string) (Config, error) {
	var data []byte
	configFile, err := os.Open(filename)
	if err != nil {
		log.Printf("Open file error - %v", err)
		return Config{}, err
	}
	defer configFile.Close()

	data, err = ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	var config Config
	jsonErr := json.Unmarshal(data, &config)
	if jsonErr != nil {
		log.Printf("Unmarhaling error - %v", err)
		return Config{}, err
	}
	return config, nil
}

// Функция асинхронной обработки RSS-лент. На вход принимает источник RSS, интерфейс БД, канал для записи статей, канал для записи ошибок парсинга RSS
func AsynParser(ctx context.Context, source string, db DB.DbInterface, news chan<- []models.NewsFullDetailed, errs chan<- error, interval int) {
	for {
		rssnews, err := rss.Parse(source)
		if err != nil {
			errs <- err
			continue
		}
		news <- rssnews
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}

// Функция принимающая сообщения из Кафки и создаеющая редирект в локалхост для срабатывания соответствующего хэндлера
// (в зависимости от полученного сообщения)
func SendRequestToLocalhost(path string) ([]byte, error) {
	url := fmt.Sprintf("http://localhost:3000%s", path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v\n", err)
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка отправки запроса: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения ответа: %v\n", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected response code: %d\n", resp.StatusCode)
	}
	return body, nil
}

func main() {
	ctxmain := context.Background()
	//Подключение к новостной БД
	pool, err := postgress.New()
	if err != nil {
		log.Printf("Error DB connection - %v", err)
	}
	defer pool.Db.Close()
	//Парсинг конфигурационного файла
	config, err := ParseConfigFile("config.json")
	if err != nil {
		log.Printf("Config decoding error - %v", err)
	}
	//Инициализация API
	api := api.New(pool)
	//Инициализация консьюмера Кафки считывающего входящие сообщения новостного сервиса
	c, err := kfk.NewConsumer([]string{"localhost:9093"}, "news_input")
	if err != nil {
		log.Printf("Kafka consumer creating error - %v", err)
	}
	//Инициализация продьюсера Кафки отправляющего сообщения во внешние сервисы
	p, err := kfk.NewProducer(config.Brokers)
	log.Printf("Producer create!Broker:%v\n", config.Brokers)
	if err != nil {
		log.Printf("Kafka producer creating error - %v", err)
	}
	//Канал для записи в новых новостей для последующего добавления в БД
	newsStream := make(chan []models.NewsFullDetailed)
	//Канал для записи ошибок парсинга
	errorStream := make(chan error)

	for _, source := range config.RSSsources {
		go AsynParser(ctxmain, source, pool, newsStream, errorStream, config.Interval)
	}
	//горутина для считывания новостей из канала и добавления их в БД
	go func() {
		for new := range newsStream {
			pool.AddNews(new)
		}
	}()
	//горутина для считывания ошибок парсинга и логирования
	go func() {
		for err := range errorStream {
			log.Println("Error:", err)
		}
	}()

	//TODO:need refactoring and simplify
	go func() {
		for {
			log.Println("Start getting messages and redirecting")
			msg, err := c.GetMessages(ctxmain)
			if err != nil {
				log.Printf("error when reading message fron Kafka - %v", err)
			}
			data, err := SendRequestToLocalhost(string(msg.Value))
			if err != nil {
				log.Printf("error reading data from Kafka message - %v", err)
				return
			}
			//Если входящие сообщение содержит часть пути /newsdetail/ отправляем его в топик newsdetail
			if strings.Contains(string(msg.Value), "/newsdetail/") {
				err = p.SendMessage(ctxmain, config.Topic[1], data)
				if err != nil {
					log.Printf("error when writing message in Kafka - %v", err)
					return
				}
			}
			//Если входящие сообщение содержит часть пути /newslist/ отправляем его в топик newslist
			if strings.Contains(string(msg.Value), "/newslist/?n=") {
				err = p.SendMessage(ctxmain, config.Topic[2], data)
				if err != nil {
					log.Printf("error when writing message in Kafka - %v", err)
					return
				}
			}
			//Если входящие сообщение содержит часть пути /newslist/filtered/ отправляем его в топик отфильтрованных
			//по контенту новостей
			if strings.Contains(string(msg.Value), "/newslist/filtered/?s=") {
				err = p.SendMessage(ctxmain, config.Topic[3], data)
				if err != nil {
					log.Printf("error when writing message in Kafka - %v", err)
					return
				}
			}
			//Если входящие сообщение содержит часть пути /newslist/filtered/date/ отправляем его в топик отфильтрованных
			//по дате новостей
			if strings.Contains(string(msg.Value), "/newslist/filtered/date/?date=") {
				err = p.SendMessage(ctxmain, config.Topic[4], data)
				if err != nil {
					log.Printf("error when writing message in Kafka - %v", err)
					return
				}
			}
		}
	}()
	//Создаем роутер и подключаем к нему middleware
	router := api.Router()
	router.Use(middleware.RequestIDMiddleware, middleware.LoggingMiddleware)

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("cant loading .env file")
	}
	port := os.Getenv("PORT")

	log.Printf("Server gonews APP start working at port %v", port)
	err = http.ListenAndServe(port, router)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
