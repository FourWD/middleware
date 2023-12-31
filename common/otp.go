package common

import (
	"time"

	"github.com/FourWD/middleware/model"
	"github.com/FourWD/middleware/orm"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func OtpRequest(payload model.OtpRequestPayload, db gorm.DB) (model.OtpResult, error) {

	result, errUpload := otpRequestToServer(payload, viper.GetString("app_id"), viper.GetString("token.upload"))
	if errUpload != nil {
		return result, errUpload
	}

	// SAVE TO LOG

	logOtp := new(orm.LogOTP)
	logOtp.OTP = result.Token
	logOtp.RefCodeOTP = result.RefNo
	logOtp.RequestDate = time.Now()

	err := db.Save(&logOtp)
	if err.Error != nil {
		PrintError("error save file", "tb file")
	}
	return result, nil
}

func otpRequestToServer(payload model.OtpRequestPayload, appID string, token string) (model.OtpResult, error) {
	result := new(model.OtpResult)

	// type Payload struct {
	// 	AppID  string `json:"app_id"`
	// 	Mobile string `json:"mobile"`
	// }

	type Params struct {
		Key    string `json:"key"`
		Secret string `json:"secret"`
		Mobile string `json:"mobile"`
	}

	app, err := GetOtpApp(payload.AppID)
	if err != nil {
		println("error to get app key and secret")
	}

	params := new(Params)
	params.Key = app.AppKey
	params.Secret = app.AppSecret
	params.Mobile = payload.Mobile

	payloadString := "key=" + params.Key + "&secret=" + params.Secret + "&msisdn=" + params.Mobile
	// req, _ := http.NewRequest("POST", viper.GetString("api.sms_request"), strings.NewReader(payloadString))

	// req.Header.Add("accept", "application/json")
	// req.Header.Add("content-type", "application/x-www-form-urlencoded")

	// res, err := http.DefaultClient.Do(req)

	// if res.StatusCode == 400 {
	// 	return common.FiberError(c, "400", err.Error())
	// }

	// defer res.Body.Close()

	// body, _ := io.ReadAll(res.Body)

	// jsonBody := strings.ReplaceAll(string(body), "/", "")
	// fmt.Println(string(body))
	jsonBody := `{"status":"success",
	"token":"1234567891011121314abcdefghijklm",
	"refno":"REF12" }`

	log := new(model.LogOtpRequest)
	// log.ID = uuid.NewString()
	// log.CreatedAt = time.Now()
	log.Mobile = params.Mobile
	log.AppID = payload.AppID
	log.AppKey = app.AppKey
	log.AppSecret = app.AppSecret
	log.Payload = payloadString
	log.Response = jsonBody
	Database.Save(log)

	// if res.StatusCode == 400 {
	// 	return common.FiberError(c, "400", "")
	// }

	//use mock	//message := `{"status":1, "message":"success", "data":` + jsonBody + `}`
	// c.Set("Content-Type", "application/json")
	// return c.SendString(string(message))

	return *result, nil
}

func GetOtpApp(appID string) (model.AppOtp, error) {
	app := new(model.AppOtp)
	app.AppKey = "1781967575388019"
	app.AppSecret = "a3c4a409ac4d7282a9adcf2600534149"
	return *app, nil
}
