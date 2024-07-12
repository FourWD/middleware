package model

type PaymentRequest2C2P struct {
	// MerchantID        string   `json:"merchantID"`        // "JT04",
	InvoiceNo         string   `json:"invoiceNo"`         //"15239536999999",
	Description       string   `json:"description"`       // "item 1",
	Amount            float32  `json:"amount"`            //1000.00,
	CurrencyCode      string   `json:"currencyCode"`      //  "THB",
	PaymentChannel    []string `json:"paymentChannel"`    //"paymentChannel": ["CC"]
	FrontendReturnUrl string   `json:"frontendReturnUrl"` //: "https://aot-limousine-website-dot-limousine-421804.as.r.appspot.com/en", // ส่งเข้า
	BackendReturnUrl  string   `json:"backendReturnUrl"`  //:
}

type PaymentRequestRespones2C2P struct {
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

type PaymentInquiryRequest2C2P struct {
	MerchantID string `json:"merchantID"`
	InvoiceNo  string `json:"invoiceNo"`
	Locale     string `json:"locale"`
}

type PaymentInquiryResponse2C2P struct {
	MerchantID                    string  `json:"merchantID"`
	InvoiceNo                     string  `json:"invoiceNo"`
	Amount                        float64 `json:"amount"`
	CurrencyCode                  string  `json:"currencyCode"`
	TransactionDateTime           string  `json:"transactionDateTime"`
	AgentCode                     string  `json:"agentCode"`
	ChannelCode                   string  `json:"channelCode"`
	ApprovalCode                  string  `json:"approvalCode"`
	ReferenceNo                   string  `json:"referenceNo"`
	AccountNo                     string  `json:"accountNo"`
	CardToken                     string  `json:"cardToken"`
	IssuerCountry                 string  `json:"issuerCountry"`
	ECI                           string  `json:"eci"`
	InstallmentPeriod             int     `json:"installmentPeriod"`
	InterestType                  string  `json:"interestType"`
	InterestRate                  float64 `json:"interestRate"`
	InstallmentMerchantAbsorbRate float64 `json:"installmentMerchantAbsorbRate"`
	RecurringUniqueID             string  `json:"recurringUniqueID"`
	FXAmount                      float64 `json:"fxAmount"`
	FXRate                        float64 `json:"fxRate"`
	FXCurrencyCode                string  `json:"fxCurrencyCode"`
	UserDefined1                  string  `json:"userDefined1"`
	UserDefined2                  string  `json:"userDefined2"`
	UserDefined3                  string  `json:"userDefined3"`
	UserDefined4                  string  `json:"userDefined4"`
	UserDefined5                  string  `json:"userDefined5"`
	AcquirerReferenceNo           string  `json:"acquirerReferenceNo"`
	AcquirerMerchantID            string  `json:"acquirerMerchantId"`
	CardType                      string  `json:"cardType"`
	IdempotencyID                 string  `json:"idempotencyID"`
	RespCode                      string  `json:"respCode"`
	RespDesc                      string  `json:"respDesc"`
}
