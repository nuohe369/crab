package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common/model"
	"github.com/nuohe369/crab/common/request"
	"github.com/nuohe369/crab/common/response"
	"github.com/nuohe369/crab/pkg/util"
)

// SetupUser registers user routes
// SetupUser 注册用户路由
func SetupUser(router fiber.Router) {
	g := router.Group("/user")
	g.Post("/", CreateUser)
	g.Get("/:id", GetUser)
	g.Put("/", UpdateUser)
	g.Delete("/:id", DeleteUser)
	g.Get("/", ListUser)

	// Saga transaction example | Saga 事务示例
	g.Delete("/:id/saga", DeleteUserWithSaga)
}

// CreateUser creates a user
// CreateUser 创建用户
// POST /testapi/user
func CreateUser(c *fiber.Ctx) error {
	var req request.CreateUserReq
	if err := c.BodyParser(&req); err != nil {
		return response.FailCode(c, response.CodeParamError)
	}

	if req.Username == "" || req.Nickname == "" || req.Password == "" {
		return response.FailMsg(c, response.CodeParamMissing, "username, nickname, password required")
	}

	user := &model.User{
		Username: req.Username,
		Nickname: req.Nickname,
	}
	if err := user.SetPassword(req.Password); err != nil {
		return response.FailMsg(c, response.CodeServerError, err.Error())
	}

	_, err := model.GetDB(user).Insert(user)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	return response.OK(c, fiber.Map{
		"id": util.Int64ToString(user.ID),
	})
}

// GetUser gets a user
// GetUser 获取用户
// GET /testapi/user/:id
func GetUser(c *fiber.Ctx) error {
	id := util.MustStringToInt64(c.Params("id"))
	if id == 0 {
		return response.FailCode(c, response.CodeParamError)
	}

	user := &model.User{}
	has, err := model.GetDB(user).ID(id).Get(user)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}
	if !has {
		return response.FailCode(c, response.CodeUserNotFound)
	}

	return response.OK(c, toUserResp(user))
}

// UpdateUser updates a user
// UpdateUser 更新用户
// PUT /testapi/user
func UpdateUser(c *fiber.Ctx) error {
	var req request.UpdateUserReq
	if err := c.BodyParser(&req); err != nil {
		return response.FailCode(c, response.CodeParamError)
	}

	id := util.MustStringToInt64(req.ID)
	if id == 0 {
		return response.FailMsg(c, response.CodeParamMissing, "id required")
	}

	user := &model.User{}
	cols := []string{}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
		cols = append(cols, "nickname")
	}
	if req.Status != nil {
		user.Status = *req.Status
		cols = append(cols, "status")
	}

	if len(cols) == 0 {
		return response.FailMsg(c, response.CodeParamMissing, "nothing to update")
	}

	_, err := model.GetDB(user).ID(id).Cols(cols...).Update(user)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	return response.OK(c, nil)
}

// DeleteUser deletes a user
// DeleteUser 删除用户
// DELETE /testapi/user/:id
func DeleteUser(c *fiber.Ctx) error {
	id := util.MustStringToInt64(c.Params("id"))
	if id == 0 {
		return response.FailCode(c, response.CodeParamError)
	}

	user := &model.User{}
	_, err := model.GetDB(user).ID(id).Delete(user)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	return response.OK(c, nil)
}

// ListUser lists users
// ListUser 用户列表
// GET /testapi/user?page=1&size=10
func ListUser(c *fiber.Ctx) error {
	var req request.PageReq
	if err := c.QueryParser(&req); err != nil {
		return response.FailCode(c, response.CodeParamError)
	}

	user := &model.User{}
	var list []model.User

	total, err := model.GetDB(user).Count(user)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	err = model.GetDB(user).
		Limit(req.GetSize(), req.GetOffset()).
		Desc("created_at").
		Find(&list)
	if err != nil {
		return response.FailMsg(c, response.CodeDBError, err.Error())
	}

	return response.OKList(c, toUserRespList(list), total, req.GetPage(), req.GetSize())
}

// ================ Response ================

type userResp struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}

func toUserResp(u *model.User) userResp {
	return userResp{
		ID:        util.Int64ToString(u.ID),
		Username:  u.Username,
		Nickname:  u.Nickname,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toUserRespList(list []model.User) []userResp {
	result := make([]userResp, len(list))
	for i, u := range list {
		result[i] = toUserResp(&u)
	}
	return result
}
