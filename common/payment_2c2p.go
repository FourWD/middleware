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
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

// paymentHTTPClient is a dedicated HTTP client for payment requests with timeout
var paymentHTTPClient = &http.Client{
	Timeout: 60 * time.Second, // Longer timeout for payment operations
}

// payment2C2PPayloadResponse is the standard response structure from 2C2P API
type payment2C2PPayloadResponse struct {
	Payload string `json:"payload"`
}

// signJWTPayload creates a signed JWT token from claims using 2C2P secret key
func signJWTPayload(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(viper.GetString("2c2p_secret_key")))
}

// send2C2PRequest sends a POST request to 2C2P API and returns the response payload
func send2C2PRequest(url string, jwtPayload string) (string, error) {
	body := strings.NewReader(`{"payload":"` + jwtPayload + `"}`)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/*+json")

	res, err := paymentHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var response payment2C2PPayloadResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", err
	}

	return response.Payload, nil
}

// parse2C2PJWTResponse parses and validates a JWT response from 2C2P
func parse2C2PJWTResponse(jwtString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(viper.GetString("2c2p_secret_key")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

func Payment2C2P(request model.Payment2C2P) (model.Payment2C2PResponse, error) {
	var reqResponse model.Payment2C2PResponse

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

	tokenString, err := signJWTPayload(payload)
	if err != nil {
		return reqResponse, err
	}

	url := viper.GetString("2c2p_payment_request_url")
	responsePayload, err := send2C2PRequest(url, tokenString)
	if err != nil {
		return reqResponse, err
	}

	reqResponse, err = decodePaymentResponse(responsePayload)
	if err != nil {
		return reqResponse, err
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
