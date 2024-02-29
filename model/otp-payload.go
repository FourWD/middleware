package model

// type OtpRequestPayload struct {
// 	AppID  string `json:"app_id"`
// 	Mobile string `json:"mobile"`
// }

type OtpVerifyPayload struct {
	Token string `json:"token"`
	Pin   string `json:"pin"`
}

type OtpRequestParams struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
	Mobile string `json:"mobile"`
}

type OtpVerifyParams struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
	Token  string `json:"token"`
	Pin    string `json:"pin"`
}
