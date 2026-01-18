package request

// PageReq represents pagination request parameters
// PageReq 分页请求参数
type PageReq struct {
	Page int `json:"page" query:"page"` // Page number | 页码
	Size int `json:"size" query:"size"` // Page size | 每页数量
}

// GetPage returns the page number, defaults to 1
// GetPage 获取页码，默认1
func (r *PageReq) GetPage() int {
	if r.Page <= 0 {
		return 1
	}
	return r.Page
}

// GetSize returns the page size, defaults to 10, max 100
// GetSize 获取每页数量，默认10，最大100
func (r *PageReq) GetSize() int {
	if r.Size <= 0 {
		return 10
	}
	if r.Size > 100 {
		return 100
	}
	return r.Size
}

// GetOffset returns the offset for database queries
// GetOffset 获取数据库查询的偏移量
func (r *PageReq) GetOffset() int {
	return (r.GetPage() - 1) * r.GetSize()
}
