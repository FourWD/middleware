package common

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/FourWD/middleware/model"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

var otpHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

func OtpRequest(mobile string) (model.OtpResult, error) {
	result, errUpload := otpRequestToServer(mobile)
	if errUpload != nil {
		return result, errUpload
	}

	return result, nil
}

func otpRequestToServer(mobile string) (model.OtpResult, error) {
	result := new(model.OtpResult)

	type Params struct {
		Key    string `json:"key"`
		Secret string `json:"secret"`
		Mobile string `json:"mobile"`
	}

	app, err := getOtpApp()
	if err != nil {
		LogError("OTP_APP_ERROR", map[string]interface{}{"error": err.Error()}, "")
	}

	params := new(Params)
	params.Key = app.AppKey
	params.Secret = app.AppSecret
	params.Mobile = mobile

	payloadString := "key=" + params.Key + "&secret=" + params.Secret + "&msisdn=" + params.Mobile
	req, _ := http.NewRequest("POST", viper.GetString("sms.url_request"), strings.NewReader(payloadString))

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := otpHTTPClient.Do(req)
	if err != nil {
		return *result, errors.New(err.Error())
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return *result, errors.New("failed to read response body: " + err.Error())
	}

	jsonBody := strings.ReplaceAll(string(body), "/", "")

	otpUnmar := new(model.OtpResult)
	errUnmar := json.Unmarshal(body, &otpUnmar)
	if errUnmar != nil {
		LogError("OTP_UNMARSHAL_ERROR", map[string]interface{}{"error": errUnmar.Error()}, "")
	}

	result.Refno = otpUnmar.Refno
	result.Status = otpUnmar.Status
	result.Token = otpUnmar.Token

	log := new(model.LogOtpRequest)
	log.ID = uuid.NewString()
	log.CreatedAt = time.Now()
	log.Mobile = params.Mobile
	log.AppID = viper.GetString("app_id")
	log.Response = jsonBody
	Database.Save(log)

	return *result, nil
}

func OtpVerify(payload model.OtpVerifyPayload) (model.OtpVeriyResult, error) {
	result, errVerify := otpVerifyServer(payload)
	if errVerify != nil {
		return result, errVerify
	}

	return result, nil
}

func otpVerifyServer(payload model.OtpVerifyPayload) (model.OtpVeriyResult, error) {
	result := new(model.OtpVeriyResult)

	app, err := getOtpApp()
	if err != nil {
		LogError("OTP_APP_ERROR", map[string]interface{}{"error": err.Error()}, "")
	}

	payloadString := "key=" + app.AppKey + "&secret=" + app.AppSecret + "&token=" + payload.Token + "&pin=" + payload.Pin

	req, _ := http.NewRequest("POST", viper.GetString("sms.url_verify"), strings.NewReader(payloadString))

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := otpHTTPClient.Do(req)
	if err != nil {
		return *result, errors.New(err.Error())
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return *result, errors.New("failed to read response body: " + err.Error())
	}

	jsonBody := strings.ReplaceAll(string(body), "/", "")

	otpUnmar := new(model.OtpVeriyResult)
	errUnmar := json.Unmarshal(body, &otpUnmar)
	if errUnmar != nil {
		LogError("OTP_VERIFY_UNMARSHAL_ERROR", map[string]interface{}{"error": errUnmar.Error()}, "")
	}

	result.Status = otpUnmar.Status
	result.Message = otpUnmar.Message

	result.Code = otpUnmar.Code
	result.Errors = otpUnmar.Errors

	saveLog(app, jsonBody)
	return *result, nil
}

func getOtpApp() (model.AppOtp, error) {
	app := new(model.AppOtp)
	app.ID = viper.GetString("app_id")
	app.AppKey = viper.GetString("sms.sms_key")
	app.AppSecret = viper.GetString("sms.sms_secret")
	return *app, nil
}

func saveLog(app model.AppOtp, response string) {
	log := new(model.LogOtpVerify)
	log.ID = uuid.NewString()
	log.CreatedAt = time.Now()
	log.AppID = app.ID
	log.Response = response
	Database.Save(log)
}
