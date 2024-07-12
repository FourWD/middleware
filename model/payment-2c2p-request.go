package model

type Payment2c2pRequest struct {
	// MerchantID        string   `json:"merchantID"`        // "JT04",
	InvoiceNo         string   `json:"invoiceNo"`         //"15239536999999",
	Description       string   `json:"description"`       // "item 1",
	Amount            float32  `json:"amount"`            //1000.00,
	CurrencyCode      string   `json:"currencyCode"`      //  "THB",
	PaymentChannel    []string `json:"paymentChannel"`    //"paymentChannel": ["CC"]
	FrontendReturnUrl string   `json:"frontendReturnUrl"` //: "https://aot-limousine-website-dot-limousine-421804.as.r.appspot.com/en", // ส่งเข้า
	BackendReturnUrl  string   `json:"backendReturnUrl"`  //:
}

type Payment2c2pRequestResponse struct {
	PaymentToken  string `json:"paymentToken"`
	RespCode      string `json:"respCode"`
	RespDesc      string `json:"respDesc"`
	WebPaymentUrl string `json:"webPaymentUrl"`
}

// type PaymentResponse struct {
// 	MerchantID          string `json:"merchantID"`
// 	InvoiceNo           string `json:"invoiceNo"`
// 	AccountNo           string `json:"accountNo"`
// 	Amount              string `json:"amount"`
// 	CurrencyCode        string `json:"currencyCode"`
// 	TranRef             string `json:"tranRef"`
// 	ReferenceNo         string `json:"referenceNo"`
// 	ApprovalCode        string `json:"approvalCode"`
// 	Eci                 string `json:"eci"`
// 	TransactionDateTime string `json:"transactionDateTime"`
// 	RespCode            string `json:"respCode"`
// 	RespDesc            string `json:"respDesc"`
// }
