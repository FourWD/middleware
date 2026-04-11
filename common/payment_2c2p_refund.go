package common

import (
	"github.com/FourWD/middleware/model"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

func Payment2C2PRefund(request model.Payment2C2PRefund) (bool, error) {
	payload := jwt.MapClaims{
		"merchantID": viper.GetString("2c2p_merchant_id"),
		"invoiceNo":  request.InvoiceNo,
		"amount":     request.Amount,
	}

	tokenString, err := signJWTPayload(payload)
	if err != nil {
		return false, err
	}

	url := viper.GetString("2c2p_payment_refund_url")
	_, err = send2C2PRequest(url, tokenString)
	if err != nil {
		return false, err
	}

	return true, nil
}
