package DB

import (
	"Skillfactory/36-GoNews/pkg/storage/models"
	"log"
)

// Интерфейс базы данных
type DbInterface interface {
	GetDetailedNews(int) (models.NewsFullDetailed, error)
	GetNewsList(int) ([]models.NewsFullDetailed, error)
	AddNews([]models.NewsFullDetailed) error
	//TODO:add method GetNewsList
}

// Метод вовзрата статей
func GetDetailedNews(id int, db DbInterface) (models.NewsFullDetailed, error) {
	result, err := db.GetDetailedNews(id)
	if err != nil {
		log.Fatalf("Error when GET articles from server: %v\n", err)
		return models.NewsFullDetailed{}, err
	}

	return result, nil
}

// Метод добавления статьи
func Add(db DbInterface, news []models.NewsFullDetailed) error {
	err := db.AddNews(news)
	if err != nil {
		log.Fatalf("Error when ADD article to database: %v\n", err)
		return err
	}
	return nil
}
