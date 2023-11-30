package model

type OtpResult struct {
	Status string `json:"status"`
	Token  string `json:"token"`
	Refno  string `json:"refno"`
}
