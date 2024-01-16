package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type User struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	Code                     string    `json:"code" query:"code" gorm:"type:varchar(500)"`
	UserTypeID               string    `json:"user_type_id" query:"user_type_id" gorm:"type:varchar(2)"`
	Username                 string    `json:"username" query:"username" gorm:"type:varchar(20)"`
	Password                 string    `json:"password" query:"password" gorm:"type:varchar(20)"`
	PrefixID                 string    `json:"prefix_id" query:"prefix_id" gorm:"type:varchar(2)"`
	Firstname                string    `json:"firstname" query:"firstname" gorm:"type:varchar(100)"`
	Lastname                 string    `json:"lastname" query:"lastname" gorm:"type:varchar(100)"`
	FileAvatarID             string    `json:"file_avartar_id" query:"file_avartar_id" gorm:"type:varchar(36)"`
	Mobile                   string    `json:"mobile" query:"mobile" gorm:"type:varchar(20); unique"`
	Email                    string    `json:"email" query:"email" gorm:"type:varchar(50)"`
	Facebook                 string    `json:"facebook" query:"facebook" gorm:"type:varchar(50)"`
	Line                     string    `json:"line" query:"line" gorm:"type:varchar(20)"`
	Tiktok                   string    `json:"tiktok" query:"tiktok" gorm:"type:varchar(50)"`
	RunningNo                int       `json:"running_no" query:"running_no" gorm:"type:tinyint"`
	CountWrongLogin          int       `json:"count_wrong_login" query:"count_wrong_login" gorm:"type:int(1)"`
	LastLoginDate            time.Time `json:"last_login_date" query:"last_login_date"`
	LastReadNotificationDate time.Time `json:"last_read_notification_date" query:"last_read_notification_date"`
	//AuctionCode              string    `json:"auction_code" query:"auction_code" gorm:"type:varchar(20)"` //รหัสผู้ประมูล
	//VerifyCode               string    `json:"verify_code" query:"verify_code" gorm:"type:varchar(20)"`   //รหัสยื่นเรื่อง

	UserStatusID         string `json:"user_status_id" query:"user_status_id" gorm:"type:varchar(2)"`                   // ban approve
	UserRegisterStatusID string `json:"user_register_status_id" query:"user_register_status_id" gorm:"type:varchar(2)"` //สถานะหน้าสมัคร ถึงขันไหนละ 1 otp เสร็จ 2 กรอกข้อมูลเสร็จ 3 แอดมินแอพ
	IsRegisterComplete   bool   `json:"is_register_complete" query:"is_register_complete" gorm:"type:bool"`
}
