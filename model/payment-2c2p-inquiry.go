package model

type Payment2c2pInquiry struct {
	MerchantID string `json:"merchantID"`
	InvoiceNo  string `json:"invoiceNo"`
	Locale     string `json:"locale"`
}

type Payment2c2pInquiryResponse struct {
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
