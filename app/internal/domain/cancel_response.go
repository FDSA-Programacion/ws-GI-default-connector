package domain

// ------------------------------------------------------------------
// Response: CancelResponse (JSON Model)
// ------------------------------------------------------------------

type CancelResponse struct {
	CancelRules struct {
		CancelRule []CancelRule `json:"cancelRule"`
	} `json:"cancelRules"`
	UniqueID interface{} `json:"uniqueID"`
}

type CancelRule struct {
	Amount        float64     `json:"amount"`
	CancelByDate  string      `json:"cancelByDate"`
	CurrencyCode  string      `json:"currencyCode"`
	DecimalPlaces int         `json:"decimalPlaces"`
	NmbrOfNights  interface{} `json:"nmbrOfNights"`
	PaymentCard   interface{} `json:"paymentCard"`
	Percent       interface{} `json:"percent"`
	Type          interface{} `json:"type"`
}
