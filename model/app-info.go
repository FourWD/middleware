package model

type AppInfo struct {
	AppVersion string `json:"app_version"`
	GaeService string `json:"gae_service"`
	GaeVersion string `json:"gae_version"`
	BucketName string `json:"bucket_name"`
	Env        string `json:"env"`
	WakeUpUrl  string `json:"wake_up_url"`
}
