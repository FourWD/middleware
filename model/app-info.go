package model

type AppInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Env        string `json:"env"`
	GaeVersion string `json:"gae_version"`
}

// GaeService string `json:"gae_service"`
// GaeProject string `json:"gae_project"`
