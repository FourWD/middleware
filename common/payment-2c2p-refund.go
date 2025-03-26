package common

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/FourWD/middleware/model"
	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)

func Payment2C2PRefund(request model.Payment2C2PRefund) (bool, error) {
	// func Payment2C2PRefund(request model.Payment2C2PRefund) (model.Payment2C2PRefundResponse, error) {
	// log.Println("InvoiceNo", request.InvoiceNo)
	// log.Println("Amount", request.Amount)

	//var reqResponse model.Payment2C2PRefundResponse

	// merchantID := viper.GetString("2c2p_merchant_id")
	url := viper.GetString("2c2p_payment_refund_url") // https://sandbox-pgw.2c2p.com/payment/4.3/paymenttoken"
	payload := jwt.MapClaims{
		"merchantID": viper.GetString("2c2p_merchant_id"),
		"invoiceNo":  request.InvoiceNo,
		"amount":     request.Amount,
	}
	// 	<PaymentProcessRequest>
	//   <version>3.8</version>
	//   <merchantID>JT07</merchantID>
	//   <invoiceNo>260121085327</invoiceNo>
	//   <actionAmount>25.00</actionAmount>
	//   <processType>R</processType>
	// </PaymentProcessRequest>

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, err := token.SignedString([]byte(viper.GetString("2c2p_secret_key")))
	if err != nil {
		return false, err
	}

	payloads := strings.NewReader("{\"payload\":\"" + tokenString + "\"}")
	req, _ := http.NewRequest("POST", url, payloads)
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/*+json")

	res, errs := http.DefaultClient.Do(req)
	if errs != nil {
		return false, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	type response struct {
		Payload string `json:"payload"`
	}
	responseJson := new(response)
	if err := json.Unmarshal(body, &responseJson); err != nil {
		return false, err
	}

	// reqResponse, errResponse := decodePaymentResponse(responseJson.Payload)
	// if errResponse != nil {
	// 	return reqResponse, errResponse
	// }

	// if reqResponse.RespCode != "0000" {
	// 	return reqResponse, errors.New(reqResponse.RespDesc)
	// }

	// reqResponse.InvoiceNo = request.InvoiceNo
	// return reqResponse, nil
	return false, nil
}

func xx(p model.Payment2C2PRefund) string {
	type PaymentProcessRequest struct {
		XMLName      xml.Name `xml:"PaymentProcessRequest"`
		Version      string   `xml:"version"`
		MerchantID   string   `xml:"merchantID"`
		InvoiceNo    string   `xml:"invoiceNo"`
		ActionAmount string   `xml:"actionAmount"`
		ProcessType  string   `xml:"processType"`
	}

	request := PaymentProcessRequest{
		Version:      "4.3",
		MerchantID:   viper.GetString("2c2p_merchant_id"),
		InvoiceNo:    p.InvoiceNo,
		ActionAmount: fmt.Sprintf("%.2f", p.Amount),
		ProcessType:  "R",
	}

	// Convert struct to XML
	xmlData, err := xml.MarshalIndent(request, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	// Convert to string and print
	xmlString := string(xmlData)
	fmt.Println(xmlString)

	return xmlString
}
