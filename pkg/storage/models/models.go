package models

type NewsFullDetailed struct {
	ID        int    `db:"id"`
	Title     string `db:"title"`
	Content   string `db:"description"`
	Preview   string `db:"preview"`
	Published int64  `db:"published"`
	Link      string `db:"link"`
}

type NewsShortDetailed struct {
	ID      int    `db:"id"`
	Title   string `db:"title"`
	Preview string `db:"preview"` //поле preview = дополнительный столбец в таблице news
}

// comments - отдельная таблица и вероятно отдельная база (т.к. - отдельный микросервис)
type Comment struct {
	ID        int    `db:"id"`
	NewsId    int    `db:"news_id"`
	Author    string `db:"author"`
	CreatedAt string `db:"created_at"`
	Сensor    bool   `db:"censor"` //true-прошел цензуру/false - нет
}

// Объкт пагинации
type Pagination struct {
	TotalResulst int                `json:"total_results"`
	TotalPages   int                `json:"total_pages"`
	CurrentPage  int                `json:"current_page"`
	NewsPerPage  int                `json:"news_per_page"`
	Results      []NewsFullDetailed `json:"results"`
}
