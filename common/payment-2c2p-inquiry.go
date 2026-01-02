package common

import (
	"errors"

	"github.com/FourWD/middleware/model"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

func Payment2C2PInquiry(info model.Payment2C2PInquiry) (model.Payment2C2PInquiryResponse, error) {
	var reqResponse model.Payment2C2PInquiryResponse

	payload := jwt.MapClaims{
		"merchantID": viper.GetString("2c2p_merchant_id"),
		"invoiceNo":  info.InvoiceNo,
		"locale":     "th",
	}

	tokenString, err := signJWTPayload(payload)
	if err != nil {
		return reqResponse, err
	}

	url := viper.GetString("2c2p_payment_inquiry_url")
	responsePayload, err := send2C2PRequest(url, tokenString)
	if err != nil {
		return reqResponse, err
	}

	reqResponse, err = decodeInquiryResponse(responsePayload)
	if err != nil {
		return reqResponse, err
	}

	if reqResponse.RespCode != "0000" {
		return reqResponse, errors.New(reqResponse.RespDesc)
	}

	return reqResponse, nil
}

func decodeInquiryResponse(jwtString string) (model.Payment2C2PInquiryResponse, error) {
	var response model.Payment2C2PInquiryResponse

	claims, err := parse2C2PJWTResponse(jwtString)
	if err != nil {
		return response, err
	}

	response.MerchantID = getStringClaim(claims, "merchantID")
	response.InvoiceNo = getStringClaim(claims, "invoiceNo")
	response.Amount = getFloat64Claim(claims, "amount")
	response.CurrencyCode = getStringClaim(claims, "currencyCode")
	response.TransactionDateTime = getStringClaim(claims, "transactionDateTime")
	response.ChannelCode = getStringClaim(claims, "channelCode")
	response.ApprovalCode = getStringClaim(claims, "approvalCode")
	response.ReferenceNo = getStringClaim(claims, "referenceNo")
	response.AccountNo = getStringClaim(claims, "accountNo")
	response.IssuerCountry = getStringClaim(claims, "issuerCountry")
	response.UserDefined1 = getStringClaim(claims, "userDefined1")
	response.UserDefined2 = getStringClaim(claims, "userDefined2")
	response.UserDefined3 = getStringClaim(claims, "userDefined3")
	response.UserDefined4 = getStringClaim(claims, "userDefined4")
	response.UserDefined5 = getStringClaim(claims, "userDefined5")
	response.CardType = getStringClaim(claims, "cardType")
	response.RespCode = getStringClaim(claims, "respCode")
	response.RespDesc = getStringClaim(claims, "respDesc")

	return response, nil
}
