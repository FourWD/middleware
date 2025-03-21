package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/FourWD/middleware/model"
	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)

func Payment2C2P(request model.Payment2C2P) (model.Payment2C2PResponse, error) {
	// log.Println("InvoiceNo", request.InvoiceNo)
	// log.Println("Amount", request.Amount)

	var reqResponse model.Payment2C2PResponse

	// merchantID := viper.GetString("2c2p_merchant_id")
	url := viper.GetString("2c2p_payment_request_url") // https://sandbox-pgw.2c2p.com/payment/4.3/paymenttoken"
	payload := jwt.MapClaims{
		"merchantID":        viper.GetString("2c2p_merchant_id"),
		"invoiceNo":         request.InvoiceNo,
		"description":       request.Description,
		"amount":            request.Amount,
		"currencyCode":      "THB",
		"paymentChannel":    request.PaymentChannel,
		"frontendReturnUrl": request.FrontendReturnUrl,
		"backendReturnUrl":  request.BackendReturnUrl,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, err := token.SignedString([]byte(viper.GetString("2c2p_secret_key")))
	if err != nil {
		return reqResponse, err
	}

	payloads := strings.NewReader("{\"payload\":\"" + tokenString + "\"}")
	req, _ := http.NewRequest("POST", url, payloads)
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/*+json")

	res, errs := http.DefaultClient.Do(req)
	if errs != nil {
		return reqResponse, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	type response struct {
		Payload string `json:"payload"`
	}
	responseJson := new(response)
	if err := json.Unmarshal(body, &responseJson); err != nil {
		return reqResponse, err
	}

	reqResponse, errResponse := decodePaymentResponse(responseJson.Payload)
	if errResponse != nil {
		return reqResponse, errResponse
	}

	if reqResponse.RespCode != "0000" {
		return reqResponse, errors.New(reqResponse.RespDesc)
	}

	reqResponse.InvoiceNo = request.InvoiceNo
	return reqResponse, nil
}

func decodePaymentResponse(requestResponseJwt string) (model.Payment2C2PResponse, error) {
	var customClaims model.Payment2C2PResponse

	token, err := jwt.Parse(requestResponseJwt, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(viper.GetString("2c2p_secret_key")), nil
	})

	if err != nil {
		return customClaims, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		customClaims.WebPaymentUrl = getStringClaim(claims, "webPaymentUrl")
		customClaims.PaymentToken = getStringClaim(claims, "paymentToken")
		customClaims.RespCode = getStringClaim(claims, "respCode")
		customClaims.RespDesc = getStringClaim(claims, "respDesc")
		return customClaims, nil
	}
	return customClaims, err
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, exists := claims[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getFloat64Claim(claims jwt.MapClaims, key string) float64 {
	if val, exists := claims[key]; exists {
		if num, ok := val.(float64); ok {
			return num
		}
	}
	return 0
}

// func getIntClaim(claims jwt.MapClaims, key string) int {
// 	if val, exists := claims[key]; exists {
// 		if num, ok := val.(float64); ok { // JSON numbers are float64 in Go
// 			return int(num)
// 		}
// 	}
// 	return 0
// }
