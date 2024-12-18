package rss

import (
	"log"
	"strings"
	"time"

	models "Skillfactory/36-GoNews/pkg/storage/models"

	strip "github.com/grokify/html-strip-tags-go"
	"github.com/mmcdole/gofeed"
)

// Метод - парсер источника RSS. На вход получается строку с URL источника, вовращает слайс объектов или ошибку.
func Parse(source string) ([]models.NewsFullDetailed, error) {
	parser := gofeed.NewParser()
	var news []models.NewsFullDetailed
	var new models.NewsFullDetailed
	feed, err := parser.ParseURL(source)
	if err != nil {
		log.Printf("Parsing error - %v", err)
		return nil, err
	}
	for _, item := range feed.Items {
		new, err = FeedItemToNews(item)
		if err != nil {
			log.Println(err)
		}
		news = append(news, new)
	}
	return news, nil
}

// Метод - конвертер объекта gofeed.Item, предоставляемаого библиотекой gofeed (объект статьи после парсинга XML),
// в объект статьи models.NewsFullDetailed. Возращает ошибку при наличии.
func FeedItemToNews(item *gofeed.Item) (news models.NewsFullDetailed, err error) {
	published := strings.ReplaceAll(item.Published, ",", "")
	t, err := time.Parse("Mon 2 Jan 2006 15:04:05 -0700", published)
	if err != nil {
		t, err = time.Parse("Mon 2 Jan 2006 15:04:05 GMT", published)
	}
	if err == nil {
		news.Published = t.Unix()
	}

	news.Content = item.Description
	news.Content = strip.StripTags(news.Content)

	news = models.NewsFullDetailed{
		Title:     item.Title,
		Content:   news.Content,
		Published: news.Published,
		Link:      item.Link,
	}
	return news, nil
}
