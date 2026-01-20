package orm

import (
	"github.com/FourWD/middleware/model"
)

type Finance struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(10);primary_key"`
	model.GormModel

	Logo            string `json:"logo" query:"logo" gorm:"type:varchar(256)"`
	LogoHome        string `json:"logo_home" query:"logo_home" gorm:"type:varchar(256)"`
	Code            string `json:"code" query:"code" gorm:"type:varchar(4)"`
	Label           string `json:"label" query:"label" gorm:"type:varchar(10)"`
	Name            string `json:"name" query:"name" gorm:"type:varchar(50)"`
	NameEn          string `json:"name_en" query:"name_en" gorm:"type:varchar(50)"`
	Color           string `json:"color" query:"color" gorm:"type:varchar(7)"`
	LoanDescription string `json:"loan_description" query:"loan_description" gorm:"type:text"`
	Detail          string `json:"detail" query:"detail" gorm:"type:text"`
	GroupCompany    string `json:"group_company" query:"group_company" gorm:"type:varchar(500)"`
	Remark          string `json:"remark" query:"remark" gorm:"type:text"`

	model.GormRowOrder
}
