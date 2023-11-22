package orm

import "github.com/FourWD/middleware/model"

type VehicleModelGift struct {
	VehicleModelGiftID string `json:"vehicle_model_gift_id" query:"vehicle_model_gift_id" gorm:"type:varchar(36); uniqueIndex:idx_vehicle_model_gift "`
	model.GormModel

	GiftID             string `json:"gift_id" query:"gift_id" gorm:"type:varchar(36); uniqueIndex:idx_vehicle_model_gift"`
	Qty                int    `json:"qty" query:"qty" gorm:"type:int; uniqueIndex:idx_vehicle_model_gift"`
	model.GormRowOrder        //
}
