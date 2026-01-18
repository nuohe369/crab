package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common/model"
	"github.com/nuohe369/crab/common/request"
	"github.com/nuohe369/crab/common/response"
	"github.com/nuohe369/crab/pkg/util"
)

// SetupCategory registers category routes
// SetupCategory 注册分类路由
func SetupCategory(router fiber.Router) {
	g := router.Group("/category")
	g.Post("/", CreateCategory)
	g.Get("/:id", GetCategory)
	g.Put("/", UpdateCategory)
	g.Delete("/:id", DeleteCategory)
	g.Get("/", ListCategory)
}

// CreateCategory creates a category
// CreateCategory 创建分类
// POST /testapi/category
func CreateCategory(c *fiber.Ctx) error {
	var req request.CreateCategoryReq
	if err := c.BodyParser(&req); err != nil {
		return response.FailCode(c, response.CodeParamError)
	}

	if req.Name == "" {
		return response.FailMsg(c, response.CodeParamMissing, "name required")
	}

	category := &model.Category{
		Name: req.Name,
		Sort: req.Sort,
	}

	_, err := model.GetDB(category).Insert(category)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	return response.OK(c, fiber.Map{
		"id": util.Int64ToString(category.ID),
	})
}

// GetCategory gets a category
// GetCategory 获取分类
// GET /testapi/category/:id
func GetCategory(c *fiber.Ctx) error {
	id := util.MustStringToInt64(c.Params("id"))
	if id == 0 {
		return response.FailCode(c, response.CodeParamError)
	}

	category := &model.Category{}
	has, err := model.GetDB(category).ID(id).Get(category)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}
	if !has {
		return response.FailCode(c, response.CodeNotFound)
	}

	return response.OK(c, toCategoryResp(category))
}

// UpdateCategory updates a category
// UpdateCategory 更新分类
// PUT /testapi/category
func UpdateCategory(c *fiber.Ctx) error {
	var req request.UpdateCategoryReq
	if err := c.BodyParser(&req); err != nil {
		return response.FailCode(c, response.CodeParamError)
	}

	id := util.MustStringToInt64(req.ID)
	if id == 0 {
		return response.FailMsg(c, response.CodeParamMissing, "id required")
	}

	category := &model.Category{}
	cols := []string{}

	if req.Name != "" {
		category.Name = req.Name
		cols = append(cols, "name")
	}
	if req.Sort != nil {
		category.Sort = *req.Sort
		cols = append(cols, "sort")
	}
	if req.Status != nil {
		category.Status = *req.Status
		cols = append(cols, "status")
	}

	if len(cols) == 0 {
		return response.FailMsg(c, response.CodeParamMissing, "nothing to update")
	}

	_, err := model.GetDB(category).ID(id).Cols(cols...).Update(category)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	return response.OK(c, nil)
}

// DeleteCategory deletes a category
// DeleteCategory 删除分类
// DELETE /testapi/category/:id
func DeleteCategory(c *fiber.Ctx) error {
	id := util.MustStringToInt64(c.Params("id"))
	if id == 0 {
		return response.FailCode(c, response.CodeParamError)
	}

	category := &model.Category{}
	_, err := model.GetDB(category).ID(id).Delete(category)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	return response.OK(c, nil)
}

// ListCategory lists categories
// ListCategory 分类列表
// GET /testapi/category?page=1&size=10
func ListCategory(c *fiber.Ctx) error {
	var req request.PageReq
	if err := c.QueryParser(&req); err != nil {
		return response.FailCode(c, response.CodeParamError)
	}

	category := &model.Category{}
	var list []model.Category

	total, err := model.GetDB(category).Count(category)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	err = model.GetDB(category).
		Limit(req.GetSize(), req.GetOffset()).
		Asc("sort").
		Desc("created_at").
		Find(&list)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	return response.OKList(c, toCategoryRespList(list), total, req.GetPage(), req.GetSize())
}

// ================ Response ================

type categoryResp struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Sort      int    `json:"sort"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}

func toCategoryResp(c *model.Category) categoryResp {
	return categoryResp{
		ID:        util.Int64ToString(c.ID),
		Name:      c.Name,
		Sort:      c.Sort,
		Status:    c.Status,
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toCategoryRespList(list []model.Category) []categoryResp {
	result := make([]categoryResp, len(list))
	for i, c := range list {
		result[i] = toCategoryResp(&c)
	}
	return result
}
