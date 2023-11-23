package model

type UploadPayload struct {
	BucketName string `json:"bucket_name" query:"bucket_name"`
	Path       string `json:"path" query:"path"`
	Filename   string `json:"filename" query:"filename"`
	FileBase64 string `json:"file_base_64" query:"file_base_64"`
}

type UploadResult struct {
	ID         string `json:"id"`
	BucketName string `json:"bucket_name"`
	Cdn        string `json:"cdn"`
	FileName   string `json:"file_name"`
	Extension  string `json:"extension"`
	Path       string `json:"path"`
	FullPath   string `json:"full_path"`
}
