package giresponsecommon

type CancelPenalties struct {
	CancelPenalty []CancelPenalty `json:"cancelPenalty"`
}

type CancelPenalty struct {
	End                string              `json:"end"`
	Start              string              `json:"start"`
	AmountPercent      AmountPercent       `json:"amountPercent"`
	PenaltyDescription *PenaltyDescription `json:"penaltyDescription,omitempty"`
}

type PenaltyDescription struct {
	Text string `json:"text"`
}

type AmountPercent struct {
	Amount       string `json:"amount,omitempty"`
	CurrencyCode string `json:"currencyCode"`
	Percent      string `json:"percent,omitempty"`
	NmbrOfNights string `json:"nmbrOfNights,omitempty"`
}
