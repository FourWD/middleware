package orm

import "github.com/FourWD/middleware/model"

type SurveyGroup struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	Name   string `json:"name" query:"name" gorm:"not null;type:varchar(100)"`
	Remark string `json:"remark" query:"remark" gorm:"type:text"`
}

// ชื่อของ survey เช่น survey "ลูกค้าทั่วไปชอบรถประเภทไหน"
