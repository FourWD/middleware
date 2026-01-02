package common

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/FourWD/middleware/model"
	"github.com/FourWD/middleware/orm"
	"github.com/spf13/viper"
)

func Upload(payload model.UploadPayload) (model.UploadResult, error) {
	result, errUpload := uploadFileToServer(payload, viper.GetString("app_id"), viper.GetString("token.upload"))
	if errUpload != nil {
		return result, errUpload
	}

	logFile := new(orm.File)
	logFile.ID = result.ID
	logFile.BucketName = payload.BucketName
	logFile.Cdn = result.Cdn
	logFile.FileName = result.FileName
	logFile.Extension = result.Extension
	logFile.Path = result.Path
	logFile.FullPath = result.FullPath
	if err := Database.Save(&logFile).Error; err != nil {
		LogError("UPLOAD_SAVE_ERROR", map[string]interface{}{"error": err.Error(), "table": "file"}, "")
		return result, err
	}
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
		LogError("UPLOAD_MARSHAL_ERROR", map[string]interface{}{"error": err.Error()}, "")
		return *result, err
	}

	uploadUrl := viper.GetString("upload_service_url")
	if uploadUrl == "" {
		uploadUrl = "https://fourwd.as.r.appspot.com/api/v1/upload/"
	}

	req, err := http.NewRequest("POST", uploadUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		LogError("UPLOAD_REQUEST_ERROR", map[string]interface{}{"error": err.Error()}, "")
		return *result, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	response, err := httpClient.Do(req)
	if err != nil {
		LogError("UPLOAD_EXECUTE_ERROR", map[string]interface{}{"error": err.Error()}, "")
		return *result, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		LogError("UPLOAD_READ_ERROR", map[string]interface{}{"error": err.Error()}, "")
		return *result, err
	}

	var resp ApiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		LogError("UPLOAD_UNMARSHAL_ERROR", map[string]interface{}{"error": err.Error()}, "")
		return *result, err
	}
	result = &resp.Data

	return *result, nil
}

func getBucketName(appID string) string {
	return "fourwd-auction"
}
