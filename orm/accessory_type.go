package orm

import "github.com/FourWD/middleware/model"

type AccessoryType struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(2); primary_key"`
	model.GormModel

	Name                 string `json:"name" query:"name" gorm:"type:varchar(256)"`
	AccessoryTypeGroupID string `json:"accessory_type_group_id" query:"accessory_type_group_id" gorm:"type:varchar(36)"`
	model.GormRowOrder
}
