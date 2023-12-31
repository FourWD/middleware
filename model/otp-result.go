package model

type OtpResult struct {
	Status string `json:"status"`
	Token  string `json:"token"`
	RefNo  string `json:"ref_no"`
}
