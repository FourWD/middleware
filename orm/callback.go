package orm

import "github.com/FourWD/middleware/model"

type Callback struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	ArticleTypeID string `json:"article_type_id" query:"article_type_id" gorm:"type:varchar(36)"`
	Description   string `json:"description" query:"description" gorm:"type:varchar(100)"`
}
