package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type Po struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);uniqueIndex:idx_id"`

	model.GormModel

	PoNo             string `json:"po_no" query:"po_no" gorm:"type:varchar(20)"` //หมายเลขรายการ
	PoStatusID       string `json:"po_status_id" query:"po_status_id" gorm:"type:varchar(2)"`
	EmployeeRmID     string `json:"employee_rm_id" query:"employee_rm_id" gorm:"type:varchar(36)"`
	PoEvidenceTypeID string `json:"po_evidence_type_id" query:"po_evidence_type_id" gorm:"type:varchar(2)"`

	UserID     string `json:"user_id" query:"user_id" gorm:"type:varchar(36)"`
	UserTypeID string `json:"user_type_id" query:"user_type_id" gorm:"type:varchar(2)"`
	PrefixID   string `json:"prefix_id" query:"prefix_id" gorm:"type:varchar(2)"`
	FirstName  string `json:"first_name" query:"first_name" gorm:"type:varchar(256)"`
	LastName   string `json:"last_name" query:"last_name" gorm:"type:varchar(256)"`
	Mobile     string `json:"mobile" query:"mobile" gorm:"type:varchar(20)"`
	Email      string `json:"email" query:"email" gorm:"type:varchar(50)"`
	Salary     int    `json:"salary" query:"salary" gorm:"type:int"`
	IsUseCar   bool   `json:"is_use_car" query:"is_use_car" gorm:"type:bool"`

	AuctionID                string `json:"auction_id" query:"auction_id" gorm:"type:varchar(36)"`
	VehicleID                string `json:"vehicle_id" query:"vehicle_id" gorm:"type:varchar(36)"`
	VehicleModelID           string `json:"vehicle_model_id" query:"vehicle_model_id" gorm:"type:varchar(36)"`
	VehicleSubModelID        string `json:"vehicle_sub_model_id" query:"vehicle_sub_model_id" gorm:"type:varchar(36)"`
	VehicleColorID           string `json:"vehicle_color_id" query:"vehicle_color_id" gorm:"type:varchar(36)"`
	IsPaid                   bool   `json:"is_paid" query:"is_paid" gorm:"type:bool"`
	LicenseRegisterCondition bool   `json:"license_register_condition" query:"license_register_condition" gorm:"type:bool"`

	// IsCancel          bool   `json:"is_cancel" query:"is_cancel" gorm:"type:varchar(50)"`

	Address       string `json:"address" query:"address" gorm:"type:text"`
	Building      string `json:"building" query:"building" gorm:"type:varchar(100)"`
	Room          string `json:"room" query:"room" gorm:"type:varchar(20)"`
	Street        string `json:"street" query:"street" gorm:"type:varchar(200)"`
	DistrictID    string `json:"district_id" query:"district_id" gorm:"type:varchar(4)"`         //อำเภอ
	SubDistrictID string `json:"sub_district_id" query:"sub_district_id" gorm:"type:varchar(6)"` //ตำบล
	ProvinceID    string `json:"province_id" query:"province_id" gorm:"type:varchar(2)"`
	Postcode      string `json:"postcode" query:"postcode" gorm:"type:varchar(5)"`

	DiscountPrice  float64 `json:"discount_price" query:"discount_price" gorm:"type:decimal(14,2)"`
	DiscountRemark float64 `json:"discount_remark" query:"discount_remark" gorm:"type:decimal(14,2)"`

	PricePreVat          float64 `json:"price_pre_vat" query:"price_pre_vat" gorm:"type:decimal(14,2)"`
	Vat                  float64 `json:"vat" query:"vat" gorm:"type:decimal(14,2)"`
	Price                float64 `json:"price" query:"price" gorm:"type:decimal(14,2)"`
	TotalAccessoriePrice float64 `json:"total_accessorie_price" query:"total_accessorie_price" gorm:"type:decimal(14,2)"`
	TotalPrice           float64 `json:"total_price" query:"total_price" gorm:"type:decimal(14,2)"`

	PaymentTypeID string  `json:"payment_type_id" query:"payment_type_id" gorm:"type:varchar(2)"`
	Paid          float64 `json:"paid" query:"paid" gorm:"type:decimal(14,2)"`
	UnPaid        float64 `json:"un_paid" query:"un_paid" gorm:"type:decimal(14,2)"`

	ExpectDeliveryDate time.Time `json:"expect_delivery_date" query:"expect_delivery_date"`
	ActualDeliveryDate time.Time `json:"actual_delivery_date" query:"actual_delivery_date"`
	Remark             string    `json:"remark" query:"remark" gorm:"type:text"`

	RunningNo int `json:"running_no" query:"running_no" gorm:"primary_key;auto_increment;not_null"`

	IsGenPo   bool      `json:"is_gen_po" query:"is_gen_po" firestore:"is_gen_po" bson:"is_gen_po" gorm:"type:bool"`
	GenPoDate time.Time `json:"gen_po_date" query:"gen_po_date" firestore:"gen_po_date" bson:"gen_po_date"`
}
