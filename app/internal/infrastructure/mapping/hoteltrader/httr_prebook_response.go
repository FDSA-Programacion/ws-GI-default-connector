package hoteltrader

// ProviderPrebookRS representa la estructura de respuesta GraphQL para PreBook (Quote)
type ProviderPrebookRS struct {
	Data   PrebookData  `json:"data"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GetErrors implementa la interfaz GraphQLResponse
func (r *ProviderPrebookRS) GetErrors() []GraphQLError {
	return r.Errors
}

// PrebookData contiene los datos de la respuesta
type PrebookData struct {
	Quote QuoteResponse `json:"quote"` // Es un objeto, NO array
}

// QuoteResponse representa una respuesta de quote
type QuoteResponse struct {
	AggregateGrossPrice    float64      `json:"aggregateGrossPrice"`
	AggregateNetPrice      float64      `json:"aggregateNetPrice"`
	AggregateTax           float64      `json:"aggregateTax"`
	AggregatePayAtProperty float64      `json:"aggregatePayAtProperty"`
	ConsolidatedComments   string       `json:"consolidatedComments"`
	Rooms                  []QuoteRoom  `json:"rooms"`
}

// QuoteRoom representa una habitación en la respuesta de quote
type QuoteRoom struct {
	HTIdentifier        string                `json:"htIdentifier"`
	GrossPriceChanged   bool                  `json:"grossPriceChanged"`
	Refundable          bool                  `json:"refundable"`
	Message             string                 `json:"message"`
	CancellationPolicies []HtCancellationPolicy `json:"cancellationPolicies"`
	MealplanOptions     MealplanOption        `json:"mealplanOptions"` // Es objeto, NO array
	Rates               QuoteRateInfo         `json:"rates"`
}

// QuoteRateInfo contiene información de tarifas en quote
type QuoteRateInfo struct {
	Bar              *bool            `json:"bar"`
	Binding          *bool            `json:"binding"`
	Commissionable   bool             `json:"commissionable"`
	CommissionAmount *float64         `json:"commissionAmount"`
	Currency         string           `json:"currency"`
	NetPrice         float64          `json:"netPrice"`
	Tax              float64          `json:"tax"`
	GrossPrice       float64          `json:"grossPrice"`
	PayAtProperty    float64          `json:"payAtProperty"` // Es NUMBER, no bool
	DailyPrice       []float64        `json:"dailyPrice"`
	DailyTax         []float64        `json:"dailyTax"`
	AggregateTaxInfo AggregateTaxInfo `json:"aggregateTaxInfo"`
}
