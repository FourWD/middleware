package orm

import "github.com/FourWD/middleware/model"

type Config struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	ConfigName  string `json:"config_name" query:"config_name" gorm:"type:varchar(50)"`
	ConfigValue string `json:"config_value" query:"config_value" gorm:"type:varchar(200)"`
}
