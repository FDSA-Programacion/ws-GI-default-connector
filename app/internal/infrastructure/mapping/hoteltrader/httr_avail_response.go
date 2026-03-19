package hoteltrader

// ProviderAvailRS representa la estructura de respuesta GraphQL para Avail
type ProviderAvailRS struct {
	Data   AvailData  `json:"data"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GetErrors implementa la interfaz GraphQLResponse
func (r *ProviderAvailRS) GetErrors() []GraphQLError {
	return r.Errors
}

// AvailData contiene los datos de la respuesta
type AvailData struct {
	GetPropertiesByIds GetPropertiesByIdsResponse `json:"getPropertiesByIds"`
}

// GetPropertiesByIdsResponse contiene las propiedades
type GetPropertiesByIdsResponse struct {
	Properties []PropertyResponseEntity `json:"properties"`
}

// PropertyResponseEntity representa una propiedad
type PropertyResponseEntity struct {
	PropertyID       int            `json:"propertyId"` // Es NUMBER en JSON, no string
	PropertyName     string         `json:"propertyName"`
	Occupancies      []Occupancy    `json:"occupancies"`
	Rooms            []RoomResponse `json:"rooms"`
	ShortDescription string         `json:"shortDescription"`
	LongDescription  string         `json:"longDescription"`
	City             string         `json:"city"`
	Latitude         string         `json:"latitude"`  // Viene como string en el JSON
	Longitude        string         `json:"longitude"` // Viene como string en el JSON
	StarRating       float64        `json:"starRating"`
	HotelImageURL    *string        `json:"hotelImageUrl"`
}

// Occupancy representa una ocupación
type Occupancy struct {
	OccupancyRefID int    `json:"occupancyRefId"` // Es NUMBER en JSON
	CheckInDate    string `json:"checkInDate"`
	CheckOutDate   string `json:"checkOutDate"`
	GuestAges      string `json:"guestAges"` // Viene como string "30,30" en el JSON
}

// RoomResponse representa una habitación
type RoomResponse struct {
	OccupancyRefID       int                      `json:"occupancyRefId"` // Es NUMBER en JSON
	HTIdentifier         string                   `json:"htIdentifier"`
	RoomName             string                   `json:"roomName"`
	RoomCode             string                   `json:"roomCode"`
	RateplanTag          string                   `json:"rateplanTag"`
	ShortDescription     string                   `json:"shortDescription"`
	NumRoomsAvail        int                      `json:"numRoomsAvail"`
	LongDescription      *string                  `json:"longDescription"`
	ConsolidatedComments string                   `json:"consolidatedComments"`
	PaymentType          *string                  `json:"paymentType"`
	RateInfo             RateInfo                 `json:"rateInfo"`
	MealplanOptions      MealplanOption           `json:"mealplanOptions"` // Es un objeto, no array
	Refundable           bool                     `json:"refundable"`
	IncludeNREF          bool                     `json:"includeNREF"`
	RateType             string                   `json:"rateType"`
	CancellationPolicies []HtCancellationPolicy   `json:"cancellationPolicies"`
}

// RateInfo contiene información de tarifas
type RateInfo struct {
	Bar              *bool            `json:"bar"`
	Binding          *bool            `json:"binding"`
	Commissionable   *bool            `json:"commissionable"`
	CommissionAmount *float64         `json:"commissionAmount"`
	Currency         string           `json:"currency"`
	NetPrice         float64          `json:"netPrice"`
	Tax              float64          `json:"tax"`
	GrossPrice       float64          `json:"grossPrice"`
	PayAtProperty    float64          `json:"payAtProperty"` // Puede ser un monto o un booleano según el proveedor
	DailyPrice       []float64        `json:"dailyPrice"`
	DailyTax         []float64        `json:"dailyTax"`
	AggregateTaxInfo AggregateTaxInfo `json:"aggregateTaxInfo"`
	TaxInfo          []TaxInfoDetail  `json:"taxInfo"`
}

// AggregateTaxInfo contiene información agregada de impuestos
type AggregateTaxInfo struct {
	PayAtBooking []TaxDetail `json:"payAtBooking"`
	PayAtProperty []TaxDetail `json:"payAtProperty"`
}

// TaxDetail representa un detalle de impuesto
type TaxDetail struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Date        string  `json:"date,omitempty"`
}

// TaxInfoDetail representa información detallada de impuestos
type TaxInfoDetail struct {
	PayAtBooking []TaxDetail `json:"payAtBooking"`
	PayAtProperty []TaxDetail `json:"payAtProperty"`
}

// MealplanOption representa una opción de plan de comidas
type MealplanOption struct {
	MealplanDescription string `json:"mealplanDescription"`
	MealplanCode        string `json:"mealplanCode"`
	MealplanName        string `json:"mealplanName"`
}

// HtCancellationPolicy representa una política de cancelación
type HtCancellationPolicy struct {
	StartWindowTime  string `json:"startWindowTime"`
	EndWindowTime    string `json:"endWindowTime"`
	CancellationCharge float64 `json:"cancellationCharge"`
	Currency         string `json:"currency"`
	TimeZone         string `json:"timeZone"`
	TimeZoneUTC      string `json:"timeZoneUTC"`
}

// GraphQLError representa un error de GraphQL
type GraphQLError struct {
	Message   string   `json:"message"`
	Locations []Location `json:"locations,omitempty"`
	Path      []interface{} `json:"path,omitempty"`
}

// Location representa la ubicación de un error
type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}
