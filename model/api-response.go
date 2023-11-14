package model

type ApiResponse struct {
	Status     int          `json:"status"`
	StatusCode string       `json:"status_code"`
	Message    string       `json:"message"`
	Data       UploadResult `json:"data"`
}
