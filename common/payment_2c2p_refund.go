package common

import (
	"github.com/FourWD/middleware/infra"
	"github.com/FourWD/middleware/model"
	"github.com/golang-jwt/jwt/v5"
)

func Payment2C2PRefund(request model.Payment2C2PRefund) (bool, error) {
	payload := jwt.MapClaims{
		"merchantID": infra.GetEnv("PAYMENT_2C2P_MERCHANT_ID", ""),
		"invoiceNo":  request.InvoiceNo,
		"amount":     request.Amount,
	}

	tokenString, err := signJWTPayload(payload)
	if err != nil {
		return false, err
	}

	url := infra.GetEnv("PAYMENT_2C2P_REFUND_URL", "")
	_, err = send2C2PRequest(url, tokenString)
	if err != nil {
		return false, err
	}

	return true, nil
}
