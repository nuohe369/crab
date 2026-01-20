package vo

// UserVO represents the user view object
// UserVO 用户视图对象
type UserVO struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}
