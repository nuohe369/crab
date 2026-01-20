package handler

import (
	"github.com/nuohe369/crab/common/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common/model"
	"github.com/nuohe369/crab/common/request"
	"github.com/nuohe369/crab/common/response"
	"github.com/nuohe369/crab/module/testapi/internal/vo"
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
		return errors.ErrParamInvalid("参数解析失败")
	}

	if req.Name == "" {
		return errors.New(response.CodeParamMissing, "name required")
	}

	category := &model.ExampleCategory{
		Name: req.Name,
		Sort: req.Sort,
	}

	_, err := model.GetDB(category).Insert(category)
	if err != nil {
		return errors.ErrDBError(err)
	}

	return response.OK(c, fiber.Map{
		"id": category.ID.String(),
	})
}

// GetCategory gets a category
// GetCategory 获取分类
// GET /testapi/category/:id
func GetCategory(c *fiber.Ctx) error {
	id := util.MustStringToInt64(c.Params("id"))
	if id == 0 {
		return errors.ErrParamInvalid("参数解析失败")
	}

	category := &model.ExampleCategory{}
	has, err := model.GetDB(category).ID(id).Get(category)
	if err != nil {
		return errors.ErrDBError(err)
	}
	if !has {
		return errors.ErrNotFound()
	}

	return response.OK(c, vo.ToCategoryVO(category))
}

// UpdateCategory updates a category
// UpdateCategory 更新分类
// PUT /testapi/category
func UpdateCategory(c *fiber.Ctx) error {
	var req request.UpdateCategoryReq
	if err := c.BodyParser(&req); err != nil {
		return errors.ErrParamInvalid("参数解析失败")
	}

	id := util.MustStringToInt64(req.ID)
	if id == 0 {
		return errors.New(response.CodeParamMissing, "id required")
	}

	category := &model.ExampleCategory{}
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
		return errors.New(response.CodeParamMissing, "nothing to update")
	}

	_, err := model.GetDB(category).ID(id).Cols(cols...).Update(category)
	if err != nil {
		return errors.ErrDBError(err)
	}

	return response.OK(c, nil)
}

// DeleteCategory deletes a category
// DeleteCategory 删除分类
// DELETE /testapi/category/:id
func DeleteCategory(c *fiber.Ctx) error {
	id := util.MustStringToInt64(c.Params("id"))
	if id == 0 {
		return errors.ErrParamInvalid("参数解析失败")
	}

	category := &model.ExampleCategory{}
	_, err := model.GetDB(category).ID(id).Delete(category)
	if err != nil {
		return errors.ErrDBError(err)
	}

	return response.OK(c, nil)
}

// ListCategory lists categories
// ListCategory 分类列表
// GET /testapi/category?page=1&size=10
func ListCategory(c *fiber.Ctx) error {
	var req request.PageReq
	if err := c.QueryParser(&req); err != nil {
		return errors.ErrParamInvalid("参数解析失败")
	}

	category := &model.ExampleCategory{}
	var list []model.ExampleCategory

	total, err := model.GetDB(category).Count(category)
	if err != nil {
		return errors.ErrDBError(err)
	}

	err = model.GetDB(category).
		Limit(req.GetSize(), req.GetOffset()).
		Asc("sort").
		Desc("created_at").
		Find(&list)
	if err != nil {
		return errors.ErrDBError(err)
	}

	return response.OKList(c, vo.ToCategoryVOList(list), total, req.GetPage(), req.GetSize())
}
