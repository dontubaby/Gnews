package api

import (
	"Skillfactory/36-GoNews/pkg/pagination"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"Skillfactory/36-GoNews/pkg/storage/postgress"

	"github.com/gorilla/mux"
)

// Объект API
type Api struct {
	db *postgress.Storage
	r  *mux.Router
}

// Конуструктор объекта API
func New(db *postgress.Storage) *Api {
	api := Api{db: db, r: mux.NewRouter()}
	api.endpoints()
	return &api
}

// Init-метод для API-роутера
func (api *Api) Router() *mux.Router {
	return api.r
}

// Метод регистратор endpoint-ов и настраивающий саброутинг для файлового сервера (веб-приложения).
func (api *Api) endpoints() {
	//маршрут для возврата детальной информации о новости
	api.r.HandleFunc("/newsdetail/{id}", api.GetDetailedNewsHandler).Methods(http.MethodGet, http.MethodOptions)
	//маршрут для возврата списка новостей
	api.r.HandleFunc("/newslist/", http.HandlerFunc(api.GetNewsListHandler)).Methods(http.MethodGet, http.MethodOptions)
	//маршрут для возврата списка  новостей отфильтрованных по контенту
	api.r.HandleFunc("/newslist/filtered/", api.FilteredByContentHandler).Methods(http.MethodGet, http.MethodOptions)
	//маршрут для возврата списка новостей отфильтрованных по дате публикации
	api.r.HandleFunc("/newslist/filtered/date/", api.FilteredByPublishedHandler).Methods(http.MethodGet, http.MethodOptions)
	//маршрут для подключения к веб-приложению
	api.r.PathPrefix("/").Handler(http.StripPrefix("/news/", http.FileServer(http.Dir("./webapp"))))
}

// хэндлер отдающий детальную информацию о новости
func (api *Api) GetDetailedNewsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		return
	}
	s := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(s)

	reqid := r.URL.Query().Get("request_id")
	fmt.Println("Request id AFTER", r.Context().Value("request_id"))

	news, err := api.db.GetDetailedNews(id)
	if err != nil {
		http.Error(w, "failed get detailed news from DB", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(news)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(reqid))
}

// хэндлер отдающий список новостей с пагинацией
func (api *Api) GetNewsListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}

	nStr := r.URL.Query().Get("n")
	n, _ := strconv.Atoi(nStr)

	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, _ := strconv.Atoi(pageStr)
	pag := pagination.New(n, page)
	results, err := api.db.GetNewsListWithPagination(n, (pag.CurrentPage-1)*pag.NewsPerPage, pag.NewsPerPage)
	if err != nil {
		http.Error(w, "failed get news from DB", http.StatusInternalServerError)
		return
	}
	pag.Results = results
	json.NewEncoder(w).Encode(pag)
	w.WriteHeader(http.StatusOK)
}

// хэндлер отдающий новости отфильтрованные по контенту c пагинацией
func (api *Api) FilteredByContentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}

	filter := r.URL.Query().Get("s")

	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, _ := strconv.Atoi(pageStr)
	//запрос в БД необходимый для подсчета количества новостей отфильтрованных с учетом заданного фильтра
	filteredData, err := api.db.FilterNewsByContent(filter)
	if err != nil {
		http.Error(w, "failed get filtered by content news from DB", http.StatusInternalServerError)
		return
	}

	pag := pagination.New(len(filteredData), page)
	results, err := api.db.FilterNewsByContentWithPagination(filter, (pag.CurrentPage-1)*pag.NewsPerPage, pag.NewsPerPage)
	if err != nil {
		http.Error(w, "failed get filtered by content news with pagination from DB", http.StatusInternalServerError)
		return
	}
	pag.Results = results
	json.NewEncoder(w).Encode(pag)
	w.WriteHeader(http.StatusOK)
}

// хэндлер отдающий новости отфильтрованные по дате публикации
func (api *Api) FilteredByPublishedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}
	filter := r.URL.Query().Get("date")
	filterInt, _ := strconv.Atoi(filter)
	news, err := api.db.FilterNewsByPublished(filterInt)
	if err != nil {
		http.Error(w, "failed get filtered by published news from DB", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(news)
	w.WriteHeader(http.StatusOK)
}
