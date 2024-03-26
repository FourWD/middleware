package orm

import "github.com/FourWD/middleware/model"

// type Redbook struct {
// 	AuctionID     string    `json:"auction_id" query:"auction_id" gorm:"type:varchar(36);"`
// 	ChassisNo     string    `json:"chassis_no" query:"chassis_no" gorm:"type:varchar(20)"`
// 	CRP           int       `json:"crp" query:"crp" gorm:"type:int"`
// 	CRPPreVat     int       `json:"crp_pre_vat" query:"crp_pre_vat" gorm:"type:int"`
// 	License       string    `json:"license" query:"license" gorm:"type:varchar(10)"`
// 	RedbookCode   string    `json:"redbook_code" query:"license" gorm:"type:varchar(20)"`
// 	RedbookDate   time.Time `json:"redbook_date" query:"redbook_date"`
// 	RedbookPreVat int       `json:"redbook_pre_vat" query:"redbook_pre_vat" gorm:"type:int"`
// 	Remark        string    `json:"remark" query:"remark" gorm:"type:varchar(20)"`
// 	CreatedAt     time.Time `json:"created_at" query:"created_at" gorm:"<-:create"`
// }

type Redbook struct {
	AuctionID string `json:"auction_id" query:"auction_id" gorm:"type:varchar(36);"`
	model.GormModel

	ChassisNo string `json:"chassis_no" query:"chassis_no" gorm:"type:varchar(30)"`
	CRP       int    `json:"crp" query:"crp" gorm:"type:int"`
	License   string `json:"license" query:"license" gorm:"type:varchar(30)"`
	Redbook   int    `json:"redbook" query:"redbook" gorm:"type:int"`
}
