package hoteltrader

// ProviderPrebookRQ representa la estructura de request GraphQL para PreBook (Quote)
type ProviderPrebookRQ struct {
	Query     string          `json:"query"`
	Variables QuoteVariables  `json:"variables"`
}

// QuoteVariables contiene las variables para la query
type QuoteVariables struct {
	Quote []QuoteRequestInput `json:"Quote"`
}

// QuoteRequestInput representa una petición de quote
type QuoteRequestInput struct {
	HTIdentifier string                `json:"htIdentifier"`
	PropertyID   string                `json:"propertyId,omitempty"`
	RoomCode     string                `json:"roomCode,omitempty"`
	RateplanTag  string                `json:"rateplanTag,omitempty"`
	CheckIn      string                `json:"checkIn,omitempty"`
	CheckOut     string                `json:"checkOut,omitempty"`
	Occupancy    *QuoteOccupancyInput  `json:"occupancy,omitempty"`
	Rates        *QuoteRatesInput      `json:"rates,omitempty"`
}

// QuoteOccupancyInput representa la ocupación en el quote request
type QuoteOccupancyInput struct {
	GuestAges string `json:"guestAges"`
}

// QuoteRatesInput representa las tarifas en el quote request
type QuoteRatesInput struct {
	NetPrice      float64   `json:"netPrice"`
	Tax           float64   `json:"tax"`
	GrossPrice    float64   `json:"grossPrice"`
	PayAtProperty float64   `json:"payAtProperty,omitempty"`
	DailyPrice    []float64 `json:"dailyPrice,omitempty"`
	DailyTax      []float64 `json:"dailyTax,omitempty"`
}

