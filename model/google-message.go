package model

type GoogleMessage struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	GormModel

	Group   string `json:"group" query:"group" gorm:"type:varchar(50)"`
	Message string `json:"message" query:"message" gorm:"type:text"`
}
