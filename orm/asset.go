package orm

import "github.com/FourWD/middleware/model"

type Asset struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	model.GormModel

	Name          string `json:"name" query:"name" gorm:"type:varchar(150)"`
	SerialNo      string `json:"serial_no" query:"serial_no" gorm:"type:varchar(150)"`
	IMEI          string `json:"imei" query:"imei" gorm:"type:varchar(150)"`
	AssetTypeID   string `json:"asset_model_id" query:"asset_model_id" gorm:"type:varchar(36)"`
	AssetStatusID string `json:"asset_status_id" query:"asset_status_id" gorm:"type:varchar(36)"`

	model.GormRowOrder
}
