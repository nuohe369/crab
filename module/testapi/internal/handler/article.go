package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common/model"
	"github.com/nuohe369/crab/common/request"
	"github.com/nuohe369/crab/common/response"
	"github.com/nuohe369/crab/pkg/util"
)

// SetupArticle registers article routes
// SetupArticle 注册文章路由
func SetupArticle(router fiber.Router) {
	g := router.Group("/article")
	g.Post("/", CreateArticle)
	g.Get("/:id", GetArticle)
	g.Put("/", UpdateArticle)
	g.Delete("/:id", DeleteArticle)
	g.Get("/", ListArticle)
}

// CreateArticle creates an article
// CreateArticle 创建文章
// POST /testapi/article
func CreateArticle(c *fiber.Ctx) error {
	var req request.CreateArticleReq
	if err := c.BodyParser(&req); err != nil {
		return response.FailCode(c, response.CodeParamError)
	}

	userID := util.MustStringToInt64(req.UserID)
	categoryID := util.MustStringToInt64(req.CategoryID)

	if userID == 0 || categoryID == 0 || req.Title == "" {
		return response.FailMsg(c, response.CodeParamMissing, "user_id, category_id, title required")
	}

	article := &model.Article{
		UserID:     userID,
		CategoryID: categoryID,
		Title:      req.Title,
		Content:    req.Content,
		Status:     req.Status,
	}

	_, err := model.GetDB(article).Insert(article)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	return response.OK(c, fiber.Map{
		"id": util.Int64ToString(article.ID),
	})
}

// GetArticle gets an article
// GetArticle 获取文章
// GET /testapi/article/:id
func GetArticle(c *fiber.Ctx) error {
	id := util.MustStringToInt64(c.Params("id"))
	if id == 0 {
		return response.FailCode(c, response.CodeParamError)
	}

	article := &model.Article{}
	has, err := model.GetDB(article).ID(id).Get(article)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}
	if !has {
		return response.FailCode(c, response.CodeNotFound)
	}

	return response.OK(c, toArticleResp(article))
}

// UpdateArticle updates an article
// UpdateArticle 更新文章
// PUT /testapi/article
func UpdateArticle(c *fiber.Ctx) error {
	var req request.UpdateArticleReq
	if err := c.BodyParser(&req); err != nil {
		return response.FailCode(c, response.CodeParamError)
	}

	id := util.MustStringToInt64(req.ID)
	if id == 0 {
		return response.FailMsg(c, response.CodeParamMissing, "id required")
	}

	article := &model.Article{}
	cols := []string{}

	if req.CategoryID != "" {
		categoryID := util.MustStringToInt64(req.CategoryID)
		if categoryID > 0 {
			article.CategoryID = categoryID
			cols = append(cols, "category_id")
		}
	}
	if req.Title != "" {
		article.Title = req.Title
		cols = append(cols, "title")
	}
	if req.Content != "" {
		article.Content = req.Content
		cols = append(cols, "content")
	}
	if req.Status != nil {
		article.Status = *req.Status
		cols = append(cols, "status")
	}

	if len(cols) == 0 {
		return response.FailMsg(c, response.CodeParamMissing, "nothing to update")
	}

	_, err := model.GetDB(article).ID(id).Cols(cols...).Update(article)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	return response.OK(c, nil)
}

// DeleteArticle deletes an article
// DeleteArticle 删除文章
// DELETE /testapi/article/:id
func DeleteArticle(c *fiber.Ctx) error {
	id := util.MustStringToInt64(c.Params("id"))
	if id == 0 {
		return response.FailCode(c, response.CodeParamError)
	}

	article := &model.Article{}
	_, err := model.GetDB(article).ID(id).Delete(article)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	return response.OK(c, nil)
}

// ListArticle lists articles
// ListArticle 文章列表
// GET /testapi/article?page=1&size=10&user_id=xxx&category_id=xxx&status=1
func ListArticle(c *fiber.Ctx) error {
	var req request.ListArticleReq
	if err := c.QueryParser(&req); err != nil {
		return response.FailCode(c, response.CodeParamError)
	}

	article := &model.Article{}
	session := model.GetDB(article).NewSession()
	defer session.Close()

	// Build query conditions | 构建查询条件
	if req.UserID != "" {
		userID := util.MustStringToInt64(req.UserID)
		if userID > 0 {
			session.Where("user_id = ?", userID)
		}
	}
	if req.CategoryID != "" {
		categoryID := util.MustStringToInt64(req.CategoryID)
		if categoryID > 0 {
			session.Where("category_id = ?", categoryID)
		}
	}
	if req.Status != nil {
		session.Where("status = ?", *req.Status)
	}

	var list []model.Article
	total, err := session.Count(article)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	err = session.
		Limit(req.GetSize(), req.GetOffset()).
		Desc("created_at").
		Find(&list)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	return response.OKList(c, toArticleRespList(list), total, req.GetPage(), req.GetSize())
}

// ================ Response ================

type articleResp struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	CategoryID string `json:"category_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	ViewCount  int64  `json:"view_count"`
	Status     int    `json:"status"`
	CreatedAt  string `json:"created_at"`
}

func toArticleResp(a *model.Article) articleResp {
	return articleResp{
		ID:         util.Int64ToString(a.ID),
		UserID:     util.Int64ToString(a.UserID),
		CategoryID: util.Int64ToString(a.CategoryID),
		Title:      a.Title,
		Content:    a.Content,
		ViewCount:  a.ViewCount,
		Status:     a.Status,
		CreatedAt:  a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toArticleRespList(list []model.Article) []articleResp {
	result := make([]articleResp, len(list))
	for i, a := range list {
		result[i] = toArticleResp(&a)
	}
	return result
}
