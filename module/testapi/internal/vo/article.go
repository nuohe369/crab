package vo

// ArticleVO represents the article view object
// ArticleVO 文章视图对象
type ArticleVO struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	CategoryID string `json:"category_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	ViewCount  int64  `json:"view_count"`
	Status     int    `json:"status"`
	CreatedAt  string `json:"created_at"`
}
