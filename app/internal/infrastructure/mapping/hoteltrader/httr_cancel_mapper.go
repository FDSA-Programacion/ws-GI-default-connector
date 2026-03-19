package hoteltrader

import (
	"regexp"
	"strconv"
	"strings"
	"ws-int-httr/internal/domain"
	giresponsecommon "ws-int-httr/internal/domain/gi_response_common"
	"ws-int-httr/internal/infrastructure/config"
	"ws-int-httr/internal/infrastructure/session"
)

// trailingIndexSuffix elimina el sufijo "-<número>" del final de un código de proveedor.
// Ej: "HT-8RPSZBSANDBOX-1" → "HT-8RPSZBSANDBOX"
var trailingIndexSuffix = regexp.MustCompile(`-\d+$`)

// ExtractHTConfirmationCode extrae el HTConfirmationCode limpio desde el uniqueID de Cancel.
// El ID llega como "<OwnReference>|<SupplierReference>[-índice]", ej: "GI20001192|HT-8RPSZBSANDBOX-1".
// Devuelve la parte del proveedor sin el sufijo de índice: "HT-8RPSZBSANDBOX".
func ExtractHTConfirmationCode(rawID string) string {
	providerPart := rawID
	if idx := strings.Index(rawID, "|"); idx >= 0 {
		providerPart = rawID[idx+1:]
	}
	return trailingIndexSuffix.ReplaceAllString(providerPart, "")
}

// GICancelRequestToProvider convierte una petición del dominio a una petición GraphQL
func GICancelRequestToProvider(req *domain.CancelRequest, cfg config.ProviderConfig) *ProviderCancelRQ {

	var htConfirmationCode string

	if len(req.UniqueID) > 0 {
		htConfirmationCode = ExtractHTConfirmationCode(req.UniqueID[0].ID)
	}

	return &ProviderCancelRQ{
		Query: CancelMutation,
		Variables: CancelVariables{
			Cancel: &CancelRequestInput{
				HTConfirmationCode: htConfirmationCode,
			},
		},
	}
}

// ProviderCancelResponseToGI convierte una respuesta GraphQL a una respuesta del dominio
func ProviderCancelResponseToGI(graphqlResp *ProviderCancelRS, req *domain.CancelRequest) *domain.BaseJsonRS[*domain.CancelResponse] {
	cancelResp := graphqlResp.Data.Cancel

	// Construir CancelRules desde la respuesta
	cancelRules := []domain.CancelRule{}

	for _, room := range cancelResp.Rooms {
		// Crear una regla de cancelación por cada habitación cancelada
		if room.Cancelled {
			cancelRule := domain.CancelRule{
				Amount:        roundTo2Decimals(room.CancellationAmount),
				CancelByDate:  normalizeDate(room.CancellationDate),
				CurrencyCode:  room.Currency,
				DecimalPlaces: 2,
				Type:          "Charge",
			}
			cancelRules = append(cancelRules, cancelRule)
		}
	}

	cancelResponse := &domain.CancelResponse{}
	cancelResponse.CancelRules.CancelRule = cancelRules

	// Construir uniqueID con el localizador
	uniqueIDStruct := struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}{
		ID:   cancelResp.HTConfirmationCode,
		Type: "14",
	}

	cancelResponse.UniqueID = uniqueIDStruct

	base := &domain.BaseJsonRS[*domain.CancelResponse]{
		Success:               "true",
		CancelInfoRS:          cancelResponse,
		TransactionIdentifier: cancelResp.HTConfirmationCode,
		Status:                "Cancelled",
	}

	errorMessage := ""
	if sessionCtx := session.FromContext(); sessionCtx != nil {
		if metrics := sessionCtx.Data().SupplierMetrics; metrics != nil {
			if metrics.ErrorMessage != "" {
				errorMessage = metrics.ErrorMessage
			} else if status := metrics.HttpStatusCode; status != 0 && status != 200 {
				errorMessage = "HTTP error: " + strconv.Itoa(status)
			}
		}
	}
	attachCancelErrors(base, errorMessage)

	return base
}

func attachCancelErrors(base *domain.BaseJsonRS[*domain.CancelResponse], errorMessage string) {
	if errorMessage == "" {
		return
	}

	base.Errors = BuildGIErrorContainer(errorMessage)
	base.Success = "false"
	base.Status = ""
	if base.InternalCondition == nil {
		base.InternalCondition = &giresponsecommon.InternalCondition{}
	}
	base.InternalCondition.ProviderStatus = "400"
	base.InternalCondition.ProviderStatusDescription = errorMessage
	base.InternalCondition.Status = "200"
	base.InternalCondition.StatusDescription = "OK"
}
