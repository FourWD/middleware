package orm

import (
	"github.com/FourWD/middleware/model"
)

type JwtBlacklist struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	Token string `json:"token" query:"token" gorm:"type:text;index"`
}
