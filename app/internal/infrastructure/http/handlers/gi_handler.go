package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"ws-int-httr/internal/domain"
	"ws-int-httr/internal/domain/log_domain"

	"ws-int-httr/internal/infrastructure"
	"ws-int-httr/internal/infrastructure/config"
	"ws-int-httr/internal/infrastructure/http/handlers/generic"
	"ws-int-httr/internal/infrastructure/logger"
	"ws-int-httr/internal/infrastructure/registry"
	"ws-int-httr/internal/infrastructure/serializer"
	"ws-int-httr/internal/infrastructure/session"

	"github.com/gin-gonic/gin"
)

const (
	RequestTypeAvail   = "avail"
	RequestTypePreBook = "prebook"
	RequestTypeBook    = "book"
	RequestTypeCancel  = "cancel"
)

type BookingHTTPHandler struct {
	bookingService domain.BookingServicer
	serializer     serializer.Serializer
}

func NewBookingHTTPHandler(bookingService domain.BookingServicer, ser serializer.Serializer) *BookingHTTPHandler {
	return &BookingHTTPHandler{
		bookingService: bookingService,
		serializer:     ser,
	}
}

// getHostname obtiene el hostname del servidor
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// getProviderInfo obtiene la información del proveedor desde la config
func getProviderInfo() (providerID string, providerCode string, integration string) {
	cfg, ok := registry.Get[config.AppConfig]("config")
	if !ok {
		return "0", "UNKNOWN", "ws-int-httr"
	}

	providerID = "0"
	if len(cfg.ProviderIdList()) > 0 {
		providerID = strconv.Itoa(cfg.ProviderIdList()[0])
	}

	providerCode = "PROVIDERCODE"
	integration = "ws-int-httr"

	return providerID, providerCode, integration
}

func (h *BookingHTTPHandler) HandleAvail(c *gin.Context) {
	generic.IncrementCounter(generic.TAG_COUNTER_AVAIL, generic.TAG_COUNTER_TOTAL)

	defer writeLog(RequestTypeAvail)
	defer checkPanic(c, RequestTypeAvail)

	sessionCtx := session.FromContext()

	// CRÍTICO - Inicializar AvailLog
	providerID, providerCode, integration := getProviderInfo()
	availLog := &log_domain.AvailLog{
		ProviderID:     providerID,
		ProviderCode:   providerCode,
		Node:           getHostname(),
		Integration:    integration,
		SentToSupplier: false,
	}

	// Guardar log en sesión para que se complete en el mapper y client
	sessionCtx.Data().AvailLog = availLog

	var giReq domain.AvailRequest

	if err := c.ShouldBindJSON(&giReq); err != nil {
		generic.IncrementCounter(generic.TAG_COUNTER_ERRORS, generic.TAG_COUNTER_AVAIL_ERRORS, "ERRORCODE#400")
		ResponseError(c, &domain.ErrorInvalidJSON, RequestTypeAvail)
		return
	}

	// CRÍTICO - Popular campos del request en el log
	availLog.EchoToken = giReq.InternalCondition.CallCondition.EchoToken
	availLog.RqType = "OTA_HotelAvail"
	availLog.PrimaryLangID = giReq.PrimaryLangID
	availLog.Version = giReq.Version
	availLog.RqTimestamp = giReq.InternalCondition.CallCondition.OriginTimeStamp
	availLog.Market = giReq.AvailRequestSegments.AvailRequestSegment.Tpa.Market
	availLog.Nationality = giReq.AvailRequestSegments.AvailRequestSegment.Tpa.Nationality
	availLog.StayDateRangeStart = giReq.AvailRequestSegments.AvailRequestSegment.StayDateRange.Start
	availLog.StayDateRangeEnd = giReq.AvailRequestSegments.AvailRequestSegment.StayDateRange.End

	// Client info
	if len(giReq.InternalCondition.Channels.Channel) > 0 {
		availLog.BookingChannel = giReq.InternalCondition.Channels.Channel[0].Code
	}
	availLog.ClientCode = giReq.InternalCondition.ClientCondition.Code
	availLog.RequestorID = giReq.Pos.Source.RequestorID.ID

	// Hotel codes solicitados
	var rqHotelCodeList []string
	for _, hotelRef := range giReq.AvailRequestSegments.AvailRequestSegment.HotelSearchCriteria.Criterion.HotelRef {
		if hotelRef.HotelCode != "" {
			rqHotelCodeList = append(rqHotelCodeList, hotelRef.HotelCode)
		}
		if hotelRef.HotelCityCode != "" {
			availLog.RqCity = hotelRef.HotelCityCode
		}
		if hotelRef.AreaID != "" {
			availLog.RqZone = hotelRef.AreaID
		}
	}
	availLog.RqHotelCodeList = rqHotelCodeList

	// Calcular distribución y número de habitaciones/huéspedes
	numRooms := len(giReq.AvailRequestSegments.AvailRequestSegment.RoomStayCandidates.RoomStayCandidateList)
	availLog.RqNumRooms = numRooms

	var distributionParts []string
	totalGuests := 0
	for _, candidate := range giReq.AvailRequestSegments.AvailRequestSegment.RoomStayCandidates.RoomStayCandidateList {
		adults := 0
		children := 0
		infants := 0
		for _, guest := range candidate.GuestCounts.GuestCountList {
			totalGuests++
			if guest.Age >= 18 {
				adults++
			} else if guest.Age >= 2 {
				children++
			} else {
				infants++
			}
		}
		distributionParts = append(distributionParts, fmt.Sprintf("%d-%d-%d", adults, children, infants))
	}
	availLog.RqNumGuests = totalGuests
	availLog.RqDistribution = strings.Join(distributionParts, ",")

	// En avail, rqInternal solo se guarda con debug activo
	if strings.EqualFold(giReq.Debug, infrastructure.DEBUG_PASSWORD) {
		rqInternalBytes, _ := json.Marshal(giReq)
		availLog.RqInternal = string(rqInternalBytes)
	}

	// Guardar echoToken en sesión para uso en otros componentes
	sessionCtx.Set("echoToken", availLog.EchoToken)

	// Ejecutar servicio
	giResp, err := h.bookingService.Availability(&giReq)

	if err != nil {
		customErr := domain.ErrorAvailCode
		customErr.Err = err

		if errors.Is(err, domain.ErrMissingDates) {
			customErr.ErrorCode = "400"
		}

		generic.IncrementCounter(generic.TAG_COUNTER_ERRORS, generic.TAG_COUNTER_AVAIL_ERRORS, "ERRORCODE#"+customErr.ErrorCode)

		// Completar y escribir el log de error antes de retornar
		addAvailErrorLog(err)

		ResponseError(c, &customErr, RequestTypeAvail)
		return
	}

	// En avail no se guarda rsInternal (alineado con OpenTours)

	// Completar métricas de resultado
	availLog.Success = "OK"
	availLog.Summary = "OK"
	availLog.Error = "false"
	availLog.ErrorCode = 0

	// El log se completa y escribe en completeAvailLog() dentro del mapper
	generic.IncrementCounter(generic.TAG_COUNTER_AVAIL_OK)
	c.JSON(http.StatusOK, giResp)
}

func (h *BookingHTTPHandler) HandlePreBook(c *gin.Context) {
	generic.IncrementCounter(generic.TAG_COUNTER_PREBOOK, generic.TAG_COUNTER_TOTAL)

	defer writeLog(RequestTypePreBook)
	defer checkPanic(c, RequestTypePreBook)

	sessionCtx := session.FromContext()

	// CRÍTICO - Inicializar PreBookLog
	providerID, providerCode, integration := getProviderInfo()
	preBookLog := &log_domain.HotelResBookLog{
		ProviderID:     providerID,
		ProviderCode:   providerCode,
		Node:           getHostname(),
		Integration:    integration,
		SentToSupplier: false,
	}

	sessionCtx.Set("preBookLog", preBookLog)

	var giReq domain.PreBookRequest

	if err := c.ShouldBindJSON(&giReq); err != nil {
		generic.IncrementCounter(generic.TAG_COUNTER_ERRORS, generic.TAG_COUNTER_PREBOOK_ERRORS, "ERRORCODE#400")
		ResponseError(c, &domain.ErrorInvalidJSON, RequestTypePreBook)
		return
	}

	// Popular campos del request
	preBookLog.EchoToken = giReq.EchoToken
	preBookLog.RqType = giReq.RqType
	preBookLog.PrimaryLangID = giReq.PrimaryLangID
	preBookLog.Version = giReq.Version
	preBookLog.ResStatus = giReq.ResStatus
	preBookLog.RqTimestamp = time.Now().UnixMilli()
	preBookLog.RequestorID = giReq.Pos.Source.RequestorID.ID

	if len(giReq.InternalCondition.Channels.Channel) > 0 {
		preBookLog.BookingChannel = giReq.InternalCondition.Channels.Channel[0].Code
	}
	preBookLog.ClientCode = giReq.InternalCondition.ClientCondition.Code

	// Extraer booking codes del request
	var rqBookingCodes []string
	for _, hotelRes := range giReq.HotelReservations.HotelReservation {
		for _, roomStay := range hotelRes.RoomStays.RoomStay {
			bookingCodeValue := roomStay.RoomRates.RoomRate.BookingCode
			if bookingCodeValue != "" {
				rqBookingCodes = append(rqBookingCodes, bookingCodeValue)
			}
		}
	}
	preBookLog.RqBookingCode = rqBookingCodes

	// Guardar rqInternal
	rqInternalBytes, _ := json.Marshal(giReq)
	preBookLog.RqInternal = string(rqInternalBytes)

	sessionCtx.Set("echoToken", preBookLog.EchoToken)

	giResp, err := h.bookingService.PreBook(&giReq)
	if err != nil {
		customErr := domain.ErrorPrebookCode
		customErr.Err = err

		// Completar y escribir el log de error antes de retornar
		addPreBookErrorLog(err)
		generic.IncrementCounter(generic.TAG_COUNTER_ERRORS, generic.TAG_COUNTER_PREBOOK_ERRORS, "ERRORCODE#"+customErr.ErrorCode)

		ResponseError(c, &customErr, RequestTypePreBook)
		return
	}

	// Guardar rsInternal
	rsInternalBytes, _ := json.Marshal(giResp)
	preBookLog.RsInternal = string(rsInternalBytes)

	// Completar log
	preBookLog.Success = "OK"
	preBookLog.Error = "false"
	preBookLog.ErrorCode = 0

	generic.IncrementCounter(generic.TAG_COUNTER_PREBOOK_OK)
	c.JSON(http.StatusOK, giResp)
}

func (h *BookingHTTPHandler) HandleBook(c *gin.Context) {
	generic.IncrementCounter(generic.TAG_COUNTER_BOOK, generic.TAG_COUNTER_TOTAL)

	defer writeLog(RequestTypeBook)
	defer checkPanic(c, RequestTypeBook)

	sessionCtx := session.FromContext()

	// CRÍTICO - Inicializar BookLog (HotelResCommitLog)
	providerID, providerCode, integration := getProviderInfo()
	bookLog := &log_domain.HotelResCommitLog{
		ProviderID:     providerID,
		ProviderCode:   providerCode,
		Node:           getHostname(),
		Integration:    integration,
		SentToSupplier: false,
	}

	sessionCtx.Set("bookLog", bookLog)

	var giReq domain.BookRequest

	if err := c.ShouldBindJSON(&giReq); err != nil {
		generic.IncrementCounter(generic.TAG_COUNTER_ERRORS, generic.TAG_COUNTER_BOOK_ERRORS, "ERRORCODE#400")
		ResponseError(c, &domain.ErrorInvalidJSON, RequestTypeBook)
		return
	}

	// Popular campos del request
	bookLog.EchoToken = giReq.EchoToken
	bookLog.RqType = giReq.RqType
	bookLog.PrimaryLangID = giReq.PrimaryLangID
	bookLog.Version = giReq.Version
	bookLog.ResStatus = giReq.ResStatus
	bookLog.RqTimestamp = time.Now().UnixMilli()
	bookLog.RequestorID = giReq.Pos.Source.RequestorID.ID

	if len(giReq.InternalCondition.Channels.Channel) > 0 {
		bookLog.BookingChannel = giReq.InternalCondition.Channels.Channel[0].Code
	}
	bookLog.ClientCode = giReq.InternalCondition.ClientCondition.Code

	// Extraer booking codes del request
	var rqBookingCodes []string
	for _, hotelRes := range giReq.HotelReservations.HotelReservation {
		for _, roomStay := range hotelRes.RoomStays.RoomStay {
			bookingCodeValue := roomStay.RoomRates.RoomRate.BookingCode
			if bookingCodeValue != "" {
				rqBookingCodes = append(rqBookingCodes, bookingCodeValue)
			}
		}
	}
	bookLog.RqBookingCode = rqBookingCodes

	// Guardar rqInternal
	rqInternalBytes, _ := json.Marshal(giReq)
	bookLog.RqInternal = string(rqInternalBytes)

	sessionCtx.Set("echoToken", bookLog.EchoToken)

	giResp, err := h.bookingService.Book(&giReq)
	if err != nil {
		customErr := domain.ErrorBookCode
		customErr.Err = err

		// Completar y escribir el log de error antes de retornar
		addBookErrorLog(err)
		generic.IncrementCounter(generic.TAG_COUNTER_ERRORS, generic.TAG_COUNTER_BOOK_ERRORS, "ERRORCODE#"+customErr.ErrorCode)

		ResponseError(c, &customErr, RequestTypeBook)
		return
	}

	// Guardar rsInternal
	rsInternalBytes, _ := json.Marshal(giResp)
	bookLog.RsInternal = string(rsInternalBytes)

	// Completar log
	bookLog.Success = "OK"
	bookLog.Error = "false"
	bookLog.ErrorCode = 0

	// Si hay un bookResponse en la sesión (BaseJsonRSBook con InternalConditionBook), usarlo
	if bookResponseInterface, ok := session.FromContext().Get("bookResponse"); ok {
		if bookResponse, ok := bookResponseInterface.(*domain.BaseJsonRSBook); ok {
			generic.IncrementCounter(generic.TAG_COUNTER_BOOK_OK)
			c.JSON(http.StatusOK, bookResponse)
			return
		}
	}

	generic.IncrementCounter(generic.TAG_COUNTER_BOOK_OK)
	c.JSON(http.StatusOK, giResp)
}

func (h *BookingHTTPHandler) HandleCancel(c *gin.Context) {
	generic.IncrementCounter(generic.TAG_COUNTER_CANCEL, generic.TAG_COUNTER_TOTAL)

	defer writeLog(RequestTypeCancel)
	defer checkPanic(c, RequestTypeCancel)

	sessionCtx := session.FromContext()

	// CRÍTICO - Inicializar CancelLog
	providerID, providerCode, integration := getProviderInfo()
	cancelLog := &log_domain.CancelLog{
		ProviderID:     providerID,
		ProviderCode:   providerCode,
		Integration:    integration,
		SentToSupplier: false,
	}

	sessionCtx.Set("cancelLog", cancelLog)

	var giReq domain.CancelRequest

	if err := c.ShouldBindJSON(&giReq); err != nil {
		generic.IncrementCounter(generic.TAG_COUNTER_ERRORS, generic.TAG_COUNTER_CANCEL_ERRORS, "ERRORCODE#400")
		ResponseError(c, &domain.ErrorInvalidJSON, RequestTypeCancel)
		return
	}

	// Popular campos del request
	echoToken := giReq.InternalCondition.CallCondition.EchoToken
	cancelLog.RqType = giReq.RqType
	cancelLog.PrimaryLangID = giReq.PrimaryLangID
	cancelLog.Version = strconv.FormatFloat(giReq.Version, 'f', -1, 64)
	cancelLog.RqTimestamp = time.Now().UnixMilli()
	cancelLog.RequestorID = giReq.Pos.Source.RequestorID.ID
	cancelLog.ClientCode = giReq.InternalCondition.ClientCondition.Code
	cancelLog.CancelType = giReq.CancelType
	cancelLog.IsRebook = giReq.InternalCondition.Rebook

	// Extraer transaction identifier
	if len(giReq.UniqueID) > 0 {
		cancelLog.TransactionIdentifier = giReq.UniqueID[0].ID
	}

	// Guardar rqInternal
	rqInternalBytes, _ := json.Marshal(giReq)
	cancelLog.RqInternal = string(rqInternalBytes)

	sessionCtx.Set("echoToken", echoToken)

	giResp, err := h.bookingService.Cancel(&giReq)
	if err != nil {
		customErr := domain.ErrorCancelCode
		customErr.Err = err

		// Completar y escribir el log de error antes de retornar
		addCancelErrorLog(err)
		generic.IncrementCounter(generic.TAG_COUNTER_ERRORS, generic.TAG_COUNTER_CANCEL_ERRORS, "ERRORCODE#"+customErr.ErrorCode)

		ResponseError(c, &customErr, RequestTypeCancel)
		return
	}

	// Guardar rsInternal
	rsInternalBytes, _ := json.Marshal(giResp)
	cancelLog.RsInternal = string(rsInternalBytes)

	// Completar log
	cancelLog.Success = "OK"
	cancelLog.Error = "false"
	cancelLog.ErrorCode = 0
	cancelLog.Status = "200"
	if metrics := sessionCtx.Data().SupplierMetrics; metrics != nil {
		cancelLog.SupplierRsTime = metrics.RsTime
		cancelLog.SupplierRsHttpStatusCode = metrics.HttpStatusCode
		cancelLog.SupplierRsLength = metrics.RsLength
		cancelLog.SupplierErrorMessage = metrics.ErrorMessage
	}

	generic.IncrementCounter(generic.TAG_COUNTER_CANCEL_OK)
	c.JSON(http.StatusOK, giResp)
}

type WebServiceContext struct {
	RequestRawData []byte
	RequestType    string
	AvailLog       *log_domain.AvailLog
	PreBookLog     *log_domain.HotelResBookLog
	ConfirmLog     *log_domain.HotelResCommitLog
	CancelLog      *log_domain.CancelLog
}

const (
	ResStatusCommit = "Commit"
	ResStatusBook   = "Book"
)

// RequestTypeInfo es un struct mínimo para identificar el tipo de request
// sin tener que deserializar completamente el JSON
type RequestTypeInfo struct {
	RqType    string `json:"rqType"`
	ResStatus string `json:"resStatus"`
}

func (h *BookingHTTPHandler) HandleWebServiceEndpoint(c *gin.Context) {
	requestRawData, ok := session.FromContext().Get("requestRawData")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "RequestRawData not found"})
		return
	}

	rawData, ok := requestRawData.([]byte)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid RequestRawData type"})
		return
	}

	// Deserializar solo los campos necesarios para identificar el tipo de request
	var reqInfo RequestTypeInfo
	if err := json.Unmarshal(rawData, &reqInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Determinar el tipo de request basado en rqType y resStatus
	switch {
	case strings.Contains(reqInfo.RqType, "GIOTAHotelAvailRQ"):
		session.FromContext().Set("requestType", "GIOTAHotelAvailRQ")
		h.HandleAvail(c)
	case strings.Contains(reqInfo.RqType, "GIOTAHotelResRQ"):
		// Para determinar si es Book o Commit, usamos resStatus
		if reqInfo.ResStatus == "Book" || reqInfo.ResStatus == "book" {
			session.FromContext().Set("requestType", "GIOTAHotelResRQ_Book")
			h.HandlePreBook(c)
		} else if reqInfo.ResStatus == "Commit" || reqInfo.ResStatus == "commit" {
			session.FromContext().Set("requestType", "GIOTAHotelResRQ_COMMIT")
			h.HandleBook(c)
		} else {
			// Por defecto, intentamos PreBook
			session.FromContext().Set("requestType", "GIOTAHotelResRQ_Book")
			h.HandlePreBook(c)
		}
	case strings.Contains(reqInfo.RqType, "GIOTACancelRQ"):
		session.FromContext().Set("requestType", "GIOTACancelRQ")
		h.HandleCancel(c)
	default:
		generic.IncrementCounter(generic.TAG_COUNTER_ERRORS, "ERRORCODE#400")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error petición desconocida."})
	}
}

// FUNCIONES DEFER
func writeLog(requestType string) {
	sessionCtx := session.FromContext()

	var log log_domain.GenericCallLog

	switch requestType {
	case RequestTypeAvail:
		log = sessionCtx.Data().AvailLog
	case RequestTypePreBook:
		log = sessionCtx.Data().PreBookLog
	case RequestTypeBook:
		log = sessionCtx.Data().BookLog
	case RequestTypeCancel:
		log = sessionCtx.Data().CancelLog
	}

	// Calcular tiempo total
	if log != nil && !sessionCtx.Data().StartTime.IsZero() {
		log.SetRsTime(int(time.Since(sessionCtx.Data().StartTime).Milliseconds()))
	}

	// Escribir log siempre (con o sin panic)
	if log != nil {
		if structuredLogger, ok := registry.Get[logger.StructuredLogger]("structuredLogger"); ok {
			structuredLogger.LogCall(log)
		}
	}
}

func checkPanic(c *gin.Context, requestType string) {
	if r := recover(); r != nil {
		sessionCtx := session.FromContext()

		// ✅ Intentar convertir a CustomError
		var customErr domain.CustomError
		if ce, ok := r.(domain.CustomError); ok {
			customErr = ce // Ya es un CustomError con código y mensaje específicos
		} else {
			// Fallback: panic con valor no CustomError
			customErr = domain.ErrorPanicInternal
			customErr.Err = fmt.Errorf("%v", r)
		}

		// Guardar en sesión para otros usos
		if sessionCtx != nil {
			sessionCtx.Data().PanicValue = customErr
		}

		// Actualizar log con información del panic
		updateLogWithPanic(requestType, customErr)

		ResponseError(c, &customErr, requestType)
		return
	}
}

// updateLogWithPanic actualiza el log correspondiente con información del panic
func updateLogWithPanic(requestType string, customErr domain.CustomError) {
	sessionCtx := session.FromContext()
	if sessionCtx == nil {
		return
	}

	errorCode := 500
	if code, err := strconv.Atoi(customErr.ErrorCode); err == nil {
		errorCode = code
	}

	errorMessage := fmt.Sprintf("PANIC: %s", customErr.Message)
	if customErr.Err != nil {
		errorMessage += fmt.Sprintf(" - %v", customErr.Err)
	}

	// Asegurar métricas de proveedor también en escenarios de panic
	supplierRsTime := int64(0)
	supplierRsHttpStatusCode := 0
	supplierRsLength := 0
	supplierErrorMessage := errorMessage
	if metrics := sessionCtx.Data().SupplierMetrics; metrics != nil {
		supplierRsTime = metrics.RsTime
		supplierRsHttpStatusCode = metrics.HttpStatusCode
		supplierRsLength = metrics.RsLength
		if metrics.ErrorMessage != "" {
			supplierErrorMessage = metrics.ErrorMessage
		}
	}

	switch requestType {
	case RequestTypeAvail:
		if sessionCtx.Data().AvailLog != nil {
			sessionCtx.Data().AvailLog.Success = "KO"
			sessionCtx.Data().AvailLog.Error = "true"
			sessionCtx.Data().AvailLog.ErrorCode = errorCode
			sessionCtx.Data().AvailLog.ErrorMessage = errorMessage
			sessionCtx.Data().AvailLog.SupplierRsTime = supplierRsTime
			sessionCtx.Data().AvailLog.SupplierRsHttpStatusCode = supplierRsHttpStatusCode
			sessionCtx.Data().AvailLog.SupplierRsLength = supplierRsLength
			sessionCtx.Data().AvailLog.SupplierErrorMessage = supplierErrorMessage
		}
	case RequestTypePreBook:
		if sessionCtx.Data().PreBookLog != nil {
			sessionCtx.Data().PreBookLog.Success = "KO"
			sessionCtx.Data().PreBookLog.Error = "true"
			sessionCtx.Data().PreBookLog.ErrorCode = errorCode
			sessionCtx.Data().PreBookLog.ErrorMessage = errorMessage
			sessionCtx.Data().PreBookLog.SupplierRsTime = supplierRsTime
			sessionCtx.Data().PreBookLog.SupplierRsHttpStatusCode = supplierRsHttpStatusCode
			sessionCtx.Data().PreBookLog.SupplierRsLength = supplierRsLength
			sessionCtx.Data().PreBookLog.SupplierErrorMessage = supplierErrorMessage
		}
	case RequestTypeBook:
		if sessionCtx.Data().BookLog != nil {
			sessionCtx.Data().BookLog.Success = "KO"
			sessionCtx.Data().BookLog.Error = "true"
			sessionCtx.Data().BookLog.ErrorCode = errorCode
			sessionCtx.Data().BookLog.ErrorMessage = errorMessage
			sessionCtx.Data().BookLog.SupplierRsTime = supplierRsTime
			sessionCtx.Data().BookLog.SupplierRsHttpStatusCode = supplierRsHttpStatusCode
			sessionCtx.Data().BookLog.SupplierRsLength = supplierRsLength
			sessionCtx.Data().BookLog.SupplierErrorMessage = supplierErrorMessage
		}
	case RequestTypeCancel:
		if sessionCtx.Data().CancelLog != nil {
			sessionCtx.Data().CancelLog.Success = "KO"
			sessionCtx.Data().CancelLog.Error = "true"
			sessionCtx.Data().CancelLog.ErrorCode = errorCode
			sessionCtx.Data().CancelLog.ErrorMessage = errorMessage
			sessionCtx.Data().CancelLog.SupplierRsTime = supplierRsTime
			sessionCtx.Data().CancelLog.SupplierRsHttpStatusCode = supplierRsHttpStatusCode
			sessionCtx.Data().CancelLog.SupplierRsLength = supplierRsLength
			sessionCtx.Data().CancelLog.SupplierErrorMessage = supplierErrorMessage
		}
	}
}

// addPreBookErrorLog completa el log de PreBook cuando hay un error
func addPreBookErrorLog(err error) {
	sessionCtx := session.FromContext()
	if sessionCtx == nil {
		return
	}

	preBookLog := sessionCtx.Data().PreBookLog
	if preBookLog == nil {
		return
	}

	// Obtener métricas del error de la sesión
	var supplierRsTime int64 = 0
	var supplierRsHttpStatusCode int = 0
	var supplierRsLength int = 0
	supplierErrorMessage := err.Error()

	if metrics := sessionCtx.Data().SupplierMetrics; metrics != nil {
		supplierRsTime = metrics.RsTime
		supplierRsHttpStatusCode = metrics.HttpStatusCode
		supplierRsLength = metrics.RsLength
		if metrics.ErrorMessage != "" {
			supplierErrorMessage = metrics.ErrorMessage
		}
	}

	// Calcular tiempo total desde el inicio
	var totalTime int = 0
	if !sessionCtx.Data().StartTime.IsZero() {
		totalTime = int(time.Since(sessionCtx.Data().StartTime).Milliseconds())
	}

	// Completar el log con información de error
	preBookLog.Success = "KO"
	preBookLog.Error = "true"
	preBookLog.ErrorCode = 1
	preBookLog.ErrorMessage = supplierErrorMessage
	preBookLog.InternalMessage = ""
	preBookLog.TotalTime = totalTime
	preBookLog.SupplierRsTime = supplierRsTime
	preBookLog.SupplierRsHttpStatusCode = supplierRsHttpStatusCode
	preBookLog.SupplierRsLength = supplierRsLength
	preBookLog.SupplierErrorMessage = supplierErrorMessage
	preBookLog.ResResponseType = "Pending"
	preBookLog.ResStatus = "500"
	preBookLog.SentToSupplier = true

	// RsInternal: JSON de respuesta vacío o error
	preBookLog.RsInternal = ""

}

// addAvailErrorLog completa el log de Avail cuando hay un error
func addAvailErrorLog(err error) {
	sessionCtx := session.FromContext()
	if sessionCtx == nil {
		return
	}

	availLog := sessionCtx.Data().AvailLog
	if availLog == nil {
		return
	}

	// Obtener métricas del error de la sesión
	var supplierRsTime int64 = 0
	var supplierRsHttpStatusCode int = 0
	var supplierRsLength int = 0
	supplierErrorMessage := err.Error()

	if metrics := sessionCtx.Data().SupplierMetrics; metrics != nil {
		supplierRsTime = metrics.RsTime
		supplierRsHttpStatusCode = metrics.HttpStatusCode
		supplierRsLength = metrics.RsLength
		if metrics.ErrorMessage != "" {
			supplierErrorMessage = metrics.ErrorMessage
		}
	}

	// Calcular tiempo total desde el inicio
	var rsTime int = 0
	if !sessionCtx.Data().StartTime.IsZero() {
		rsTime = int(time.Since(sessionCtx.Data().StartTime).Milliseconds())
	}

	// Completar el log con información de error
	availLog.Success = "KO"
	availLog.Error = "false"
	availLog.ErrorCode = 3
	availLog.ErrorMessage = "Error on avail"
	availLog.InternalMessage = ""
	availLog.RsTime = rsTime
	availLog.SupplierRsTime = supplierRsTime
	availLog.SupplierRsHttpStatusCode = supplierRsHttpStatusCode
	availLog.SupplierRsLength = supplierRsLength
	availLog.SupplierErrorMessage = supplierErrorMessage
	availLog.Summary = "KO"
	availLog.SentToSupplier = true
	availLog.RsNumHotels = 0
	availLog.RsNumRoomStay = 0
	availLog.SupplierNumHotels = 0
	availLog.SupplierNumRooms = 0
	availLog.SupplierNumRates = 0
	availLog.RsLength = 0
	availLog.CachedProviderResponse = false

}

// addBookErrorLog completa el log de Book cuando hay un error
func addBookErrorLog(err error) {
	sessionCtx := session.FromContext()
	if sessionCtx == nil {
		return
	}

	bookLog := sessionCtx.Data().BookLog
	if bookLog == nil {
		return
	}

	// Obtener métricas del error de la sesión
	var supplierRsTime int64 = 0
	var supplierRsHttpStatusCode int = 0
	var supplierRsLength int = 0
	supplierErrorMessage := err.Error()

	if metrics := sessionCtx.Data().SupplierMetrics; metrics != nil {
		supplierRsTime = metrics.RsTime
		supplierRsHttpStatusCode = metrics.HttpStatusCode
		supplierRsLength = metrics.RsLength
		if metrics.ErrorMessage != "" {
			supplierErrorMessage = metrics.ErrorMessage
		}
	}

	// Calcular tiempo total desde el inicio
	var totalTime int = 0
	if !sessionCtx.Data().StartTime.IsZero() {
		totalTime = int(time.Since(sessionCtx.Data().StartTime).Milliseconds())
	}

	// Completar el log con información de error
	bookLog.Success = "KO"
	bookLog.Error = "true"
	bookLog.ErrorCode = 1
	bookLog.ErrorMessage = supplierErrorMessage
	bookLog.InternalMessage = ""
	bookLog.TotalTime = totalTime
	bookLog.SupplierRsTime = supplierRsTime
	bookLog.SupplierRsHttpStatusCode = supplierRsHttpStatusCode
	bookLog.SupplierRsLength = supplierRsLength
	bookLog.SupplierErrorMessage = supplierErrorMessage
	bookLog.ResResponseType = "Pending"
	bookLog.ResStatus = "500"
	bookLog.SentToSupplier = true

	// RsInternal: JSON de respuesta vacío o error
	bookLog.RsInternal = ""

}

// addCancelErrorLog completa el log de Cancel cuando hay un error
func addCancelErrorLog(err error) {
	sessionCtx := session.FromContext()
	if sessionCtx == nil {
		return
	}

	cancelLog := sessionCtx.Data().CancelLog
	if cancelLog == nil {
		return
	}

	// Obtener métricas del error de la sesión
	var supplierRsTime int64 = 0
	var supplierRsHttpStatusCode int = 0
	var supplierRsLength int = 0
	supplierErrorMessage := err.Error()

	if metrics := sessionCtx.Data().SupplierMetrics; metrics != nil {
		supplierRsTime = metrics.RsTime
		supplierRsHttpStatusCode = metrics.HttpStatusCode
		supplierRsLength = metrics.RsLength
		if metrics.ErrorMessage != "" {
			supplierErrorMessage = metrics.ErrorMessage
		}
	}

	// Calcular tiempo total desde el inicio
	var totalTime int = 0
	if !sessionCtx.Data().StartTime.IsZero() {
		totalTime = int(time.Since(sessionCtx.Data().StartTime).Milliseconds())
	}

	// Completar el log con información de error
	cancelLog.Success = "KO"
	cancelLog.Error = "true"
	cancelLog.ErrorCode = 1
	cancelLog.ErrorMessage = supplierErrorMessage
	cancelLog.InternalMessage = ""
	cancelLog.TotalTime = totalTime
	cancelLog.SupplierRsTime = supplierRsTime
	cancelLog.SupplierRsHttpStatusCode = supplierRsHttpStatusCode
	cancelLog.SupplierRsLength = supplierRsLength
	cancelLog.SupplierErrorMessage = supplierErrorMessage
	cancelLog.Status = "500"
	cancelLog.SentToSupplier = true

	// RsInternal: JSON de respuesta vacío o error
	cancelLog.RsInternal = ""
}
