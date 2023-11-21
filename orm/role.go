package orm

type Role struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	GormModel
	Name string `json:"name" query:"name" gorm:"type:varchar(100)"`
	GormRowOrder
}
