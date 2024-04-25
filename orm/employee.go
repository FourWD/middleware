package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type Employee struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	model.GormModel
	RoleID string `json:"role_id" query:"role_id" gorm:"type:varchar(36);"`

	Code            string    `json:"code" query:"code" gorm:"type:varchar(10)"`
	PrefixID        string    `json:"prefix_id" query:"prefix_id" gorm:"type:varchar(2);"`
	EmployeeTypeID  string    `json:"employee_type_id" query:"employee_type_id" gorm:"type:varchar(36);"`
	Username        string    `json:"username" query:"username" gorm:"type:varchar(20);"`
	Firstname       string    `json:"firstname" query:"firstname" gorm:"type:varchar(100);"`
	Lastname        string    `json:"lastname" query:"lastname" gorm:"type:varchar(20);"`
	Password        string    `json:"password" query:"password" gorm:"type:varchar(150);"`
	FileAvatarID    string    `json:"file_avartar_id" query:"file_avartar_id" gorm:"type:varchar(36)"`
	Mobile          string    `json:"mobile" query:"mobile" gorm:"type:varchar(20);"`
	Email           string    `json:"email" query:"email" gorm:"type:varchar(20);"`
	RunningNo       int       `json:"running_no" query:"running_no" gorm:"type:int;"`
	IsActive        bool      `json:"is_active" query:"is_active" gorm:"type:bool"`
	CountWrongLogin int       `json:"count_wrong_login" query:"count_wrong_login" gorm:"type:int(1)"`
	LastLoginDate   time.Time `json:"last_login_date" query:"last_login_date"`
	BranchID        string    `json:"branch_id" query:"branch_id" gorm:"type:varchar(2);"`
}
