package hoteltrader

// ProviderCancelRQ representa la estructura de request GraphQL para Cancel
type ProviderCancelRQ struct {
	Query     string          `json:"query"`
	Variables CancelVariables `json:"variables"`
}

// CancelVariables contiene las variables para la mutation
type CancelVariables struct {
	Cancel *CancelRequestInput `json:"Cancel"`
}

// CancelRequestInput representa una petición de cancel
type CancelRequestInput struct {
	HTConfirmationCode string `json:"htConfirmationCode"`
}
