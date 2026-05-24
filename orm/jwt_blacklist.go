package orm

import (
	"github.com/FourWD/middleware/model"
)

type JwtBlacklist struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	Md5   string `json:"md5" query:"md5" gorm:"type:varchar(32);index"`
	Token string `json:"token" query:"token" gorm:"type:text"`
}
