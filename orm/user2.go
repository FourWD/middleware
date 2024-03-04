package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type User2 struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);uniqueIndex:idx_id"`
	model.GormModel

	Code                     string    `json:"code" query:"code" gorm:"type:varchar(20);unique"`
	UserTypeID               string    `json:"user_type_id" query:"user_type_id" gorm:"type:varchar(2)"`
	Username                 string    `json:"username" query:"username" gorm:"type:varchar(20)"`
	Password                 string    `json:"password" query:"password" gorm:"type:varchar(20)"`
	PrefixID                 string    `json:"prefix_id" query:"prefix_id" gorm:"type:varchar(2)"`
	Firstname                []byte    `json:"firstname" query:"firstname" gorm:"type:blob"`
	Lastname                 []byte    `json:"lastname" query:"lastname" gorm:"type:blob"`
	FileAvatarID             string    `json:"file_avartar_id" query:"file_avartar_id" gorm:"type:varchar(36)"`
	Mobile                   []byte    `json:"mobile" query:"mobile" gorm:"type:blob"`
	Email                    []byte    `json:"email" query:"email" gorm:"type:blob"`
	Facebook                 []byte    `json:"facebook" query:"facebook" gorm:"type:blob"`
	Line                     []byte    `json:"line" query:"line" gorm:"type:blob"`
	Tiktok                   []byte    `json:"tiktok" query:"tiktok" gorm:"type:blob"`
	RunningNo                int       `json:"running_no" query:"running_no" gorm:"type:int"`
	CountWrongLogin          int       `json:"count_wrong_login" query:"count_wrong_login" gorm:"type:int(1)"`
	LastLoginDate            time.Time `json:"last_login_date" query:"last_login_date"`
	LastReadNotificationDate time.Time `json:"last_read_notification_date" query:"last_read_notification_date"`
	//AuctionCode              string    `json:"auction_code" query:"auction_code" gorm:"type:varchar(20)"` //รหัสผู้ประมูล
	//VerifyCode               string    `json:"verify_code" query:"verify_code" gorm:"type:varchar(20)"`   //รหัสยื่นเรื่อง

	UserStatusID         string `json:"user_status_id" query:"user_status_id" gorm:"type:varchar(2)"`                   // ban approve
	UserRegisterStatusID string `json:"user_register_status_id" query:"user_register_status_id" gorm:"type:varchar(2)"` //สถานะหน้าสมัคร ถึงขันไหนละ 1 otp เสร็จ 2 กรอกข้อมูลเสร็จ 3 แอดมินแอพ
	IsRegisterComplete   bool   `json:"is_register_complete" query:"is_register_complete" gorm:"type:bool"`

	// old user profile section
	Address       []byte `json:"address" query:"address" gorm:"type:blob"`
	Building      []byte `json:"building" query:"building" gorm:"type:blob"`
	Room          []byte `json:"room" query:"room" gorm:"type:blob"`
	Street        []byte `json:"street" query:"street" gorm:"type:blob"`
	DistrictID    string `json:"district_id" query:"district_id" gorm:"type:varchar(4)"`         //อำเภอ
	SubDistrictID string `json:"sub_district_id" query:"sub_district_id" gorm:"type:varchar(6)"` //ตำบล
	ProvinceID    string `json:"province_id" query:"province_id" gorm:"type:varchar(2)"`
	Postcode      string `json:"postcode" query:"postcode" gorm:"type:varchar(5)"` //รหัส
	OccupationID  string `json:"occupation_id" query:"occupation_id" gorm:"type:varchar(36)"`
}
