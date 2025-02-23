package model

type Payment2C2PInquiry struct {
	MerchantID string `json:"merchantID"`
	InvoiceNo  string `json:"invoiceNo"`
	Locale     string `json:"locale"`
}

type Payment2C2PInquiryResponse struct {
	MerchantID          string  `json:"merchantID"`
	InvoiceNo           string  `json:"invoiceNo"`
	Amount              float64 `json:"amount"`
	CurrencyCode        string  `json:"currencyCode"`
	TransactionDateTime string  `json:"transactionDateTime"`
	ChannelCode         string  `json:"channelCode"`
	ApprovalCode        string  `json:"approvalCode"`
	ReferenceNo         string  `json:"referenceNo"`
	AccountNo           string  `json:"accountNo"`
	IssuerCountry       string  `json:"issuerCountry"`
	UserDefined1        string  `json:"userDefined1"`
	UserDefined2        string  `json:"userDefined2"`
	UserDefined3        string  `json:"userDefined3"`
	UserDefined4        string  `json:"userDefined4"`
	UserDefined5        string  `json:"userDefined5"`
	CardType            string  `json:"cardType"`
	RespCode            string  `json:"respCode"`
	RespDesc            string  `json:"respDesc"`
}

// CardToken                     string  `json:"cardToken"`
