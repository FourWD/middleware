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

func Payment2C2PInquiry(info model.Payment2C2PInquiry) (model.Payment2C2PInquiryResponse, error) {
	var reqResponse model.Payment2C2PInquiryResponse

	payload := jwt.MapClaims{
		"merchantID": viper.GetString("2c2p_merchant_id"),
		"invoiceNo":  info.InvoiceNo,
		"locale":     "th",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, err := token.SignedString([]byte(viper.GetString("2c2p_secret_key")))
	if err != nil {
		return reqResponse, err
	}

	url := viper.GetString("2c2p_payment_inquiry_url") //"https://sandbox-pgw.2c2p.com/payment/4.3/paymentInquiry"
	payloads := strings.NewReader("{\"payload\":\"" + tokenString + "\"}")
	req, _ := http.NewRequest("POST", url, payloads)

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/*+json")

	res, errs := http.DefaultClient.Do(req)
	if errs != nil {
		return reqResponse, errs
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	type Payload struct {
		Payload string `json:"payload"`
	}

	responseJson := new(Payload)
	if err := json.Unmarshal(body, &responseJson); err != nil {
		return reqResponse, err
	}

	reqResponse, errResponse := decodeCheckResponse(responseJson.Payload)
	if errResponse != nil {
		return reqResponse, errResponse
	}

	if reqResponse.RespCode != "0000" {
		return reqResponse, errors.New(reqResponse.RespDesc)
	}

	return reqResponse, nil
}

func decodeCheckResponse(inquiryResponseJwt string) (model.Payment2C2PInquiryResponse, error) {
	var customClaims model.Payment2C2PInquiryResponse

	token, err := jwt.Parse(inquiryResponseJwt, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(viper.GetString("2c2p_secret_key")), nil
	})

	if err != nil {
		return customClaims, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		customClaims.MerchantID = getStringClaim(claims, "merchantID")
		customClaims.InvoiceNo = getStringClaim(claims, "invoiceNo")
		customClaims.Amount = getFloat64Claim(claims, "amount")
		customClaims.CurrencyCode = getStringClaim(claims, "currencyCode")
		customClaims.TransactionDateTime = getStringClaim(claims, "transactionDateTime")
		customClaims.ChannelCode = getStringClaim(claims, "channelCode")
		customClaims.ApprovalCode = getStringClaim(claims, "approvalCode")
		customClaims.ReferenceNo = getStringClaim(claims, "referenceNo")
		customClaims.AccountNo = getStringClaim(claims, "accountNo")
		// customClaims.CardToken = claims["cardToken"].(string)
		customClaims.IssuerCountry = getStringClaim(claims, "issuerCountry")
		customClaims.UserDefined1 = getStringClaim(claims, "userDefined1")
		customClaims.UserDefined2 = getStringClaim(claims, "userDefined2")
		customClaims.UserDefined3 = getStringClaim(claims, "userDefined3")
		customClaims.UserDefined4 = getStringClaim(claims, "userDefined4")
		customClaims.UserDefined5 = getStringClaim(claims, "userDefined5")
		customClaims.CardType = getStringClaim(claims, "cardType")
		customClaims.RespCode = getStringClaim(claims, "respCode")
		customClaims.RespDesc = getStringClaim(claims, "respDesc")
		// fmt.Println("Resp Payment Inquiry : " + customClaims.RespCode + " " + customClaims.RespDesc)
		return customClaims, nil
	}

	return customClaims, err
}
