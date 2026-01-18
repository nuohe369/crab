package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common/model"
	"github.com/nuohe369/crab/common/response"
	"github.com/nuohe369/crab/pkg/transaction"
	"github.com/nuohe369/crab/pkg/util"
)

// DeleteUserWithSaga deletes a user using Saga pattern (cross-database transaction example)
// DeleteUserWithSaga 使用 Saga 模式删除用户（跨库事务示例）
// DELETE /testapi/user/:id/saga?fail_at=step1|step2
//
// Scenario: When deleting a user, all articles by that user must also be deleted
// 场景：删除用户时，需要同时删除该用户的所有文章
// Problem: User is in crab_usercenter, Article is in crab_business (cross-database)
// 问题：User 在 crab_usercenter，Article 在 crab_business（跨库）
// Solution: Use Saga pattern with automatic compensation on failure
// 解决：使用 Saga 模式，失败时自动补偿
//
// Parameters | 参数:
//   - fail_at: simulate failure at step (step1=delete articles fails, step2=delete user fails)
//   - fail_at: 模拟失败的步骤（step1=删除文章失败, step2=删除用户失败）
func DeleteUserWithSaga(c *fiber.Ctx) error {
	id := util.MustStringToInt64(c.Params("id"))
	if id == 0 {
		return response.FailCode(c, response.CodeParamError)
	}

	// Get failure simulation parameter | 获取失败模拟参数
	failAt := c.Query("fail_at", "")

	// Data for compensation | 用于补偿的数据
	var deletedArticles []model.Article
	var deletedUser *model.User

	// Create Saga transaction | 创建 Saga 事务
	saga := transaction.NewSaga().
		// Step 1: Delete all articles by the user (crab_business database)
		// 步骤 1: 删除用户的所有文章（crab_business 库）
		AddStep(transaction.SagaStep{
			Name: "delete_user_articles",
			Execute: func(ctx context.Context) error {
				article := &model.Article{}

				// Query articles to be deleted (for compensation)
				// 查询要删除的文章（用于补偿）
				err := model.GetDB(article).
					Where("user_id = ?", id).
					Find(&deletedArticles)
				if err != nil {
					return err
				}

				// Simulate failure | 模拟失败
				if failAt == "step1" {
					return fiber.NewError(fiber.StatusInternalServerError, "simulated failure at step1")
				}

				// Delete articles | 删除文章
				_, err = model.GetDB(article).
					Where("user_id = ?", id).
					Delete(&model.Article{})
				return err
			},
			Compensate: func(ctx context.Context) error {
				// Compensation: restore deleted articles
				// 补偿：恢复删除的文章
				if len(deletedArticles) > 0 {
					article := &model.Article{}
					for _, a := range deletedArticles {
						_, err := model.GetDB(article).Insert(&a)
						if err != nil {
							return err
						}
					}
				}
				return nil
			},
		}).
		// Step 2: Delete user (crab_usercenter database)
		// 步骤 2: 删除用户（crab_usercenter 库）
		AddStep(transaction.SagaStep{
			Name: "delete_user",
			Execute: func(ctx context.Context) error {
				user := &model.User{}

				// Query user info (for compensation)
				// 查询用户信息（用于补偿）
				has, err := model.GetDB(user).ID(id).Get(user)
				if err != nil {
					return err
				}
				if !has {
					return fiber.NewError(fiber.StatusNotFound, "user not found")
				}
				deletedUser = user

				// Simulate failure | 模拟失败
				if failAt == "step2" {
					return fiber.NewError(fiber.StatusInternalServerError, "simulated failure at step2")
				}

				// Delete user | 删除用户
				_, err = model.GetDB(user).ID(id).Delete(user)
				return err
			},
			Compensate: func(ctx context.Context) error {
				// Compensation: restore deleted user
				// 补偿：恢复删除的用户
				if deletedUser != nil {
					user := &model.User{}
					_, err := model.GetDB(user).Insert(deletedUser)
					return err
				}
				return nil
			},
		}).
		// Success callback | 成功回调
		OnSuccess(func(ctx context.Context) error {
			// Can log, send notifications, etc.
			// 可以记录日志、发送通知等
			return nil
		}).
		// Failure callback | 失败回调
		OnFailure(func(ctx context.Context, err error) error {
			// Log failure | 记录失败日志
			return nil
		})

	// Execute Saga | 执行 Saga
	if err := saga.Execute(c.Context()); err != nil {
		return response.FailMsg(c, response.CodeServerError, err.Error())
	}

	return response.OK(c, fiber.Map{
		"message":          "user and articles deleted successfully",
		"deleted_articles": len(deletedArticles),
	})
}
