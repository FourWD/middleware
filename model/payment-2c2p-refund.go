package model

type Payment2C2PRefund struct {
	// MerchantID        string   `json:"merchantID"`        // "JT04",
	InvoiceNo string `json:"invoice_no"` //"15239536999999",
	// Description       string   `json:"description"`       // "item 1",
	Amount float32 `json:"amount"` //1000.00,
	// PaymentChannel    []string `json:"paymentChannel"`    //"paymentChannel": ["CC"]
	// FrontendReturnUrl string   `json:"frontendReturnUrl"` //: "https://aot-limousine-website-dot-limousine-421804.as.r.appspot.com/en", // ส่งเข้า
	// BackendReturnUrl  string   `json:"backendReturnUrl"`  //:
} // CurrencyCode      string   `json:"currencyCode"`      //  "THB",

type Payment2C2PRefundResponse struct {
	PaymentToken  string `json:"paymentToken"`
	RespCode      string `json:"respCode"`
	RespDesc      string `json:"respDesc"`
	WebPaymentUrl string `json:"webPaymentUrl"`
	InvoiceNo     string `json:"invoice_no"`
}

// <PaymentProcessRequest>
//   <version>3.8</version>
//   <merchantID>JT07</merchantID>
//   <invoiceNo>260121085327</invoiceNo>
//   <actionAmount>25.00</actionAmount>
//   <processType>R</processType>
// </PaymentProcessRequest>
