package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/FourWD/middleware/model"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func OtpRequest(payload model.OtpRequestPayload, db gorm.DB) (model.OtpResult, error) {

	result, errUpload := otpRequestToServer(payload, viper.GetString("app_id"), viper.GetString("token.upload"))
	if errUpload != nil {
		return result, errUpload
	}

	// // SAVE TO LOG

	// logOtp := new(orm.LogOTP)
	// logOtp.OTP = result.Token
	// logOtp.RefCodeOTP = result.RefNo
	// logOtp.RequestDate = time.Now()

	// err := db.Save(&logOtp)
	// if err.Error != nil {
	// 	PrintError("error save file", "tb file")
	// }
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

	// payloadss := strings.NewReader("key=1792158286047316&secret=10da83ff9be3f7007eaa9ef3250c2547&msisdn=0908979774")
	payloadString := "key=" + params.Key + "&secret=" + params.Secret + "&msisdn=" + params.Mobile
	req, _ := http.NewRequest("POST", viper.GetString("api.sms_request"), strings.NewReader(payloadString))

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return *result, errors.New(err.Error())
	}
	// if res.StatusCode == 400 {
	// 	return *result, errors.New("error code 400")
	// }

	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	jsonBody := strings.ReplaceAll(string(body), "/", "")
	fmt.Println(string(body))

	otpUnmar := new(model.OtpResult)
	errUnmar := json.Unmarshal(body, &otpUnmar)
	if errUnmar != nil {
		fmt.Println("Error unmarshalling JSON:", errUnmar)
	}

	result.Refno = otpUnmar.Refno
	result.Status = otpUnmar.Status
	result.Token = otpUnmar.Token

	// jsonBody := `{"status":"success",
	// "token":"1234567891011121314abcdefghijklm",
	// "refno":"REF12" }`

	log := new(model.LogOtpRequest)
	log.ID = uuid.NewString()
	log.CreatedAt = time.Now()
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

func OtpVerify(payload model.OtpVerifyPayload, db gorm.DB) (model.OtpVeriyResult, error) {

	result, errVerify := otpVerifyServer(payload)
	if errVerify != nil {
		return result, errVerify
	}

	// SAVE TO LOG

	// logOtp := new(orm.LogOTP)
	// logOtp.VerifyDate = time.Now()

	// err := db.Save(&logOtp)
	// if err.Error != nil {
	// 	PrintError("error save file", "tb file")
	// }
	return result, nil
}

func otpVerifyServer(payload model.OtpVerifyPayload) (model.OtpVeriyResult, error) {
	result := new(model.OtpVeriyResult)

	app, err := GetOtpApp(payload.AppID)
	if err != nil {
		PrintError("cant get", "OTP init")
	}

	payloadString := "key=" + app.AppKey + "&secret=" + app.AppSecret + "&token=" + payload.Token + "&pin=" + payload.Pin
	// results, err := request(viper.GetString("api.sms_verify"), payloadString)
	req, _ := http.NewRequest("POST", viper.GetString("api.sms_verify"), strings.NewReader(payloadString))

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return *result, errors.New(err.Error())
	}
	// if res.StatusCode == 400 {
	// 	return *result, errors.New("error code 400")
	// }

	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	jsonBody := strings.ReplaceAll(string(body), "/", "")
	fmt.Println(string(body))

	otpUnmar := new(model.OtpVeriyResult)
	errUnmar := json.Unmarshal(body, &otpUnmar)
	if errUnmar != nil {
		fmt.Println("Error unmarshalling JSON:", errUnmar)
	}

	result.Status = otpUnmar.Status
	result.Message = otpUnmar.Message

	result.Code = otpUnmar.Code
	result.Errors = otpUnmar.Errors

	saveLog(app, payloadString, jsonBody)
	return *result, nil
}

func GetOtpApp(appID string) (model.AppOtp, error) {
	app := new(model.AppOtp)
	app.AppKey = "1792158286047316"
	app.AppSecret = "10da83ff9be3f7007eaa9ef3250c2547"
	return *app, nil
}

// func request(url string, body string) (string, error) {
// 	req, _ := http.NewRequest("POST", url, strings.NewReader(body))

// 	req.Header.Add("accept", "application/json")
// 	req.Header.Add("content-type", "application/x-www-form-urlencoded")
// 	res, err := http.DefausltClient.Do(req)
// 	if res.StatusCode == 400 {
// 		return "", err
// 	}
// 	defer res.Body.Close()

// 	result, _ := io.ReadAll(res.Body)
// 	jsonResult := strings.ReplaceAll(string(result), "/", "")
// 	return jsonResult, nil
// }

func saveLog(app model.AppOtp, payload string, response string) {
	log := new(model.LogOtpVerify)
	log.ID = uuid.NewString()
	log.CreatedAt = time.Now()
	log.AppID = app.ID
	log.AppKey = app.AppKey
	log.AppSecret = app.AppSecret
	log.Payload = payload
	log.Response = response
	Database.Save(log)
}
