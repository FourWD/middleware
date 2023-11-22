package model

type GormRowOrder struct {
	RowOrder int `json:"row_order" query:"row_order" gorm:"type:int(4);"`
}
