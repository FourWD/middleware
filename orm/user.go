package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type User struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	Code         string `json:"code" query:"code" gorm:"type:varchar(500)"`
	UserTypeID   string `json:"user_type_id" query:"user_type_id" gorm:"type:varchar(2)"`
	Username            string    `json:"username" query:"username" gorm:"type:varchar(20)"`
	Password            string    `json:"password" query:"password" gorm:"type:varchar(20)"`
	PrefixID            string    `json:"prefix_id" query:"prefix_id" gorm:"type:varchar(2)"`
	Firstname           string    `json:"firstname" query:"firstname" gorm:"type:varchar(100)"`
	Lastname            string    `json:"lastname" query:"lastname" gorm:"type:varchar(100)"`
	FileAvatarID        string    `json:"file_avartar_id" query:"file_avartar_id" gorm:"type:varchar(36)"`
	Mobile              string    `json:"mobile" query:"mobile" gorm:"type:varchar(20); unique"`
	Email               string    `json:"email" query:"email" gorm:"type:varchar(20)"`
	RunningNo           int       `json:"running_no" query:"running_no" gorm:"type:tinyint"`
	CountWrongLogin     int       `json:"count_wrong_login" query:"count_wrong_login" gorm:"type:int(1)"`
	LastLoginDate       time.Time `json:"last_login_date" query:"last_login_date"`
	UserStatus          string    `json:"user_status" query:"user_status" gorm:"type:varchar(15)"`                   // ban approve
	UserRegisterStatus  string    `json:"user_register_status" query:"user_register_status" gorm:"type:varchar(15)"` //สถานะหน้าสมัคร ถึงขันไหนละ 1 otp เสร็จ 2 กรอกข้อมูลเสร็จ 3 แอดมินแอพ
}
