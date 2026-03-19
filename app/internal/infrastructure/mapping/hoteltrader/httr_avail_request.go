package hoteltrader

// ProviderAvailRQ representa la estructura de request GraphQL para Avail
type ProviderAvailRQ struct {
	Query     string                  `json:"query"`
	Variables SearchCriteriaVariables `json:"variables"`
}

// SearchCriteriaVariables contiene las variables para la query
type SearchCriteriaVariables struct {
	SearchCriteriaByIds *SearchCriteriaByIdsInput `json:"SearchCriteriaByIds"`
}

// SearchCriteriaByIdsInput representa los criterios de búsqueda
type SearchCriteriaByIdsInput struct {
	PropertyIds []string         `json:"propertyIds"`
	Occupancies []OccupancyInput `json:"occupancies"`
}

// OccupancyInput representa una ocupación en la petición
type OccupancyInput struct {
	OccupancyRefId int    `json:"occupancyRefId"`
	CheckInDate    string `json:"checkInDate"`
	CheckOutDate   string `json:"checkOutDate"`
	GuestAges      string `json:"guestAges"` // Formato: "30,30" como string
}
