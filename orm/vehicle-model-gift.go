package orm

import "github.com/FourWD/middleware/model"

type VehicleModelGift struct {
	ID             string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	VehicleModelID string `json:"vehicle_model_id" query:"vehicle_model_id" gorm:"type:varchar(36);"`
	model.GormModel
	GiftID string `json:"gift_id" query:"gift_id" gorm:"type:varchar(36);"`
	Qty    int    `json:"qty" query:"qty" gorm:"type:int;"`
	model.GormRowOrder
}
