package common

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/FourWD/middleware/infra"
	"github.com/FourWD/middleware/kit"
	"github.com/FourWD/middleware/model"
	"github.com/FourWD/middleware/orm"
)

func Upload(payload model.UploadPayload) (model.UploadResult, error) {
	result, errUpload := uploadFileToServer(payload, infra.GetEnv("APP_ID", ""), infra.GetEnv("UPLOAD_TOKEN", ""))
	if errUpload != nil {
		return result, errUpload
	}

	logFile := orm.File{
		ID:         result.ID,
		BucketName: payload.BucketName,
		Cdn:        result.Cdn,
		FileName:   result.FileName,
		Extension:  result.Extension,
		Path:       result.Path,
		FullPath:   result.FullPath,
	}
	if err := Database.Save(&logFile).Error; err != nil {
		infra.AppLog.EventError(err, "UPLOAD_SAVE_FAILURE", map[string]any{
			"file_id": result.ID,
		}, "",
			infra.WithComponent(infra.ComponentDB),
			infra.WithOperation("create"),
			infra.WithLogKind(infra.LogKindError),
			infra.WithField("table", "files"))
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

	var result model.UploadResult
	p.FileBase64 = stripBase64DataURIPrefix(p.FileBase64)

	jsonData, err := json.Marshal(p)
	if err != nil {
		infra.AppLog.EventError(err, "UPLOAD_MARSHAL_FAILURE", nil, "",
			infra.WithComponent(infra.ComponentUpload),
			infra.WithOperation("marshal_request"),
			infra.WithLogKind(infra.LogKindError))
		return result, err
	}

	uploadURL := infra.GetEnv("UPLOAD_SERVICE_URL", "")
	if uploadURL == "" {
		uploadURL = "https://fourwd.as.r.appspot.com/api/v1/upload/"
	}

	req, err := http.NewRequest(http.MethodPost, uploadURL, bytes.NewBuffer(jsonData))
	if err != nil {
		infra.AppLog.EventError(err, "UPLOAD_REQUEST_BUILD_FAILURE", nil, "",
			infra.WithComponent(infra.ComponentUpload),
			infra.WithOperation("build_request"),
			infra.WithLogKind(infra.LogKindError))
		return result, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	response, err := kit.NewHttpClient(30).Do(req)
	if err != nil {
		infra.AppLog.EventError(err, "UPLOAD_EXECUTE_FAILURE", map[string]any{
			"url": uploadURL,
		}, "",
			infra.WithComponent(infra.ComponentUpload),
			infra.WithOperation("post"),
			infra.WithLogKind(infra.LogKindError))
		return result, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		infra.AppLog.EventError(err, "UPLOAD_READ_FAILURE", nil, "",
			infra.WithComponent(infra.ComponentUpload),
			infra.WithOperation("read_response"),
			infra.WithLogKind(infra.LogKindError))
		return result, err
	}

	var resp ApiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		infra.AppLog.EventError(err, "UPLOAD_UNMARSHAL_FAILURE", nil, "",
			infra.WithComponent(infra.ComponentUpload),
			infra.WithOperation("unmarshal_response"),
			infra.WithLogKind(infra.LogKindError))
		return result, err
	}
	return resp.Data, nil
}

func stripBase64DataURIPrefix(s string) string {
	for _, prefix := range []string{
		"data:image/png;base64,",
		"data:image/jpeg;base64,",
		"data:image/jpg;base64,",
	} {
		s = strings.TrimPrefix(s, prefix)
	}
	return s
}

func getBucketName(appID string) string {
	return "fourwd-auction"
}
