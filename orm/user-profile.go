package orm

import "github.com/FourWD/middleware/model"

type UserProfile struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	UserID                 string `json:"user_id" query:"user_id" gorm:"type:varchar(36); uniqueIndex:idx_user_id"`
	Address                string `json:"address" query:"address" gorm:"type:text"`
	Building               string `json:"building" query:"building" gorm:"type:varchar(100)"`
	Room                   string `json:"room" query:"room" gorm:"type:varchar(20)"`
	Street                 string `json:"street" query:"street" gorm:"type:varchar(200)"`
	DistrictID             string `json:"district_id" query:"district_id" gorm:"type:varchar(4)"`
	SubDistrictID          string `json:"sub_district_id" query:"sub_district_id" gorm:"type:varchar(6)"`
	ProvinceID             string `json:"province_id" query:"province_id" gorm:"type:varchar(2)"`
	Postcode               string `json:"postcode" query:"postcode" gorm:"type:varchar(5)"`
	TaxAddress             string `json:"tax_address" query:"tax_address" gorm:"type:text"`
	TaxBuilding            string `json:"tax_building" query:"tax_building" gorm:"type:varchar(100)"`
	TaxRoom                string `json:"tax_room" query:"tax_room" gorm:"type:varchar(20)"`
	TaxStreet              string `json:"tax_street" query:"tax_street" gorm:"type:varchar(200)"`
	TaxDistrictID          string `json:"tax_district_id" query:"tax_district_id" gorm:"type:varchar(4)"`
	TaxSubDistrictID       string `json:"tax_sub_district_id" query:"tax_sub_district_id" gorm:"type:varchar(6)"`
	TaxProvinceID          string `json:"tax_province_id" query:"tax_province_id" gorm:"type:varchar(2)"`
	TaxPostcode            string `json:"tax_postcode" query:"tax_postcode" gorm:"type:varchar(5)"`
	OccupationID           string `json:"occupation_id" query:"occupation_id" gorm:"type:varchar(36)"`
	CompanyVatTypeID       string `json:"company_vat_type_id" query:"company_vat_type_id" gorm:"type:varchar(2)"`
	CompanyRegisterTypeID  string `json:"company_register_type_id" query:"company_register_type_id" gorm:"type:varchar(2)"`
	BusinessTypeID         string `json:"business_type_id" query:"business_type_id" gorm:"type:varchar(2)"`
	CompanyName            string `json:"company_name" query:"company_name" gorm:"type:varchar(50)"`
	CompanyPhone           string `json:"company_phone" query:"company_phone" gorm:"type:varchar(50)"`
	Tax                    string `json:"tax" query:"tax" gorm:"type:varchar(50)"`
	HQShowroomName         string `json:"hq_show_room_name" query:"hq_show_room_name" gorm:"type:varchar(500)"`
	BankID                 string `json:"bank_id" query:"bank_id" gorm:"type:varchar(2)"`
	BankAccountName        string `json:"refund_bank_account_name" query:"refund_bank_account_name" gorm:"type:varchar(15)"`
	BankAccountNo          string `json:"refund_bank_account_no" query:"refund_bank_account_no" gorm:"type:varchar(15)"`
	FileIdcardID           string `json:"file_idcard_id" query:"file_idcard_id" gorm:"type:varchar(36)"`
	FileBookbankID         string `json:"file_bookbank_id" query:"file_bookbank_id" gorm:"type:varchar(36)"`
	FileCompanyRegisterID  string `json:"file_company_register_id" query:"file_company_register_id" gorm:"type:varchar(36)"`
	FilePP20ID             string `json:"file_pp20_id" query:"file_pp20_id" gorm:"type:varchar(36)"` // ภ.พ.20
	FileHouseParticularsID string `json:"file_house_particulars_id" query:"file_house_particulars_id" gorm:"type:varchar(36)"`
	FilePayslipID          string `json:"file_payslip_id" query:"file_payslip_id" gorm:"type:varchar(36)"`
}
