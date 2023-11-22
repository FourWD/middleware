package orm

type SourceFinance struct {
	SourceID  string `json:"source_id" query:"source_id" gorm:"type:varchar(10); uniqueIndex:idx_source_finances"`
	FinanceID string `json:"finance_id" query:"finance_id" gorm:"type:varchar(10); uniqueIndex:idx_source_finances"`
}
