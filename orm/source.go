package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type Source struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(10);primary_key"`
	model.GormModel

	Code string `json:"code" query:"code" gorm:"type:varchar(4)"`

	Name                 string    `json:"name" query:"name" gorm:"type:varchar(500)"`
	ShortName            string    `json:"short_name" query:"short_name" gorm:"type:varchar(2);uniqueIndex:idx_source"`
	Phone                string    `json:"phone" query:"phone" gorm:"type:varchar(10)"`
	Email                string    `json:"email" query:"email" gorm:"type:varchar(50)"`
	Address              string    `json:"address" query:"address" gorm:"type:varchar(255)"`
	ProvinceID           string    `json:"province_id" query:"province_id" gorm:"type:varchar(2)"`
	DistrictID           string    `json:"district_id" query:"district_id" gorm:"type:varchar(4)"`
	SubdistrictID        string    `json:"subdistrict_id" query:"subdistrict_id" gorm:"type:varchar(36)"`
	PostCode             string    `json:"postcode" query:"postcode" gorm:"type:varchar(5)"`
	RegistrationNumber   string    `json:"registration_number" query:"registration_number" gorm:"type:varchar(36)"`
	CompanyName          string    `json:"company_name" query:"company_name" gorm:"type:varchar(255)"`
	ContractSigningDate  time.Time `json:"contract_signing_date" query:"contract_signing_date"`
	ContractClostingDate time.Time `json:"contract_closting_date" query:"contract_closting_date"`
}
