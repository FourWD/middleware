package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/FourWD/middleware/infra"
	"github.com/FourWD/middleware/model"
	"github.com/golang-jwt/jwt/v5"
)

// paymentHTTPClient is a dedicated client with a longer timeout — payment
// gateways take longer than typical API calls.
var paymentHTTPClient = &http.Client{Timeout: 60 * time.Second}

type payment2C2PPayloadResponse struct {
	Payload string `json:"payload"`
}

// signJWTPayload signs claims with the 2C2P merchant secret (HS256).
func signJWTPayload(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(payment2C2PSecret())
}

// send2C2PRequest POSTs a JWT-wrapped payload to the given 2C2P endpoint
// and returns the JWT string from the response body.
func send2C2PRequest(url string, jwtPayload string) (string, error) {
	body, err := json.Marshal(map[string]string{"payload": jwtPayload})
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/*+json")

	res, err := paymentHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	var response payment2C2PPayloadResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", fmt.Errorf("unmarshal body: %w", err)
	}
	return response.Payload, nil
}

// parse2C2PJWTResponse verifies an HS256 JWT signed with the 2C2P secret
// and returns the claims.
func parse2C2PJWTResponse(jwtString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(jwtString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return payment2C2PSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func Payment2C2P(request model.Payment2C2P) (model.Payment2C2PResponse, error) {
	var reqResponse model.Payment2C2PResponse

	payload := jwt.MapClaims{
		"merchantID":        infra.GetEnv("PAYMENT_2C2P_MERCHANT_ID", ""),
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

	responsePayload, err := send2C2PRequest(infra.GetEnv("PAYMENT_2C2P_REQUEST_URL", ""), tokenString)
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

// decodePaymentResponse parses the JWT response from a Payment2C2P request
// and extracts the URL/token/code/desc claims.
func decodePaymentResponse(jwtString string) (model.Payment2C2PResponse, error) {
	var resp model.Payment2C2PResponse
	claims, err := parse2C2PJWTResponse(jwtString)
	if err != nil {
		return resp, err
	}
	resp.WebPaymentUrl = getStringClaim(claims, "webPaymentUrl")
	resp.PaymentToken = getStringClaim(claims, "paymentToken")
	resp.RespCode = getStringClaim(claims, "respCode")
	resp.RespDesc = getStringClaim(claims, "respDesc")
	return resp, nil
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	if v, ok := claims[key].(string); ok {
		return v
	}
	return ""
}

func getFloat64Claim(claims jwt.MapClaims, key string) float64 {
	if v, ok := claims[key].(float64); ok {
		return v
	}
	return 0
}

// payment2C2PSecret reads the merchant secret per call so an env rotation
// takes effect without a restart.
func payment2C2PSecret() []byte {
	return []byte(infra.GetEnv("PAYMENT_2C2P_SECRET", ""))
}
