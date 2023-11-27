package orm

import "github.com/FourWD/middleware/model"

type Employee struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	model.GormModel
	RoleID string `json:"role_id" query:"role_id" gorm:"type:varchar(36);"`

	Code string `json:"code" query:"code" gorm:"type:varchar(10)"`

	Username     string `json:"username" query:"username" gorm:"type:varchar(20);"`
	Firstname    string `json:"firstname" query:"firstname" gorm:"type:varchar(100);"`
	Lastname     string `json:"lastname" query:"lastname" gorm:"type:varchar(20);"`
	Password     string `json:"password" query:"password" gorm:"type:varchar(20);"`
	FileAvatarID string `json:"file_avartar_id" query:"file_avartar_id" gorm:"type:varchar(36)"`
	Mobile       string `json:"mobile" query:"mobile" gorm:"type:varchar(20); unique"`
	Email        string `json:"email" query:"email" gorm:"type:varchar(20);"`
	RunningNo    int    `json:"running_no" query:"running_no" gorm:"type:int;"`
}