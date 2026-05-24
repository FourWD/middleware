package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/FourWD/middleware/infra"
	"github.com/FourWD/middleware/kit"
	"github.com/FourWD/middleware/model"
	"github.com/google/uuid"
)

var otpHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

func OtpRequest(mobile string) (model.OtpResult, error) {
	return otpRequestToServer(mobile)
}

func otpRequestToServer(mobile string) (model.OtpResult, error) {
	var result model.OtpResult
	app := getOtpApp()

	payload := "key=" + app.AppKey + "&secret=" + app.AppSecret + "&msisdn=" + mobile
	body, err := postOtpForm(infra.GetEnv("OTP_URL_REQUEST", ""), payload)
	if err != nil {
		infra.AppLog.EventError(err, "OTP_REQUEST_FAILURE", map[string]any{
			"mobile_last_4": kit.MaskMobile(mobile),
		}, "",
			infra.WithComponent(infra.ComponentOTP),
			infra.WithOperation("request"),
			infra.WithLogKind(infra.LogKindError))
		return result, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		infra.AppLog.EventError(err, "OTP_REQUEST_UNMARSHAL_FAILURE", map[string]any{
			"mobile_last_4": kit.MaskMobile(mobile),
		}, "",
			infra.WithComponent(infra.ComponentOTP),
			infra.WithOperation("request"),
			infra.WithLogKind(infra.LogKindError))
		return result, fmt.Errorf("unmarshal otp response: %w", err)
	}

	Database.Save(&model.LogOtpRequest{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		Mobile:    mobile,
		AppID:     app.ID,
		Response:  strings.ReplaceAll(string(body), "/", ""),
	})

	return result, nil
}

func OtpVerify(payload model.OtpVerifyPayload) (model.OtpVeriyResult, error) {
	return otpVerifyServer(payload)
}

func otpVerifyServer(payload model.OtpVerifyPayload) (model.OtpVeriyResult, error) {
	var result model.OtpVeriyResult
	app := getOtpApp()

	form := "key=" + app.AppKey + "&secret=" + app.AppSecret + "&token=" + payload.Token + "&pin=" + payload.Pin
	body, err := postOtpForm(infra.GetEnv("OTP_URL_VERIFY", ""), form)
	if err != nil {
		infra.AppLog.EventError(err, "OTP_VERIFY_FAILURE", nil, "",
			infra.WithComponent(infra.ComponentOTP),
			infra.WithOperation("verify"),
			infra.WithLogKind(infra.LogKindError))
		return result, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		infra.AppLog.EventError(err, "OTP_VERIFY_UNMARSHAL_FAILURE", nil, "",
			infra.WithComponent(infra.ComponentOTP),
			infra.WithOperation("verify"),
			infra.WithLogKind(infra.LogKindError))
		return result, fmt.Errorf("unmarshal otp verify response: %w", err)
	}

	saveLog(app, strings.ReplaceAll(string(body), "/", ""))
	return result, nil
}

func postOtpForm(url, body string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	res, err := otpHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return raw, nil
}

func getOtpApp() model.AppOtp {
	return model.AppOtp{
		ID:        infra.GetEnv("APP_ID", ""),
		AppKey:    infra.GetEnv("OTP_SMS_KEY", ""),
		AppSecret: infra.GetEnv("OTP_SMS_SECRET", ""),
	}
}

func saveLog(app model.AppOtp, response string) {
	Database.Save(&model.LogOtpVerify{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		AppID:     app.ID,
		Response:  response,
	})
}
