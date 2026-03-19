package hoteltrader

// ProviderBookRQ representa la estructura de request GraphQL para Book
type ProviderBookRQ struct {
	Query     string        `json:"query"`
	Variables BookVariables `json:"variables"`
}

// BookVariables contiene las variables para la mutation
type BookVariables struct {
	Book *BookRequestInput `json:"Book"`
}

// BookRequestInput representa la petición de book al proveedor
type BookRequestInput struct {
	ClientConfirmationCode string          `json:"clientConfirmationCode"`
	OtaConfirmationCode    string          `json:"otaConfirmationCode"`
	OtaClientName          string          `json:"otaClientName"`
	SpecialRequests        []string        `json:"specialRequests,omitempty"`
	PaymentInformation     interface{}     `json:"paymentInformation,omitempty"`
	Rooms                  []BookRoomInput `json:"rooms"`
}

// BookRoomInput representa una habitación en la petición de book
type BookRoomInput struct {
	HTIdentifier               string              `json:"htIdentifier"`
	ClientRoomConfirmationCode string              `json:"clientRoomConfirmationCode"`
	RoomSpecialRequests        []string            `json:"roomSpecialRequests,omitempty"`
	Rates                      *BookRatesInput     `json:"rates,omitempty"`
	Occupancy                  *BookOccupancyInput `json:"occupancy,omitempty"`
	Guests                     []BookGuestInput    `json:"guests"`
}

// BookRatesInput representa las tarifas en el book (desde quote en ExtraParams)
type BookRatesInput struct {
	NetPrice      float64   `json:"netPrice"`
	Tax           float64   `json:"tax"`
	GrossPrice    float64   `json:"grossPrice"`
	PayAtProperty float64   `json:"payAtProperty,omitempty"`
	DailyPrice    []float64 `json:"dailyPrice,omitempty"`
	DailyTax      []float64 `json:"dailyTax,omitempty"`
}

// BookOccupancyInput representa la ocupación en el book (desde quote en ExtraParams)
type BookOccupancyInput struct {
	GuestAges string `json:"guestAges"`
}

// BookGuestInput representa un huésped en la petición de book
type BookGuestInput struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email,omitempty"`
	Adult     bool   `json:"adult"`
	Age       int    `json:"age,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Primary   bool   `json:"primary"`
}
