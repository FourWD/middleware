package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/FourWD/middleware/model"
	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)

func Payment2c2pInquiry(info model.Payment2c2pInquiry) (model.Payment2c2pInquiryResponse, error) {
	var empty model.Payment2c2pInquiryResponse

	url := viper.GetString("2c2p_payment_inquiry_url") //"https://sandbox-pgw.2c2p.com/payment/4.3/paymentInquiry"

	// Define the payload
	// Request Params Ref : https://developer.2c2p.com/docs/api-payment-token-request-parameter
	payload := jwt.MapClaims{
		"merchantID": viper.GetString("2c2p_merchant_id"),
		"invoiceNo":  info.InvoiceNo,
		"locale":     "th",
	}

	// Define the secret key
	//secretKey := []byte(viper.GetString("2c2p_secret_key"))

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, err := token.SignedString([]byte(viper.GetString("2c2p_secret_key")))
	if err != nil {
		fmt.Println("Error signing token:", err)
		return empty, err
	}

	// payloadss := strings.NewReader("{\"payload\":\"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJtZXJjaGFudElEIjoiSlQwNCIsImludm9pY2VObyI6IjE1MjM5NTM2OTk5OTk5IiwiZGVzY3JpcHRpb24iOiJpdGVtIDEiLCJhbW91bnQiOjEwMDAsImN1cnJlbmN5Q29kZSI6IlRIQiJ9.m4Z_GIWWR9f31GZhs2yFNW6896xf9760rNBMRO9WtA8\"}")
	payloads := strings.NewReader("{\"payload\":\"" + tokenString + "\"}")

	// fmt.Println("default : ")
	// fmt.Println(payloadss)
	fmt.Println("new : ")
	fmt.Println(payloads)

	req, _ := http.NewRequest("POST", url, payloads)

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/*+json")

	res, errs := http.DefaultClient.Do(req)
	if errs != nil {
		fmt.Println("Error response api request :", errs)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(string(body))

	type Payload struct {
		Payload string `json:"payload"`
	}

	remarkUnmar := new(Payload)
	errUnmar := json.Unmarshal(body, &remarkUnmar)
	if errUnmar != nil {
		fmt.Println("Error unmarshalling JSON:", errUnmar)
	}
	resp, errss := decodePayment2c2pInquiryResponeJwt(remarkUnmar.Payload)
	if errss != nil {
		return resp, errss
	}
	return resp, nil
}

func decodePayment2c2pInquiryResponeJwt(inquiryResponseJwt string) (model.Payment2c2pInquiryResponse, error) {
	var customClaims model.Payment2c2pInquiryResponse
	// responsePayload := "{\"payload\": \"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ3ZWJQYXltZW50VXJsIjoiaHR0cHM6Ly9zYW5kYm94LXBndy11aS4yYzJwLmNvbS9wYXltZW50LzQuMS8jL3Rva2VuL2tTQW9wczlad2hvczhoU1RTZUxUVWNKMFVRaVZhYTZ2Qmk1YXo5UWlmRUUlMmJSZDY1Y00zZE55ZjRXNWFZVmlxemthajVzTGRUbW9lSSUyYjAyMSUyZllyb0tEYjRSbVZvcWc4YVAlMmJoT0VKRDB0JTJiZyUzZCIsInBheW1lbnRUb2tlbiI6ImtTQW9wczlad2hvczhoU1RTZUxUVWNKMFVRaVZhYTZ2Qmk1YXo5UWlmRUUrUmQ2NWNNM2ROeWY0VzVhWVZpcXprYWo1c0xkVG1vZUkrMDIxL1lyb0tEYjRSbVZvcWc4YVAraE9FSkQwdCtnPSIsInJlc3BDb2RlIjoiMDAwMCIsInJlc3BEZXNjIjoiU3VjY2VzcyJ9.0YQthKwZEjR9giHWc3mkce9ngQnCNi0asXFWPHP_81k\"}"
	// responseToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ3ZWJQYXltZW50VXJsIjoiaHR0cHM6Ly9zYW5kYm94LXBndy11aS4yYzJwLmNvbS9wYXltZW50LzQuMS8jL3Rva2VuL2tTQW9wczlad2hvczhoU1RTZUxUVWNKMFVRaVZhYTZ2Qmk1YXo5UWlmRUUlMmJSZDY1Y00zZE55ZjRXNWFZVmlxemthajVzTGRUbW9lSSUyYjAyMSUyZllyb0tEYjRSbVZvcWc4YVAlMmJoT0VKRDB0JTJiZyUzZCIsInBheW1lbnRUb2tlbiI6ImtTQW9wczlad2hvczhoU1RTZUxUVWNKMFVRaVZhYTZ2Qmk1YXo5UWlmRUUrUmQ2NWNNM2ROeWY0VzVhWVZpcXprYWo1c0xkVG1vZUkrMDIxL1lyb0tEYjRSbVZvcWc4YVAraE9FSkQwdCtnPSIsInJlc3BDb2RlIjoiMDAwMCIsInJlc3BEZXNjIjoiU3VjY2VzcyJ9.0YQthKwZEjR9giHWc3mkce9ngQnCNi0asXFWPHP_81k"
	//responseToken := inquiryResponseJwt
	//secret := []byte(viper.GetString("2c2p_secret_key")) // Merchant SHA Key

	token, err := jwt.Parse(inquiryResponseJwt, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(viper.GetString("2c2p_secret_key")), nil
	})

	if err != nil {
		fmt.Printf("Error parsing token: %v\n", err)
		return customClaims, err
	}

	// Validate the claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println("Token is valid")
		fmt.Println("Decoded payload:")
		fmt.Println(claims)

		customClaims.MerchantID = claims["merchantID"].(string)
		customClaims.InvoiceNo = claims["invoiceNo"].(string)
		customClaims.Amount = claims["amount"].(float64)
		customClaims.CurrencyCode = claims["currencyCode"].(string)
		customClaims.TransactionDateTime = claims["transactionDateTime"].(string)
		customClaims.AgentCode = claims["agentCode"].(string)
		customClaims.ChannelCode = claims["channelCode"].(string)
		customClaims.ApprovalCode = claims["approvalCode"].(string)
		customClaims.ReferenceNo = claims["referenceNo"].(string)
		customClaims.AccountNo = claims["accountNo"].(string)
		customClaims.CardToken = claims["cardToken"].(string)
		customClaims.IssuerCountry = claims["issuerCountry"].(string)
		customClaims.ECI = claims["eci"].(string)
		customClaims.InstallmentPeriod = claims["installmentPeriod"].(int)
		customClaims.InterestType = claims["interestType"].(string)
		customClaims.InterestRate = claims["interestRate"].(float64)
		customClaims.InstallmentMerchantAbsorbRate = claims["installmentMerchantAbsorbRate"].(float64)
		customClaims.RecurringUniqueID = claims["recurringUniqueID"].(string)
		customClaims.FXAmount = claims["fxAmount"].(float64)
		customClaims.FXRate = claims["fxRate"].(float64)
		customClaims.FXCurrencyCode = claims["fxCurrencyCode"].(string)
		customClaims.UserDefined1 = claims["userDefined1"].(string)
		customClaims.UserDefined2 = claims["userDefined2"].(string)
		customClaims.UserDefined3 = claims["userDefined3"].(string)
		customClaims.UserDefined4 = claims["userDefined4"].(string)
		customClaims.UserDefined5 = claims["userDefined5"].(string)
		customClaims.AcquirerReferenceNo = claims["acquirerReferenceNo"].(string)
		customClaims.AcquirerMerchantID = claims["acquirerMerchantId"].(string)
		customClaims.CardType = claims["cardType"].(string)
		customClaims.IdempotencyID = claims["idempotencyID"].(string)
		customClaims.RespCode = claims["respCode"].(string)
		customClaims.RespDesc = claims["respDesc"].(string)

		fmt.Println("Resp Payment Inquiry : " + customClaims.RespCode + " " + customClaims.RespDesc)

		return customClaims, nil
	}
	fmt.Println("Token is invalid:", err)
	return customClaims, err
}
