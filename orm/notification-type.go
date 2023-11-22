package orm

import (
	"github.com/FourWD/middleware/model"
)

type NotificationType struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	Name      string `json:"name" query:"name" gorm:"not null;type:varchar(50)"`
	ImagePath string `json:"image_path" query:"image_path" gorm:"type:varchar(200)"`
	model.GormRowOrder
}
