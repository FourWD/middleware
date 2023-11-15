package model

type UploadPayload struct {
	AppID      string `json:"app_id" query:"app_id"`
	BucketName string `json:"bucket_name" query:"bucket_name"`
	Path       string `json:"path" query:"path"`
	Filename   string `json:"filename" query:"filename"`
	FileBase64 string `json:"file_base_64" query:"file_base_64"`
}
