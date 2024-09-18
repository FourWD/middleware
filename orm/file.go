package orm

import (
	"github.com/FourWD/middleware/model"
)

type File struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	model.GormModel

	FileName   string `json:"file_name" query:"file_name" gorm:"type:varchar(500)"`
	Extension  string `json:"extension" query:"extension" gorm:"type:varchar(5)"`
	Path       string `json:"path" query:"path" gorm:"type:varchar(500)"`
	FullPath   string `json:"full_path"  query:"full_path" gorm:"type:varchar(500)"`
	Cdn        string `json:"cdn" query:"cdn" gorm:"type:varchar(100)"`
	BucketName string `json:"bucket_name" query:"bucket_name" gorm:"type:varchar(500)"`
}
