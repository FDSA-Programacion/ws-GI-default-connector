package hoteltrader

import (
	"encoding/json"
	"strconv"
	"strings"
)

// flexIntSlice acepta en JSON array de números o string (p. ej. "30" o "30,25") para guestAges.
type flexIntSlice []int

func (s *flexIntSlice) UnmarshalJSON(data []byte) error {
	var arr []int
	if err := json.Unmarshal(data, &arr); err == nil {
		*s = arr
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	str = strings.TrimSpace(str)
	if str == "" {
		*s = nil
		return nil
	}
	// Si parece JSON array como "[30,25]"
	if strings.HasPrefix(str, "[") {
		if err := json.Unmarshal([]byte(str), &arr); err == nil {
			*s = arr
			return nil
		}
	}
	// Comma-separated
	parts := strings.Split(str, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return err
		}
		result = append(result, n)
	}
	*s = result
	return nil
}

// ProviderBookRS representa la estructura de respuesta GraphQL para Book
type ProviderBookRS struct {
	Data   BookData       `json:"data"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GetErrors implementa la interfaz GraphQLResponse
func (r *ProviderBookRS) GetErrors() []GraphQLError {
	return r.Errors
}

// BookData contiene los datos de la respuesta
type BookData struct {
	Book BookResponse `json:"book"`
}

// BookResponse representa una respuesta de book
type BookResponse struct {
	HTConfirmationCode       string             `json:"htConfirmationCode"`
	ClientConfirmationCode   string             `json:"clientConfirmationCode"`
	OTAConfirmationCode      string             `json:"otaConfirmationCode"`
	ConsolidatedComments     string             `json:"consolidatedComments"`
	ConsolidatedHTMLComments string             `json:"consolidatedHTMLComments"`
	BookingDate              string             `json:"bookingDate"`
	SpecialRequests          []string           `json:"specialRequests"`
	PropertyDetails          PropertyDetails    `json:"propertyDetails"`
	Rooms                    []BookRoomResponse `json:"rooms"`
}

// PropertyDetails contiene detalles de la propiedad
type PropertyDetails struct {
	PropertyID            int            `json:"propertyId"` // La API puede devolverlo como string o número
	PropertyName          string         `json:"propertyName"`
	Address               AddressDetails `json:"address"`
	CheckInTime           string         `json:"checkInTime"`
	CheckOutTime          string         `json:"checkOutTime"`
	City                  string         `json:"city"`
	HotelImageURL         string         `json:"hotelImageUrl"`
	Latitude              string         `json:"latitude"`  // La API puede devolverlo como string
	Longitude             string         `json:"longitude"` // La API puede devolverlo como string
	StarRating            *float32       `json:"starRating"`
	CheckInPolicy         string         `json:"checkInPolicy"`
	MinAdultAgeForCheckIn int            `json:"minAdultAgeForCheckIn"`
}

// AddressDetails contiene detalles de dirección
type AddressDetails struct {
	Address1    string `json:"address1"`
	Address2    string `json:"address2"`
	CityName    string `json:"cityName"`
	CountryCode string `json:"countryCode"`
	StateName   string `json:"stateName"`
	ZipCode     string `json:"zipCode"`
}

// BookRoomOccupancy contiene occupancy de una habitación (guestAges puede venir como string o array).
type BookRoomOccupancy struct {
	GuestAges flexIntSlice `json:"guestAges"`
}

// BookRoomResponse representa una habitación en la respuesta de book
type BookRoomResponse struct {
	CancellationDate           string                 `json:"cancellationDate"`
	CancellationFee            *float64               `json:"cancellationFee"` // Puede ser null
	Cancelled                  bool                   `json:"cancelled"`
	CancellationPolicies       []HtCancellationPolicy `json:"cancellationPolicies"`
	CheckInDate                string                 `json:"checkInDate"`
	CheckOutDate               string                 `json:"checkOutDate"`
	ClientRoomConfirmationCode string                 `json:"clientRoomConfirmationCode"`
	HTRoomConfirmationCode     string                 `json:"htRoomConfirmationCode"`
	CRSConfirmationCode        string                 `json:"crsConfirmationCode"`
	CRSCancelConfirmationCode  string                 `json:"crsCancelConfirmationCode"`
	PMSConfirmationCode        string                 `json:"pmsConfirmationCode"`
	Refundable                 bool                   `json:"refundable"`
	RoomName                   string                 `json:"roomName"`
	RateplanTag                string                 `json:"rateplanTag"`
	MealplanOptions            MealplanOption         `json:"mealplanOptions"` // La API devuelve un objeto, no array
	Rates                      BookRatesDetails       `json:"rates"`
	Occupancy                  BookRoomOccupancy      `json:"occupancy"`
	Guests                     []GuestDetails         `json:"guests"`
	RoomSpecialRequests []string       `json:"roomSpecialRequests"`
}

// BookRatesDetails contiene detalles de tarifas en book
type BookRatesDetails struct {
	Bar              *bool            `json:"bar"`    // Puede ser null
	Binding          *bool            `json:"binding"` // Puede ser null
	Commissionable   bool             `json:"commissionable"`
	CommissionAmount float64          `json:"commissionAmount"`
	CurrencyCode     string           `json:"currencyCode"`
	NetPrice         float64          `json:"netPrice"`
	Tax              float64          `json:"tax"`
	GrossPrice       float64          `json:"grossPrice"`
	DailyPrice       []float64        `json:"dailyPrice"`
	DailyTax         []float64        `json:"dailyTax"`
	PayAtProperty    float64          `json:"payAtProperty"` // La API devuelve número, no bool
	AggregateTaxInfo AggregateTaxInfo `json:"aggregateTaxInfo"`
}

// GuestDetails contiene detalles de un huésped
type GuestDetails struct {
	Adult     bool   `json:"adult"`
	Age       *int   `json:"age"` // Puede ser null
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Primary   bool   `json:"primary"`
}
