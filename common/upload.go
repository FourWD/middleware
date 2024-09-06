package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/FourWD/middleware/model"
	"github.com/FourWD/middleware/orm"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func Upload(payload model.UploadPayload, db gorm.DB) (model.UploadResult, error) {
	result, errUpload := uploadFileToServer(payload, viper.GetString("app_id"), viper.GetString("token.upload"))
	if errUpload != nil {
		return result, errUpload
	}

	println(payload.Filename + " : " + payload.FileBase64 + " : " + payload.Path)
	println(result.ID + " : " + result.FileName + " : " + result.FullPath)

	// SAVE TO LOG
	logFile := new(orm.File)
	logFile.ID = result.ID
	logFile.BucketName = payload.BucketName
	logFile.Cdn = result.Cdn
	logFile.FileName = result.FileName
	logFile.Extension = result.Extension
	logFile.Path = result.Path
	logFile.FullPath = result.FullPath
	err := db.Save(&logFile)
	if err.Error != nil {
		PrintError("error save file", "tb file")
	} //
	return result, nil
}

func uploadFileToServer(p model.UploadPayload, appID string, token string) (model.UploadResult, error) {
	type ApiResponse struct {
		Status     int                `json:"status"`
		StatusCode string             `json:"status_code"`
		Message    string             `json:"message"`
		Data       model.UploadResult `json:"data"`
	}

	p.BucketName = getBucketName(appID)

	result := new(model.UploadResult)
	p.FileBase64 = strings.Replace(p.FileBase64, "data:image/png;base64,", "", -1)
	p.FileBase64 = strings.Replace(p.FileBase64, "data:image/jpeg;base64,", "", -1)
	p.FileBase64 = strings.Replace(p.FileBase64, "data:image/jpg;base64,", "", -1)

	jsonData, err := json.Marshal(p)

	if err != nil {
		fmt.Println("there was an error with the JSON", err.Error())
		return *result, err
	} else {
		client := &http.Client{}
		uploadUrl := "https://pakwan-service.fourwd.me/api/v1/upload/"
		// uploadUrl := "https://fourwd.as.r.appspot.com/api/v1/upload/"
		// uploadUrl := "https://pakwan-service.fourwd.me/api/v1/upload/" //
		Print("pakwan-service", uploadUrl)

		req, err := http.NewRequest("POST", uploadUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println(err)
			return *result, err
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+token)

		response, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return *result, err
		}
		defer response.Body.Close()

		body, _ := io.ReadAll(response.Body)
		var resp ApiResponse
		errJson := json.Unmarshal(body, &resp)
		if errJson != nil {
			return *result, errJson
		}
		result = &resp.Data
	}

	return *result, nil
}

func getBucketName(appID string) string {
	return "fourwd-auction"
}
