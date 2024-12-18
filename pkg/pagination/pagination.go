package pagination

import "Skillfactory/36-GoNews/pkg/storage/models"

// Число новостей на одной страннице
const (
	NEWS_PER_PAGE = 10
)

// Функция подсчета количества страниц
func PageCounter(totalResults int) int {
	var totalPages int = 1
	totalPages = totalResults / NEWS_PER_PAGE
	if totalPages*NEWS_PER_PAGE < totalResults {
		totalPages++
	}
	return totalPages
}

// Конструктор объекта пагинации
func New(totalResults, currentPage int) *models.Pagination {
	return &models.Pagination{
		TotalResulst: totalResults,
		TotalPages:   PageCounter(totalResults),
		CurrentPage:  currentPage,
		NewsPerPage:  NEWS_PER_PAGE,
	}
}
