package orm

import "github.com/FourWD/middleware/model"

type FavoriteArticle struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	UserID    string `json:"user_id" query:"user_id" gorm:"type:varchar(36)"`
	ArticleID string `json:"article_id" query:"article_id" gorm:"type:varchar(36)"`
	IsDelete  bool   `json:"is_delete" query:"is_delete" gorm:"type:bool"`
}
