package handlers

import (
	"encoding/json"
	"net/http"
	"ws-int-httr/internal/domain"
	giresponsecommon "ws-int-httr/internal/domain/gi_response_common"
	hoteltrader "ws-int-httr/internal/infrastructure/mapping/provider"

	"github.com/gin-gonic/gin"
)

// ResponseError maneja errores y devuelve una respuesta estructurada al cliente
// Basado en el patrón del otro conector en producción
func ResponseError(c *gin.Context, ce *domain.CustomError, requestType string) {
	// Construir respuesta de error estructurada según el tipo de request
	var responseData []byte
	var err error

	switch requestType {
	case "avail":
		errorResp := createAvailErrorResponse(ce)
		responseData, err = json.Marshal(errorResp)
	case "prebook":
		errorResp := createPreBookErrorResponse(ce)
		responseData, err = json.Marshal(errorResp)
	case "book":
		errorResp := createBookErrorResponse(ce)
		responseData, err = json.Marshal(errorResp)
	case "cancel":
		errorResp := createCancelErrorResponse(ce)
		responseData, err = json.Marshal(errorResp)
	default:
		// Error genérico
		c.JSON(http.StatusOK, gin.H{
			"errors": hoteltrader.BuildGIErrorContainer(ce.Message),
			"internalConditionRS": giresponsecommon.InternalCondition{
				Status:                    "200",
				StatusDescription:         "OK",
				ProviderStatus:            "400",
				ProviderStatusDescription: getErrorDescription(ce),
			},
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar respuesta de error"})
		return
	}

	// Responder con HTTP 200 pero con error estructurado (como hace el otro conector)
	c.Data(http.StatusOK, "application/json", responseData)
}

func createAvailErrorResponse(ce *domain.CustomError) domain.BaseJsonRS[domain.AvailResponse] {
	return domain.BaseJsonRS[domain.AvailResponse]{
		Errors: hoteltrader.BuildGIErrorContainer(ce.Message),
		InternalCondition: &giresponsecommon.InternalCondition{
			Status:                    "200",
			StatusDescription:         "OK",
			ProviderStatus:            "400",
			ProviderStatusDescription: getErrorDescription(ce),
		},
		Success: "false",
		GiRoomStays: domain.AvailResponse{
			GiRoomStayGroup: []domain.GiRoomStayGroup{},
		},
	}
}

func createPreBookErrorResponse(ce *domain.CustomError) domain.BaseJsonRS[domain.PreBookResponse] {
	return domain.BaseJsonRS[domain.PreBookResponse]{
		Errors: hoteltrader.BuildGIErrorContainer(ce.Message),
		InternalCondition: &giresponsecommon.InternalCondition{
			Status:                    "200",
			StatusDescription:         "OK",
			ProviderStatus:            "400",
			ProviderStatusDescription: getErrorDescription(ce),
		},
		Success: "false",
		HotelReservations: domain.PreBookResponse{
			HotelReservation: []domain.HotelReservation{},
		},
	}
}

func createBookErrorResponse(ce *domain.CustomError) domain.BaseJsonRSBook {
	providerStatus := "400"
	providerStatusDesc := getErrorDescription(ce)

	return domain.BaseJsonRSBook{
		Errors: hoteltrader.BuildGIErrorContainer(ce.Message),
		InternalCondition: &giresponsecommon.InternalConditionBook{
			Status:                    "200",
			StatusDescription:         "OK",
			ProviderStatus:            &providerStatus,
			ProviderStatusDescription: &providerStatusDesc,
		},
		Success: "false",
		HotelReservations: domain.BookResponse{
			HotelReservation: []domain.BookHotelReservation{},
		},
	}
}

func createCancelErrorResponse(ce *domain.CustomError) domain.BaseJsonRS[domain.CancelResponse] {
	return domain.BaseJsonRS[domain.CancelResponse]{
		Errors: hoteltrader.BuildGIErrorContainer(ce.Message),
		InternalCondition: &giresponsecommon.InternalCondition{
			Status:                    "200",
			StatusDescription:         "OK",
			ProviderStatus:            "400",
			ProviderStatusDescription: getErrorDescription(ce),
		},
		Success: "false",
		CancelInfoRS: domain.CancelResponse{
			CancelRules: struct {
				CancelRule []domain.CancelRule `json:"cancelRule"`
			}{
				CancelRule: []domain.CancelRule{},
			},
		},
	}
}

func getErrorDescription(ce *domain.CustomError) string {
	if ce.Err != nil {
		return ce.Err.Error()
	}
	return ce.Message
}
