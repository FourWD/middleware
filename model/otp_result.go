package model

type OtpResult struct {
	Status string `json:"status"`
	Token  string `json:"token"`
	Refno  string `json:"refno"`
}

type OtpVeriyResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`

	Code   string `json:"code"`
	Errors errors `json:"errors"`
}

type errors struct {
	Detail  string `json:"detail"`
	Message string `json:"message"`
}
