package orm

type WhitelistMobile struct {
	ID     string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	Mobile string `json:"mobile" query:"mobile" gorm:"type:varchar(10);unique"`
}
