package orm

// midOrm "github.com/FourWD/middleware/orm"

type Promotion struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36)"` //

	Name string `json:"name" query:"name" gorm:"type:varchar(256)"`
}
