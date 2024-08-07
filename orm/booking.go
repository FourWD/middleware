package orm

import "github.com/FourWD/middleware/model"

// midOrm "github.com/FourWD/middleware/orm"

type Booking struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);uniqueIndex:idx_id"`
	model.GormModel
	BookingNo       string `json:"booking_no" query:"booking_no" gorm:"type:varchar(20)"` //หมายเลขรายการ
	BookingStatusID string `json:"booking_status_id" query:"booking_status_id" gorm:"type:varchar(2)"`
	BookingCancelID string `json:"booking_cancel_id" query:"booking_cancel_id" gorm:"type:varchar(2)"`

	UserID    string `json:"user_id" query:"user_id" gorm:"type:varchar(36)"`
	PrefixID  string `json:"prefix_id" query:"prefix_id" gorm:"type:varchar(2)"`
	FirstName string `json:"first_name" query:"name" gorm:"type:varchar(256)"`
	LastName  string `json:"last_name" query:"name" gorm:"type:varchar(256)"`
	Mobile    string `json:"mobile" query:"mobile" gorm:"type:varchar(20)"`
	Email     string `json:"email" query:"email" gorm:"type:varchar(50)"`

	VehicleID           string `json:"vehicle_id" query:"vehicle_id" gorm:"type:varchar(36)"`
	VehicleModelID      string `json:"vehicle_model_id" query:"vehicle_model_id" gorm:"type:varchar(36)"`
	VehicleSubModelID   string `json:"vehicle_sub_model_id" query:"vehicle_sub_model_id" gorm:"type:varchar(36)"`
	VehicleColorID      string `json:"vehicle_color_id" query:"vehicle_color_id" gorm:"type:varchar(36)"`
	IsPaid              bool   `json:"is_paid" query:"is_paid" gorm:"type:bool"`
	IsCancel            bool   `json:"is_cancel" query:"is_cancel" gorm:"type:bool"`
	Remark              string `json:"remark" query:"remark" gorm:"type:text"`
	BookingCancelRemark string `json:"booking_cancel_remark" query:"booking_cancel_remark" gorm:"type:text"`
	EmployeeID          string `json:"employee_id" query:"employee_id" gorm:"type:varchar(36)"`

	Address       string `json:"address" query:"address" gorm:"type:text"`
	Building      string `json:"building" query:"building" gorm:"type:varchar(100)"`
	Room          string `json:"room" query:"room" gorm:"type:varchar(20)"`
	Street        string `json:"street" query:"street" gorm:"type:varchar(200)"`
	DistrictID    string `json:"district_id" query:"district_id" gorm:"type:varchar(4)"`         //อำเภอ
	SubDistrictID string `json:"sub_district_id" query:"sub_district_id" gorm:"type:varchar(6)"` //ตำบล
	ProvinceID    string `json:"province_id" query:"province_id" gorm:"type:varchar(2)"`
	Postcode      string `json:"postcode" query:"postcode" gorm:"type:varchar(5)"`

	Facebook string `json:"facebook" query:"facebook" gorm:"type:varchar(50)"`
	Line     string `json:"line" query:"line" gorm:"type:varchar(20)"`
	Tiktok   string `json:"tiktok" query:"tiktok" gorm:"type:varchar(50)"`

	PricePreVat float64 `json:"price_pre_vat" query:"price_pre_vat" gorm:"type:decimal(14,2)"`
	Vat         float64 `json:"vat" query:"vat" gorm:"type:decimal(14,2)"`
	Price       float64 `json:"price" query:"price" gorm:"type:decimal(14,2)"`

	RunningNo int `json:"running_no" query:"running_no" gorm:"primary_key;auto_increment;not_null"`
}
