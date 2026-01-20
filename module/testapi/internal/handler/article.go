package handler

import (
	"github.com/nuohe369/crab/common/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common/model"
	"github.com/nuohe369/crab/common/request"
	"github.com/nuohe369/crab/common/response"
	"github.com/nuohe369/crab/module/testapi/internal/vo"
	"github.com/nuohe369/crab/pkg/snowflake"
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
		return errors.ErrParamInvalid("参数解析失败")
	}

	userID := snowflake.SnowflakeID(util.MustStringToInt64(req.UserID))
	categoryID := snowflake.SnowflakeID(util.MustStringToInt64(req.CategoryID))

	if userID.IsZero() || categoryID.IsZero() || req.Title == "" {
		return errors.New(response.CodeParamMissing, "user_id, category_id, title required")
	}

	article := &model.ExampleArticle{
		UserID:     userID,
		CategoryID: categoryID,
		Title:      req.Title,
		Content:    req.Content,
		Status:     req.Status,
	}

	_, err := model.GetDB(article).Insert(article)
	if err != nil {
		return errors.ErrDBError(err)
	}

	return response.OK(c, fiber.Map{
		"id": article.ID.String(),
	})
}

// GetArticle gets an article
// GetArticle 获取文章
// GET /testapi/article/:id
func GetArticle(c *fiber.Ctx) error {
	id := util.MustStringToInt64(c.Params("id"))
	if id == 0 {
		return errors.ErrParamInvalid("参数解析失败")
	}

	article := &model.ExampleArticle{}
	has, err := model.GetDB(article).ID(id).Get(article)
	if err != nil {
		return errors.ErrDBError(err)
	}
	if !has {
		return errors.ErrNotFound()
	}

	return response.OK(c, vo.ToArticleVO(article))
}

// UpdateArticle updates an article
// UpdateArticle 更新文章
// PUT /testapi/article
func UpdateArticle(c *fiber.Ctx) error {
	var req request.UpdateArticleReq
	if err := c.BodyParser(&req); err != nil {
		return errors.ErrParamInvalid("参数解析失败")
	}

	id := util.MustStringToInt64(req.ID)
	if id == 0 {
		return errors.New(response.CodeParamMissing, "id required")
	}

	article := &model.ExampleArticle{}
	cols := []string{}

	if req.CategoryID != "" {
		categoryID := snowflake.SnowflakeID(util.MustStringToInt64(req.CategoryID))
		if categoryID.Valid() {
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
		return errors.New(response.CodeParamMissing, "nothing to update")
	}

	_, err := model.GetDB(article).ID(id).Cols(cols...).Update(article)
	if err != nil {
		return errors.ErrDBError(err)
	}

	return response.OK(c, nil)
}

// DeleteArticle deletes an article
// DeleteArticle 删除文章
// DELETE /testapi/article/:id
func DeleteArticle(c *fiber.Ctx) error {
	id := util.MustStringToInt64(c.Params("id"))
	if id == 0 {
		return errors.ErrParamInvalid("参数解析失败")
	}

	article := &model.ExampleArticle{}
	_, err := model.GetDB(article).ID(id).Delete(article)
	if err != nil {
		return errors.ErrDBError(err)
	}

	return response.OK(c, nil)
}

// ListArticle lists articles
// ListArticle 文章列表
// GET /testapi/article?page=1&size=10&user_id=xxx&category_id=xxx&status=1
func ListArticle(c *fiber.Ctx) error {
	var req request.ListArticleReq
	if err := c.QueryParser(&req); err != nil {
		return errors.ErrParamInvalid("参数解析失败")
	}

	article := &model.ExampleArticle{}
	session := model.GetDB(article).NewSession()
	defer session.Close()

	// Build query conditions | 构建查询条件
	if req.UserID != "" {
		userID := snowflake.SnowflakeID(util.MustStringToInt64(req.UserID))
		if userID.Valid() {
			session.Where("user_id = ?", userID)
		}
	}
	if req.CategoryID != "" {
		categoryID := snowflake.SnowflakeID(util.MustStringToInt64(req.CategoryID))
		if categoryID.Valid() {
			session.Where("category_id = ?", categoryID)
		}
	}
	if req.Status != nil {
		session.Where("status = ?", *req.Status)
	}

	var list []model.ExampleArticle
	total, err := session.Count(article)
	if err != nil {
		return errors.ErrDBError(err)
	}

	err = session.
		Limit(req.GetSize(), req.GetOffset()).
		Desc("created_at").
		Find(&list)
	if err != nil {
		return errors.ErrDBError(err)
	}

	return response.OKList(c, vo.ToArticleVOList(list), total, req.GetPage(), req.GetSize())
}
