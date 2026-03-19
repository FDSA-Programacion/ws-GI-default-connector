package hoteltrader

// ProviderCancelRS representa la estructura de respuesta GraphQL para Cancel
type ProviderCancelRS struct {
	Data   CancelData     `json:"data"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GetErrors implementa la interfaz GraphQLResponse
func (r *ProviderCancelRS) GetErrors() []GraphQLError {
	return r.Errors
}

// CancelData contiene los datos de la respuesta
type CancelData struct {
	Cancel CancelResponse `json:"cancel"`
}

// CancelResponse representa una respuesta de cancel
type CancelResponse struct {
	HTConfirmationCode     string               `json:"htConfirmationCode"`
	ClientConfirmationCode string               `json:"clientConfirmationCode"`
	AllRoomsCancelled      bool                 `json:"allRoomsCancelled"`
	Rooms                  []CancelRoomResponse `json:"rooms"`
}

// CancelRoomResponse representa una habitación cancelada
type CancelRoomResponse struct {
	HTRoomConfirmationCode     string  `json:"htRoomConfirmationCode"`
	ClientRoomConfirmationCode string  `json:"clientRoomConfirmationCode"`
	Cancelled                  bool    `json:"cancelled"`
	Currency                   string  `json:"currency"`
	CancellationAmount         float64 `json:"cancellationAmount"`
	CancellationDate           string  `json:"cancellationDate"`
}
