package postgress

import (
	models "Skillfactory/36-GoNews/pkg/storage/models"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

type Storage struct {
	Db *pgxpool.Pool
}

// Storage конструктор. Пароль БД загружается из переменной окружения.
func New() (*Storage, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("cant loading .env file")
		return nil, err
	}
	pwd := os.Getenv("DBPASSWORD")

	connString := "postgres://postgres:" + pwd + "@localhost:5432/gonews"

	db, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		log.Printf("cant create new instance of DB: %v\n", err)
		return nil, err
	}
	s := Storage{
		Db: db,
	}
	return &s, nil
}

// конструктор тестовой БД новостей
func NewMock() (*Storage, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("cant loading .env file")
		return nil, err
	}
	pwd := os.Getenv("DBPASSWORD")

	connString := "postgres://postgres:" + pwd + "@localhost:5432/gonewsmock"

	db, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		log.Printf("cant create new instance of DB: %v\n", err)
		return nil, err
	}
	s := Storage{
		Db: db,
	}
	return &s, nil
}

// Метод получения статей из базы данных. Принимает количество, необходимых к возврату статей. Возвращает слайс обхектов или ошибку.
func (s *Storage) GetDetailedNews(id int) (models.NewsFullDetailed, error) {
	if id < 1 {
		err := fmt.Errorf("invalid news ID - got %v", id)
		log.Println(err)
		e := errors.New("invalid news ID")
		return models.NewsFullDetailed{}, e
	}

	q := strconv.Itoa(id)
	rows, err := s.Db.Query(context.Background(), `SELECT id,title,content,published,link FROM news WHERE id = $1`, q)
	if err != nil {
		log.Printf("Cant read data from database: %v\n", err)
		return models.NewsFullDetailed{}, err
	}
	defer rows.Close()
	news := models.NewsFullDetailed{}
	for rows.Next() {

		err = rows.Scan(
			&news.ID,
			&news.Title,
			&news.Content,
			&news.Published,
			&news.Link,
		)
		if err != nil {
			return models.NewsFullDetailed{}, fmt.Errorf("unable scan row: %w", err)
		}

	}
	return news, nil
}

// Метод получения из БД списка новостей. n - количество новостей для возврата.
func (s *Storage) GetNewsList(n int) ([]models.NewsFullDetailed, error) {
	if n < 1 {
		err := fmt.Errorf("Error!Invalid count of new - got %v", n)
		log.Println(err)
		e := errors.New("invalid count of news")
		return nil, e
	}
	if n == 0 {
		n = 10
	}
	q := strconv.Itoa(n)

	rows, err := s.Db.Query(context.Background(), `SELECT id,title,preview,published,link FROM news ORDER BY published DESC LIMIT $1`, q)
	if err != nil {
		log.Printf("Cant read data from database: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	news := []models.NewsFullDetailed{}

	for rows.Next() {
		new := models.NewsFullDetailed{}
		err = rows.Scan(
			&new.ID,
			&new.Title,
			&new.Preview,
			&new.Published,
			&new.Link,
		)
		if err != nil {
			return nil, fmt.Errorf("unable scan row: %w", err)
		}
		news = append(news, new)
	}

	return news, nil
}

// Метод для возврата списка новостей с пагинацией
func (s *Storage) GetNewsListWithPagination(n, offset, limit int) ([]models.NewsFullDetailed, error) {
	query := `WITH subquery AS (SELECT id, title, preview, published, link FROM news ORDER BY published LIMIT $1)
	SELECT id, title, preview, published, link FROM subquery OFFSET $2 LIMIT $3`
	rows, err := s.Db.Query(context.Background(), query, n, offset, limit)
	if err != nil {
		log.Printf("Cant read data from database: %v\n", err)
	}
	defer rows.Close()

	var news []models.NewsFullDetailed

	for rows.Next() {
		new := models.NewsFullDetailed{}
		err = rows.Scan(
			&new.ID,
			&new.Title,
			&new.Preview,
			&new.Published,
			&new.Link,
		)
		if err != nil {
			return news, fmt.Errorf("unable scan row: %w", err)
		}
		news = append(news, new)
	}
	return news, nil
}

// Функция для создания превью новости
func PrevieMaker(detailNews string) string {
	var preview []rune
	runes := []rune(detailNews)
	if len(runes) >= 100 {
		preview = runes[:len(runes)/4]
	}
	if len(runes) < 100 {
		preview = runes[:len(runes)/2]
	}
	return string(preview) + "..."
}

// Метод добавления статьи в БД. На вход принимает слайс объектов, возвращает ошибку, при наличии.
func (s *Storage) AddNews(news []models.NewsFullDetailed) error {
	for _, n := range news {
		n.Preview = PrevieMaker(n.Content)
		_, err := s.Db.Exec(context.Background(), `INSERT INTO news 
		(title,content,preview,published,link) VALUES ($1,$2,$3,$4,$5);`,
			n.Title, n.Content, n.Preview, n.Published, n.Link)
		if err != nil {
			log.Printf("Cant add data in database! %v\n", err)
			return err
		}
	}
	return nil
}

// Метод для выборки из БД новостей с учетом заданного фильтра
func (s *Storage) FilterNewsByContent(filter string) ([]models.NewsFullDetailed, error) {
	query := `SELECT id,title,preview,published,link FROM news WHERE 
              LOWER(content) LIKE $1 OR LOWER(title) LIKE $1 OR LOWER(preview) LIKE $1 ORDER BY published DESC;`
	rows, err := s.Db.Query(context.Background(), query, "%"+strings.ToLower(filter)+"%")
	if err != nil {
		log.Printf("cant read filtered data from database: %v\n", err)
	}
	defer rows.Close()

	news := []models.NewsFullDetailed{}
	for rows.Next() {
		new := models.NewsFullDetailed{}
		err = rows.Scan(
			&new.ID,
			&new.Title,
			&new.Preview,
			&new.Published,
			&new.Link,
		)
		if err != nil {
			return nil, fmt.Errorf("unable scan row: %w", err)
		}
		news = append(news, new)
	}
	return news, nil
}

// Метод для выборки из БД новостей с учетом заданного фильтра и пагинацией
func (s *Storage) FilterNewsByContentWithPagination(filter string, offset, limit int) ([]models.NewsFullDetailed, error) {
	query := `WITH subquery AS (SELECT id, title, preview, published, link FROM news WHERE LOWER(content) LIKE $1 OR LOWER(title) LIKE $1
	OR LOWER(preview) LIKE $1 ORDER BY published DESC) SELECT id, title, preview, published, link FROM subquery OFFSET $2 LIMIT $3;`
	rows, err := s.Db.Query(context.Background(), query, ("%" + strings.ToLower(filter) + "%"), offset, limit)
	if err != nil {
		log.Printf("cant read filtered data from database: %v\n", err)
	}
	defer rows.Close()

	news := []models.NewsFullDetailed{}
	for rows.Next() {
		new := models.NewsFullDetailed{}
		err = rows.Scan(
			&new.ID,
			&new.Title,
			&new.Preview,
			&new.Published,
			&new.Link,
		)
		if err != nil {
			return nil, fmt.Errorf("unable scan row: %w", err)
		}
		news = append(news, new)
	}
	return news, nil
}

// Метод для выборки из БД новостей с учетом заданного фильтра по дате публикации
func (s *Storage) FilterNewsByPublished(filter int) ([]models.NewsFullDetailed, error) {
	q := strconv.Itoa(filter)
	rows, err := s.Db.Query(context.Background(), `SELECT id,title,preview,published,link FROM news
	 WHERE published = $1;`, q)
	if err != nil {
		log.Printf("cant read filtered data from database: %v\n", err)
	}
	defer rows.Close()

	news := []models.NewsFullDetailed{}

	for rows.Next() {
		new := models.NewsFullDetailed{}
		err = rows.Scan(
			&new.ID,
			&new.Title,
			&new.Preview,
			&new.Published,
			&new.Link,
		)
		if err != nil {
			return nil, fmt.Errorf("unable scan row: %w", err)
		}
		news = append(news, new)
	}
	return news, nil
}
