package vo

// CategoryVO represents the category view object
// CategoryVO 分类视图对象
type CategoryVO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Sort      int    `json:"sort"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}
