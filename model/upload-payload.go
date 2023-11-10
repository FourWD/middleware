package model

type UploadPayload struct {
	BucketName string `query:"bucket_name" json:"bucket_name"`
	Path       string `query:"path" json:"path"`
	Filename   string `query:"filename" json:"filename"`
	FileBase64 string `query:"file_base_64" json:"file_base_64"`
}
