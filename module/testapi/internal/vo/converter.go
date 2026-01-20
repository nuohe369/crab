package vo

import "github.com/nuohe369/crab/common/model"

// ToUserVO converts model.ExampleUser to UserVO
// ToUserVO 将 model.ExampleUser 转换为 UserVO
func ToUserVO(u *model.ExampleUser) UserVO {
	return UserVO{
		ID:        u.ID.String(),
		Username:  u.Username,
		Nickname:  u.Nickname,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ToUserVOList converts []model.ExampleUser to []UserVO
// ToUserVOList 将 []model.ExampleUser 转换为 []UserVO
func ToUserVOList(list []model.ExampleUser) []UserVO {
	result := make([]UserVO, len(list))
	for i, u := range list {
		result[i] = ToUserVO(&u)
	}
	return result
}

// ToArticleVO converts model.ExampleArticle to ArticleVO
// ToArticleVO 将 model.ExampleArticle 转换为 ArticleVO
func ToArticleVO(a *model.ExampleArticle) ArticleVO {
	return ArticleVO{
		ID:         a.ID.String(),
		UserID:     a.UserID.String(),
		CategoryID: a.CategoryID.String(),
		Title:      a.Title,
		Content:    a.Content,
		ViewCount:  a.ViewCount,
		Status:     a.Status,
		CreatedAt:  a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ToArticleVOList converts []model.ExampleArticle to []ArticleVO
// ToArticleVOList 将 []model.ExampleArticle 转换为 []ArticleVO
func ToArticleVOList(list []model.ExampleArticle) []ArticleVO {
	result := make([]ArticleVO, len(list))
	for i, a := range list {
		result[i] = ToArticleVO(&a)
	}
	return result
}

// ToCategoryVO converts model.ExampleCategory to CategoryVO
// ToCategoryVO 将 model.ExampleCategory 转换为 CategoryVO
func ToCategoryVO(c *model.ExampleCategory) CategoryVO {
	return CategoryVO{
		ID:        c.ID.String(),
		Name:      c.Name,
		Sort:      c.Sort,
		Status:    c.Status,
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ToCategoryVOList converts []model.ExampleCategory to []CategoryVO
// ToCategoryVOList 将 []model.ExampleCategory 转换为 []CategoryVO
func ToCategoryVOList(list []model.ExampleCategory) []CategoryVO {
	result := make([]CategoryVO, len(list))
	for i, c := range list {
		result[i] = ToCategoryVO(&c)
	}
	return result
}
