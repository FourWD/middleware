package common

import (
	"github.com/FourWD/middleware/common"
	"github.com/FourWD/middleware/model"
	middlewareOrm "github.com/FourWD/middleware/orm"
	midUpload "github.com/FourWD/middleware/upload"
	"github.com/spf13/viper"
)

func Upload(payload model.UploadPayload, db gorm.DB) (model.UploadResult, error) {

	result, errUpload := midUpload.Upload(payload, viper.GetString("app_id"), viper.GetString("token.upload"))
	if errUpload != nil {
		return result, errUpload
	}

	// SAVE TO LOG

	logFile := new(middlewareOrm.File)
	logFile.ID = result.ID
	logFile.BucketName = payload.BucketName
	logFile.Cdn = result.Cdn
	logFile.FileName = result.FileName
	logFile.Extension = result.Extension
	logFile.Path = result.Path
	logFile.FullPath = result.FullPath
	err := db.Save(&logFile)
	if err.Error != nil {
		common.PrintError("error save file", "tb file")
	}
	return result, nil
}
