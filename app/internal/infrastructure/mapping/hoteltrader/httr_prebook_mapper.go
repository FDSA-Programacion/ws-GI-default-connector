package hoteltrader

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"ws-int-httr/internal/domain"
	giresponsecommon "ws-int-httr/internal/domain/gi_response_common"
	log_domain "ws-int-httr/internal/domain/log_domain"
	bookingcode "ws-int-httr/internal/infrastructure/booking_code"
	"ws-int-httr/internal/infrastructure/config"
	"ws-int-httr/internal/infrastructure/persistence"
	"ws-int-httr/internal/infrastructure/registry"
	"ws-int-httr/internal/infrastructure/session"
)

func GIPrebookRequestToProvider(domainReq *domain.PreBookRequest, cfg config.ProviderConfig) ProviderPrebookRQ {
	sessionCtx := session.FromContext()
	contextData := map[string]interface{}{}
	requestContext := map[string]interface{}{}
	contextData["request"] = requestContext

	// GI
	giHotelReservations := domainReq.HotelReservations.HotelReservation

	// Inicializar PreBookLog
	hostname, _ := os.Hostname()
	providerId := domainReq.InternalCondition.Channels.Channel[0].ID
	providerCode := domainReq.InternalCondition.Channels.Channel[0].Code
	preBookLog := &log_domain.HotelResBookLog{
		ProviderID:     strconv.Itoa(providerId),
		ProviderCode:   providerCode,
		Integration:    "WS-INT-HTTR",
		Node:           hostname,
		EchoToken:      domainReq.InternalCondition.CallCondition.EchoToken,
		RqType:         domainReq.RqType,
		RequestorID:    domainReq.Pos.Source.RequestorID.ID,
		RqTimestamp:    int64(domainReq.InternalCondition.CallCondition.OriginTimeStamp),
		PrimaryLangID:  domainReq.PrimaryLangID,
		Version:        domainReq.Version,
		ClientName:     domainReq.Pos.Source.RequestorID.CompanyName.CompanyShortName,
		ClientCode:     domainReq.InternalCondition.ClientCondition.Code,
		BookingChannel: domainReq.Pos.Source.BookingChannel.Code,
		RqInternal:     "",
		RqProvider:     "",
		ResStatus:      domainReq.ResStatus,
		SentToSupplier: false,
		IsRebook:       domainReq.InternalCondition.Rebook,
		RqBookingCode:  []string{},
		Success:        "OK",
	}
	// Guardar siempre RqInternal en prebook
	if rqBytes, err := json.Marshal(domainReq); err == nil {
		preBookLog.RqInternal = string(rqBytes)
	}

	// Guardar debug en sesión para usar en la respuesta
	sessionCtx.Data().Debug = domainReq.Debug

	// OT
	providerRQ := ProviderPrebookRQ{}

	// Obtener credenciales según el canal (OPENB2B o OPENB2C)
	// providerUsername := cfg.ProviderUsernameForChannel(providerCode)
	// providerPassword := cfg.ProviderPasswordForChannel(providerCode)

	// Extraer datos de bookingCodes para el log
	var bookingCodes []*bookingcode.InternalBookingCode
	var allQuotes []QuoteRequestInput

	for _, giHotelReservation := range giHotelReservations {
		for _, giRoomStay := range giHotelReservation.RoomStays.RoomStay {
			bookingCode := &bookingcode.InternalBookingCode{}
			bookingCode.Deserialize(giRoomStay.RoomRates.RoomRate.OpenBookingCode)
			bookingCodes = append(bookingCodes, bookingCode)
			preBookLog.RqBookingCode = append(preBookLog.RqBookingCode, giRoomStay.RoomRates.RoomRate.OpenBookingCode)

			// Parsear el quote de extraParams (guardado en base64 en Avail)
			extraParams := bookingcode.StringToMap(bookingCode.ExtraParams)
			quoteJSON := bookingcode.DecodeQuoteFromExtraParams(extraParams["quote"])

			if quoteJSON != "" {
				var quoteData map[string]interface{}
				if err := json.Unmarshal([]byte(quoteJSON), &quoteData); err != nil {
					continue
				}

				// Construir QuoteRequestInput
				quoteInput := QuoteRequestInput{}

				// HTIdentifier
				if htId, ok := quoteData["htIdentifier"].(string); ok {
					quoteInput.HTIdentifier = htId
				}

				// Occupancy
				if occupancyMap, ok := quoteData["occupancy"].(map[string]interface{}); ok {
					if guestAges, ok := occupancyMap["guestAges"].(string); ok {
						quoteInput.Occupancy = &QuoteOccupancyInput{
							GuestAges: guestAges,
						}
					}
				}

				// Rates
				if ratesMap, ok := quoteData["rates"].(map[string]interface{}); ok {
					rates := &QuoteRatesInput{}

					if netPrice, ok := ratesMap["netPrice"].(float64); ok {
						rates.NetPrice = netPrice
					}
					if tax, ok := ratesMap["tax"].(float64); ok {
						rates.Tax = tax
					}
					if grossPrice, ok := ratesMap["grossPrice"].(float64); ok {
						rates.GrossPrice = grossPrice
					}
					if payAtProperty, ok := ratesMap["payAtProperty"].(float64); ok {
						rates.PayAtProperty = payAtProperty
					}

					quoteInput.Rates = rates
				}

				// Agregar a la lista de quotes
				allQuotes = append(allQuotes, quoteInput)
			} else {
			}
		}
	}

	// Llenar campos T* del log desde el primer bookingCode (si hay)
	if len(bookingCodes) > 0 {
		bc := bookingCodes[0]
		preBookLog.TProviderID = strconv.Itoa(providerId)
		preBookLog.TCheckInDate = bc.CheckInDate.Format("2006-01-02")
		preBookLog.TCheckOutDate = bc.CheckOutDate.Format("2006-01-02")
		preBookLog.TNumberOfRooms = bc.NumberOfRooms
		preBookLog.TAdults = bc.Adults
		preBookLog.TChildren = bc.Children
		preBookLog.TChildrenAges = bc.ChildrenAges
		preBookLog.TInfant = bc.Infant
		preBookLog.TCustomerCountry = bc.CustomerCountry
		preBookLog.TMarket = bc.Market
		preBookLog.TGiHotelCode = bc.GiHotelCode
		preBookLog.TPrvHotelCode = bc.PrvHotelCode
		preBookLog.TDistribution = bc.Distribution
		preBookLog.TAdultAges = bc.AdultAges
		preBookLog.TNumberOfService = len(bookingCodes)

		// Arrays de datos por room
		preBookLog.TRoomId = []string{}
		preBookLog.TBoardId = []string{}
		preBookLog.TBuyPriceAvail = []float32{}
		preBookLog.TGiRoomID = []string{}
		preBookLog.TGiRoomCode = []string{}
		preBookLog.TGiRoomName = []string{}
		preBookLog.TPrvRoomCode = []string{}
		preBookLog.TPrvRoomName = []string{}
		preBookLog.TIsDynamicRoom = []bool{}

		var totalBuyPriceAvail float32 = 0
		for _, bc := range bookingCodes {
			preBookLog.TRoomId = append(preBookLog.TRoomId, bc.Id)
			preBookLog.TBoardId = append(preBookLog.TBoardId, bc.BoardId)
			preBookLog.TBuyPriceAvail = append(preBookLog.TBuyPriceAvail, bc.BuyPrice)
			totalBuyPriceAvail += bc.BuyTotalPrice
			preBookLog.TGiRoomID = append(preBookLog.TGiRoomID, bc.GIRoomID)
			preBookLog.TGiRoomCode = append(preBookLog.TGiRoomCode, bc.GIRoomCode)
			preBookLog.TGiRoomName = append(preBookLog.TGiRoomName, bc.GIRoomName)
			preBookLog.TPrvRoomCode = append(preBookLog.TPrvRoomCode, bc.PrvRoomCode)
			preBookLog.TPrvRoomName = append(preBookLog.TPrvRoomName, bc.PrvRoomName)
			preBookLog.TIsDynamicRoom = append(preBookLog.TIsDynamicRoom, bc.IsDynamicRoom == 1)
		}
		preBookLog.TBuyTotalPriceAvail = totalBuyPriceAvail
	}

	// Guardar TODOS los roomStays de la request para procesarlos en la respuesta
	requestContext["roomStays"] = giHotelReservations[0].RoomStays.RoomStay
	requestContext["primaryLangID"] = domainReq.PrimaryLangID
	requestContext["version"] = domainReq.Version

	// Guardar boardId (MealPlanCode) del primer bookingCode para usar en la respuesta
	if len(bookingCodes) > 0 {
		requestContext["boardId"] = bookingCodes[0].BoardId
	}

	// Guardar bookingChannel para usar en la respuesta
	requestContext["bookingChannel"] = map[string]interface{}{
		"id":   domainReq.Pos.Source.BookingChannel.ID,
		"code": domainReq.Pos.Source.BookingChannel.Code,
		"name": domainReq.Pos.Source.BookingChannel.Name,
	}

	sessionCtx.Data().ProviderCode = domainReq.InternalCondition.Channels.Channel[0].ID

	// Guardar el log en la sesión para completarlo después en la respuesta
	sessionCtx.Data().PreBookLog = preBookLog
	sessionCtx.Data().PreBookLogStartTime = time.Now()

	sessionCtx.Data().ContextData = contextData

	// Asignar query y variables al request del proveedor
	providerRQ.Query = QuoteQuery
	providerRQ.Variables = QuoteVariables{
		Quote: allQuotes,
	}

	return providerRQ
}

// Helper function para convertir interface{} a string
func getStringFromInterface(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return ""
	}
}

func attachPrebookErrors(domainResp *domain.BaseJsonRS[*domain.PreBookResponse], errorMessage string) {
	if errorMessage == "" {
		return
	}

	domainResp.Errors = BuildGIErrorContainer(errorMessage)

	if domainResp.InternalCondition == nil {
		domainResp.InternalCondition = &giresponsecommon.InternalCondition{}
	}
	domainResp.InternalCondition.ProviderStatus = "400"
	domainResp.InternalCondition.ProviderStatusDescription = errorMessage
}

// buildEmptyPreBookResponse construye una respuesta vacía cuando rsXML es nulo
func buildEmptyPreBookResponse(
	refId string,
	requestContext map[string]interface{},
	preBookLog *log_domain.HotelResBookLog,
	startTime time.Time,
	supplierRsTime int64,
	supplierRsHttpStatusCode int,
	supplierRsLength int,
	supplierErrorMessage string,
) *domain.BaseJsonRS[*domain.PreBookResponse] {
	// Crear respuesta vacía con HotelReservation y TpaExtensions vacíos
	preBookResponseOnly := &domain.PreBookResponse{
		HotelReservation: []domain.HotelReservation{},
	}

	emptyString := ""

	domainResp := &domain.BaseJsonRS[*domain.PreBookResponse]{
		EchoToken:       refId,
		PrimaryLangID:   requestContext["primaryLangID"].(string),
		SchemaLocation:  "",
		Success:         "",
		Version:         requestContext["version"].(string),
		ResResponseType: "Pending",
		TpaExtensions: &giresponsecommon.TpaExtensions{
			TpaExtension: []giresponsecommon.TpaExtension{
				{
					Client: &giresponsecommon.TpaClient{
						ID:   &emptyString,
						Name: &emptyString,
					},
					File:         &giresponsecommon.TpaFile{},
					InternalUser: &giresponsecommon.TpaInternalUser{ID: &emptyString, MessagePassword: &emptyString},
				},
			},
		},
		InternalCondition: &giresponsecommon.InternalCondition{
			Message:                   "",
			ProviderStatus:            "",
			ProviderStatusDescription: "",
			RsList:                    nil,
			Status:                    "200",
			StatusDescription:         "OK",
		},
		HotelReservations: preBookResponseOnly,
	}

	// Calcular TotalTime (tiempo total en ms)
	totalTime := int(time.Since(startTime).Milliseconds())

	// Guardar siempre RsInternal en prebook
	var rsInternal string = ""
	if rsBytes, err := json.Marshal(domainResp); err == nil {
		rsInternal = string(rsBytes)
	}

	// Determinar Success y Error según el patrón del otro conector:
	// - "OK": Operación exitosa
	// - "KO": Error de infraestructura (timeout, conexión)
	// - "Error": Error del proveedor (SOAP Fault, Errors en XML)
	success := "OK"
	errorValue := "false"
	errorCode := 0
	errorMessage := ""
	internalMessage := ""

	// Verificar si hay error del proveedor
	if supplierErrorMessage != "" {
		success = "KO"
		errorValue = "true"
		errorCode = 1
		errorMessage = supplierErrorMessage
	} else if supplierRsHttpStatusCode != 200 {
		success = "KO"
		errorValue = "true"
		errorCode = 1
		errorMessage = "HTTP error: " + strconv.Itoa(supplierRsHttpStatusCode)
	}

	attachPrebookErrors(domainResp, errorMessage)

	// Extraer datos de la respuesta para completar el log
	resResponseType := "Pending"
	domainResp.ResResponseType = resResponseType
	resStatus := "200"
	if domainResp.InternalCondition != nil {
		resStatus = domainResp.InternalCondition.Status
	}

	// Calcular TaxesAmountAvail y RetailAmountAvail desde la respuesta (vacíos)
	taxesAmountAvail := []float64{}
	retailAmountAvail := []string{}
	var totalTaxesAmountAvail float64 = 0

	// TaxesAmountBook y RetailAmountBook se establecerán cuando se haga el book
	taxesAmountBook := []float64{}
	retailAmountBook := []string{}
	var totalTaxesAmountBook float64 = 0

	// RsBookingCode desde la respuesta (vacío)
	rsBookingCode := []string{}

	// Completar todos los campos del log
	preBookLog.Success = success
	preBookLog.Error = errorValue
	preBookLog.ErrorCode = errorCode
	preBookLog.ErrorMessage = errorMessage
	preBookLog.InternalMessage = internalMessage
	preBookLog.TotalTime = totalTime
	preBookLog.SupplierRsTime = supplierRsTime
	preBookLog.SupplierRsHttpStatusCode = supplierRsHttpStatusCode
	preBookLog.SupplierRsLength = supplierRsLength
	preBookLog.SupplierErrorMessage = supplierErrorMessage
	preBookLog.RsInternal = rsInternal
	preBookLog.ResResponseType = resResponseType
	preBookLog.ResStatus = resStatus
	preBookLog.SentToSupplier = true
	preBookLog.TaxesAmountAvail = taxesAmountAvail
	preBookLog.RetailAmountAvail = retailAmountAvail
	preBookLog.TotalTaxesAmountAvail = totalTaxesAmountAvail
	preBookLog.TaxesAmountBook = taxesAmountBook
	preBookLog.RetailAmountBook = retailAmountBook
	preBookLog.TotalTaxesAmountBook = totalTaxesAmountBook
	preBookLog.RsBookingCode = rsBookingCode

	// Mantener el campo success vacío en la respuesta
	domainResp.Success = ""

	return domainResp
}

func ProviderPrebookResponseToGI(providerPrebookRS *ProviderPrebookRS, req *domain.PreBookRequest) *domain.BaseJsonRS[*domain.PreBookResponse] {
	sessionCtx := session.FromContext()

	// Obtener el log de la sesión
	preBookLog := sessionCtx.Data().PreBookLog
	if preBookLog == nil {
		panic(domain.ErrorPreBookLogNotFound)
	}

	// Obtener tiempo de inicio para calcular TotalTime
	startTime := sessionCtx.Data().PreBookLogStartTime
	if startTime.IsZero() {
		startTime = time.Now()
	}

	// Obtener datos del HTTP response del proveedor
	var supplierRsTime int64 = 0
	var supplierRsHttpStatusCode int = 200
	var supplierRsLength int = 0
	var supplierErrorMessage string = ""

	if metrics := sessionCtx.Data().SupplierMetrics; metrics != nil {
		supplierRsTime = metrics.RsTime
		supplierRsHttpStatusCode = metrics.HttpStatusCode
		supplierRsLength = metrics.RsLength
		supplierErrorMessage = metrics.ErrorMessage
	}

	contextData := sessionCtx.Data().ContextData
	if contextData == nil {
		panic(domain.ErrorContextDataNotFound)
	}
	requestContext := contextData["request"].(map[string]interface{})

	// Obtener TODOS los roomStays de la request
	roomStaysInterface, ok := requestContext["roomStays"]
	if !ok {
		panic(domain.ErrorRoomStaysNotFound)
	}

	// Convertir a slice de RQRoomStay
	var requestRoomStays []domain.RQRoomStay
	if roomStays, ok := roomStaysInterface.([]domain.RQRoomStay); ok {
		requestRoomStays = roomStays
	} else {
		customErr := domain.ErrorRoomStaysNotFound
		customErr.Err = fmt.Errorf("roomStays is not of type []domain.RQRoomStay")
		panic(customErr)
	}

	// Obtener providerCode y primaryLangID
	intId := sessionCtx.Data().ProviderCode
	// primaryLangID := requestContext["primaryLangID"].(string)

	// Obtener repositoryService
	repositoryService, ok := registry.Get[persistence.RepositoryService]("repository")
	if !ok {
		panic(domain.ErrorRepositoryNotFound)
	}

	// GI
	// PreBookResponse solo contiene HotelReservation
	// TpaExtensions e InternalCondition se asignan directamente en BaseJsonRS
	tpaExtension := &giresponsecommon.TpaExtension{}
	emptyString := ""
	tpaExtension.Client = &giresponsecommon.TpaClient{
		ID:   &emptyString,
		Name: &emptyString,
	}
	tpaExtension.InternalUser = &giresponsecommon.TpaInternalUser{
		ID:              &emptyString,
		MessagePassword: &emptyString,
	}
	tpaExtension.File = &giresponsecommon.TpaFile{}
	tpaExtension.File.Services = &giresponsecommon.TpaServices{}

	// Crear servicios y roomStays para cada roomStay de la request
	allServices := []giresponsecommon.TpaService{}
	allRoomStays := []domain.RoomStay{}

	// Acumuladores
	totalDistributionBuilder := strings.Builder{}
	var totalAdults, totalChildren int

	// Iterar sobre cada room de la respuesta del proveedor
	providerRooms := providerPrebookRS.Data.Quote.Rooms
	for roomStayIndex, providerRoom := range providerRooms {
		// Deserializar el bookingCode correspondiente de la request
		if roomStayIndex >= len(requestRoomStays) {
			break
		}

		requestRoomStay := requestRoomStays[roomStayIndex]
		bookingCodeStr := requestRoomStay.RoomRates.RoomRate.OpenBookingCode
		bookingCode := &bookingcode.InternalBookingCode{}
		bookingCode.Deserialize(bookingCodeStr)

		// Extraer guestAges del extraParams (quote guardado en base64)
		extraParams := bookingcode.StringToMap(bookingCode.ExtraParams)
		quoteJSON := bookingcode.DecodeQuoteFromExtraParams(extraParams["quote"])
		var quoteData map[string]interface{}
		json.Unmarshal([]byte(quoteJSON), &quoteData)

		// Obtener guestAges del quote (formato "30,30")
		guestAgesStr := ""
		if occupancy, ok := quoteData["occupancy"].(map[string]interface{}); ok {
			if ages, ok := occupancy["guestAges"].(string); ok {
				guestAgesStr = ages
			}
		}

		// Parsear guestAges para calcular ocupación
		roomAdults := 0
		roomChildren := 0
		roomInfants := 0
		adultAgesBuilder := strings.Builder{}
		childAges := []string{}

		if guestAgesStr != "" {
			ages := strings.Split(guestAgesStr, ",")
			for _, ageStr := range ages {
				age, _ := strconv.Atoi(ageStr)
				ageFormatted := CastTo2Digits(ageStr)

				if age >= 18 {
					roomAdults++
					adultAgesBuilder.WriteString(ageFormatted)
				} else if age >= 2 {
					roomChildren++
					childAges = append(childAges, ageFormatted)
				} else {
					roomInfants++
					childAges = append(childAges, ageFormatted)
				}
			}
		} else {
			// Fallback: usar datos del bookingCode
			roomAdults = bookingCode.Adults
			roomChildren = bookingCode.Children
		}

		// Distribución (ej: "20" = 2 adultos, 0 niños)
		totalNonAdults := roomChildren + roomInfants
		roomDistribution := fmt.Sprintf("%d%d", roomAdults, totalNonAdults)
		roomAdultAges := adultAgesBuilder.String()
		roomChildrenAges := strings.Join(childAges, "")

		totalDistributionBuilder.WriteString(roomDistribution)
		totalAdults += roomAdults
		totalChildren += roomChildren + roomInfants

		service := &giresponsecommon.TpaService{}

		// Actualizar bookingCode con datos de la respuesta de PreBook
		bookingCode.PreBookCode = providerRoom.HTIdentifier
		extraParams["htIdentifier"] = providerRoom.HTIdentifier
		bookingCode.ExtraParams = bookingcode.MapToString(extraParams)

		// Serializar el bookingCode actualizado
		encryptedBookingCode := bookingCode.Serialize()

		service.AccommodationID = bookingCode.GiHotelCode

		// Obtener nombre del hotel desde BD
		giHotelName := ""
		if hotel := repositoryService.GiHotelFromExternalCode(bookingCode.PrvHotelCode, intId); hotel != nil {
			giHotelName = hotel.HotelName
		}
		service.AccommodationName = giHotelName

		service.AdultAges = formatAges(roomAdultAges)
		service.BillingProviderID = nil
		service.BookingCode = encryptedBookingCode

		// Usar los precios de la respuesta del proveedor
		roomGrossPrice := providerRoom.Rates.GrossPrice

		service.BuyAmount = &roomGrossPrice
		service.RetailAmount = &roomGrossPrice

		service.BuyChannel = &giresponsecommon.TpaSaleChannel{
			ID:   bookingCode.ProviderCode,
			Name: "",
		}
		service.BuyCurrency = providerRoom.Rates.Currency

		// Mapear CancellationPolicies
		cancelPolicies := &giresponsecommon.TpaCancelPolicies{}
		cancelPolicies.Identifier = 0
		cancelPolicies.DateRange = []giresponsecommon.TpaCancelDate{}

		if len(providerRoom.CancellationPolicies) > 0 {
			for _, cancelPolicy := range providerRoom.CancellationPolicies {
				startSpain, endSpain := convertPolicyWindowToSpain(cancelPolicy)

				giCancelPolicy := &giresponsecommon.TpaCancelDate{
					Currency: cancelPolicy.Currency,
					Start:    startSpain,
					End:      endSpain,
					Type:     "amount",
					Amount:   &cancelPolicy.CancellationCharge,
					Value:    0,
					Source:   nil,
				}
				cancelPolicies.DateRange = append(cancelPolicies.DateRange, *giCancelPolicy)
			}
		} else {
			// Si no hay penalties, poner por defecto 100% desde hoy hasta check-in
			defaultCancelPolicy := &giresponsecommon.TpaCancelDate{
				Currency: providerRoom.Rates.Currency,
				Start:    time.Now().Format("2006-01-02"),
				End:      bookingCode.CheckInDate.Format("2006-01-02"),
				Type:     "percentage",
				Value:    100,
				Amount:   nil,
				Source:   nil,
			}
			cancelPolicies.DateRange = append(cancelPolicies.DateRange, *defaultCancelPolicy)
		}
		service.CancelPolicies = cancelPolicies

		service.ChildrensAges = formatAges(roomChildrenAges)

		currency := providerRoom.Rates.Currency
		service.Currency = &currency
		service.DecreaseQuota = false
		service.Distribution = roomDistribution
		service.ExchangeRate = nil
		service.Identifier = roomStayIndex
		service.Market = bookingCode.Market
		service.Markup = nil
		service.Nationality = bookingCode.CustomerCountry
		service.NonRefundableRate = !providerRoom.Refundable
		service.NumberOfAdults = roomAdults
		service.NumberOfChildren = roomChildren + roomInfants
		service.NumberOfRooms = 1
		service.RebookProserID = nil
		service.Remarks = nil

		// Comments del proveedor
		consolidatedComments := providerPrebookRS.Data.Quote.ConsolidatedComments
		if consolidatedComments != "" || providerRoom.Message != "" {
			comments := []giresponsecommon.TpaComment{}
			if consolidatedComments != "" {
				comments = append(comments, giresponsecommon.TpaComment{
					Text: consolidatedComments,
					Type: domain.CommentTypeProvider,
				})
			}
			if providerRoom.Message != "" {
				comments = append(comments, giresponsecommon.TpaComment{
					Text: providerRoom.Message,
					Type: domain.CommentTypeProvider,
				})
			}
			service.Comments = &giresponsecommon.TpaComments{
				Comment: comments,
			}
		}

		tpaRooms := &giresponsecommon.TpaRooms{}
		tpaRoom := &giresponsecommon.TpaRoom{}

		tpaRoom.AdultsAges = formatAges(roomAdultAges)

		// Board: obtener desde la respuesta del proveedor
		giBoard := repositoryService.GetBoardFromExternalCode(providerRoom.MealplanOptions.MealplanCode)
		if giBoard != nil {
			tpaRoom.Board = &giresponsecommon.TpaBoard{
				ID:   giBoard.RegimenID,
				Name: giBoard.Descripcion,
			}
		} else {
			// Fallback: usar bookingCode.BoardId
			tpaRoom.Board = &giresponsecommon.TpaBoard{
				ID:   bookingCode.BoardId,
				Name: providerRoom.MealplanOptions.MealplanName,
			}
		}
		if len(bookingCode.GIRoomID) > 0 {
			tpaRoom.ID = bookingCode.GIRoomID
			rmID, _ := strconv.Atoi(bookingCode.GIRoomID)
			tpaRoom.Identifier = rmID
		}
		tpaRoom.SupplierName = bookingCode.PrvRoomName
		tpaRoom.SupplierRoomID = bookingCode.PrvRoomCode
		tpaRoom.Code = bookingCode.GIRoomCode
		tpaRoom.Name = bookingCode.GIRoomName
		tpaRoom.Language = bookingCode.Lang

		tpaRoom.ChildrenAges = formatAges(roomChildrenAges)
		tpaRoom.Distribution = roomDistribution
		// tpaRoom.ID = giRoom.Id
		// tpaRoom.Identifier, _ = strconv.Atoi(giRoom.Id)
		tpaRoom.NumberOfAdults = roomAdults
		tpaRoom.NumberOfChildren = roomChildren + roomInfants
		tpaRoom.Status = nil

		// SumUpDetails con los precios de la respuesta
		tpaRoom.SumUpDetails = &giresponsecommon.TpaSumUpContainer{
			SumUpDetails: []giresponsecommon.TpaSumUp{
				{
					BuyAmount:    &roomGrossPrice,
					BuyCurrency:  providerRoom.Rates.Currency,
					Concept:      bookingCode.PrvRoomName,
					Quantity:     1,
					SaleAmount:   &roomGrossPrice,
					SaleCurrency: bookingCode.SaleCurrency,
					Type:         "base",
				},
			},
		}
		// tpaRoom.SupplierName = bookingCode.PrvRoomName
		// tpaRoom.SupplierRoomID = bookingCode.PrvRoomCode

		tpaRooms.Rooms = append(tpaRooms.Rooms, *tpaRoom)
		service.Rooms = tpaRooms

		service.SaleAmount = &roomGrossPrice
		service.SaleCurrency = bookingCode.SaleCurrency
		service.Status = nil

		service.StayDateRange = &giresponsecommon.TpaStayRange{
			Start: bookingCode.CheckInDate.Format("2006-01-02"),
			End:   bookingCode.CheckOutDate.Format("2006-01-02"),
		}
		service.SupplierAccommodationID = bookingCode.PrvHotelCode
		service.SupplierReference = nil

		// Agregar este servicio a la lista
		allServices = append(allServices, *service)

		// Crear un RoomStay para esta habitación
		preBookRoomStay := domain.RoomStay{}

		// CancelPenalties - mapear desde la respuesta
		cancelPenaltiesRS := &giresponsecommon.CancelPenalties{}
		cancelPenaltiesRS.CancelPenalty = []giresponsecommon.CancelPenalty{}

		if len(providerRoom.CancellationPolicies) > 0 {
			for _, cancelPolicy := range providerRoom.CancellationPolicies {
				startSpain, endSpain := convertPolicyWindowToSpain(cancelPolicy)

				giCancelPenalty := &giresponsecommon.CancelPenalty{
					AmountPercent: giresponsecommon.AmountPercent{
						CurrencyCode: cancelPolicy.Currency,
						Amount:       fmt.Sprintf("%.2f", cancelPolicy.CancellationCharge),
					},
					End:   endSpain,
					Start: startSpain,
				}

				cancelPenaltiesRS.CancelPenalty = append(cancelPenaltiesRS.CancelPenalty, *giCancelPenalty)
			}
		} else {
			// Si no hay penalties, poner por defecto 100% desde hoy hasta check-in
			defaultCancelPenalty := &giresponsecommon.CancelPenalty{
				AmountPercent: giresponsecommon.AmountPercent{
					CurrencyCode: providerRoom.Rates.Currency,
					Percent:      "100",
				},
				End:   bookingCode.CheckInDate.Format("2006-01-02"),
				Start: time.Now().Format("2006-01-02"),
			}
			cancelPenaltiesRS.CancelPenalty = append(cancelPenaltiesRS.CancelPenalty, *defaultCancelPenalty)
		}
		preBookRoomStay.CancelPenalties = cancelPenaltiesRS

		// Comments - usar consolidatedComments y message de la respuesta
		if consolidatedComments != "" {
			preBookRoomStay.Comments.Comment = append(preBookRoomStay.Comments.Comment, domain.Comment{
				Text: consolidatedComments,
			})
		}
		if providerRoom.Message != "" {
			preBookRoomStay.Comments.Comment = append(preBookRoomStay.Comments.Comment, domain.Comment{
				Text: providerRoom.Message,
			})
		}

		// RoomRates
		preBookRoomStay.RoomRates.RoomRate.OpenBookingCode = encryptedBookingCode
		preBookRoomStay.RoomRates.RoomRate.BookingCode = ""

		// Usar precios de la respuesta
		preBookRoomStay.RoomRates.RoomRate.Total.Taxes.Amount = roomGrossPrice
		preBookRoomStay.RoomRates.RoomRate.Total.Taxes.CurrencyCode = providerRoom.Rates.Currency
		preBookRoomStay.RoomRates.RoomRate.Total.Taxes.RetailAmount = &roomGrossPrice

		// NonRefundable
		if !providerRoom.Refundable {
			preBookRoomStay.RoomRates.RoomRate.Total.NonRefundable = "1"
		} else {
			preBookRoomStay.RoomRates.RoomRate.Total.NonRefundable = "0"
		}

		// Agregar este RoomStay a la lista
		allRoomStays = append(allRoomStays, preBookRoomStay)
	}

	// Asignar todos los servicios al tpaExtension
	tpaExtension.File.Services.Service = allServices

	// Asignar totales al File (calculados durante el loop)
	tpaExtension.File.AdultsTotal = totalAdults
	tpaExtension.File.ChildrenTotal = totalChildren
	tpaExtension.File.RoomsTotal = len(allRoomStays)
	tpaExtension.File.ServiceTotal = len(allServices)
	tpaExtension.File.Distribution = totalDistributionBuilder.String() // Distribución total concatenada

	// Agregar saleChannel desde el request (pos.source.bookingChannel)
	if bookingChannelInterface, ok := requestContext["bookingChannel"]; ok {
		if bookingChannel, ok := bookingChannelInterface.(map[string]interface{}); ok {
			channelID := ""
			channelName := ""
			if id, ok := bookingChannel["id"].(int); ok {
				channelID = strconv.Itoa(id)
			}
			if name, ok := bookingChannel["name"].(string); ok {
				channelName = name
			}
			tpaExtension.File.SaleChannel = &giresponsecommon.TpaSaleChannel{
				ID:   channelID,
				Name: channelName,
			}
		}
	}

	// Crear PreBookResponse con HotelReservation
	preBookResponseWithHotelRes := &domain.PreBookResponse{}

	// HotelReservation con todos los RoomStays
	preBookHotelReservation := domain.HotelReservation{}
	preBookHotelReservation.RoomStays.RoomStay = allRoomStays // Usar todos los roomStays creados

	preBookResponseWithHotelRes.HotelReservation = append(preBookResponseWithHotelRes.HotelReservation, preBookHotelReservation)

	// Crear PreBookResponse que solo contiene HotelReservation
	// El resto de campos (TpaExtensions, InternalCondition, etc.) van en BaseJsonRS
	preBookResponseOnly := &domain.PreBookResponse{
		HotelReservation: preBookResponseWithHotelRes.HotelReservation,
	}

	domainResp := &domain.BaseJsonRS[*domain.PreBookResponse]{
		EchoToken:       preBookLog.EchoToken,
		PrimaryLangID:   requestContext["primaryLangID"].(string),
		SchemaLocation:  "",
		Success:         "",
		Version:         requestContext["version"].(string),
		ResResponseType: "Pending",
		TpaExtensions: &giresponsecommon.TpaExtensions{
			TpaExtension: []giresponsecommon.TpaExtension{*tpaExtension},
		},
		InternalCondition: &giresponsecommon.InternalCondition{
			Message:                   "",
			ProviderStatus:            "",
			ProviderStatusDescription: "",
			RsList:                    nil,
			Status:                    "200",
			StatusDescription:         "OK",
		},
		// Solo asignar HotelReservations (que contiene hotelReservation[])
		HotelReservations: preBookResponseOnly,
	}

	// Calcular TotalTime (tiempo total en ms)
	totalTime := int(time.Since(startTime).Milliseconds())

	// Guardar siempre RsInternal en prebook (JSON que devuelve el conector)
	var rsInternal string = ""
	if rsBytes, err := json.Marshal(domainResp); err == nil {
		rsInternal = string(rsBytes)
	}
	// rsProvider se asigna en ot_client.go con el XML que devuelve el proveedor
	// No se asigna aquí

	// Determinar Success y Error según el patrón del otro conector:
	// - "OK": Operación exitosa
	// - "KO": Error de infraestructura (timeout, conexión)
	// - "Error": Error del proveedor (SOAP Fault, Errors en XML)
	success := "OK"
	errorValue := "false"
	errorCode := 0
	errorMessage := ""
	internalMessage := ""

	// Primero verificar si hay error del proveedor
	if supplierErrorMessage != "" {
		success = "KO"
		errorValue = "true"
		errorCode = 1
		errorMessage = supplierErrorMessage
	} else if supplierRsHttpStatusCode != 200 {
		success = "KO"
		errorValue = "true"
		errorCode = 1
		errorMessage = "HTTP error: " + strconv.Itoa(supplierRsHttpStatusCode)
	} else if domainResp.Success != "" && domainResp.Success != "true" && domainResp.Success != "OK" {
		success = "KO"
		errorValue = "true"
		errorCode = 1
		if domainResp.InternalCondition != nil {
			internalMessage = domainResp.InternalCondition.StatusDescription
			if domainResp.InternalCondition.Status != "200" {
				errorMessage = domainResp.InternalCondition.StatusDescription
			}
		}
	}

	attachPrebookErrors(domainResp, errorMessage)

	// Extraer datos de la respuesta para completar el log
	resResponseType := "Pending"
	domainResp.ResResponseType = resResponseType
	resStatus := ""
	if domainResp.InternalCondition != nil {
		resStatus = domainResp.InternalCondition.Status
	}

	// Calcular TaxesAmountAvail y RetailAmountAvail desde la respuesta
	taxesAmountAvail := []float64{}
	retailAmountAvail := []string{}
	var totalTaxesAmountAvail float64 = 0

	// TaxesAmountBook y RetailAmountBook se establecerán cuando se haga el book
	taxesAmountBook := []float64{}
	retailAmountBook := []string{}
	var totalTaxesAmountBook float64 = 0

	// Si hay servicios en la respuesta, extraer precios
	totalTaxesAmountAvail = 0
	if len(allServices) > 0 {
		for _, svc := range allServices {
			if svc.BuyAmount != nil {
				taxesAmountAvail = append(taxesAmountAvail, float64(*svc.BuyAmount))
				totalTaxesAmountAvail += float64(*svc.BuyAmount)
			}
			if svc.RetailAmount != nil {
				retailAmountAvail = append(retailAmountAvail, strconv.FormatFloat(float64(*svc.RetailAmount), 'f', -1, 32))
			}
		}
	}

	// RsBookingCode desde la respuesta
	rsBookingCode := []string{}
	for _, svc := range allServices {
		if svc.BookingCode != "" {
			rsBookingCode = append(rsBookingCode, svc.BookingCode)
		}
	}

	// Completar todos los campos del log
	preBookLog.Success = success
	preBookLog.Error = errorValue
	preBookLog.ErrorCode = errorCode
	preBookLog.ErrorMessage = errorMessage
	preBookLog.InternalMessage = internalMessage
	preBookLog.TotalTime = totalTime
	preBookLog.SupplierRsTime = supplierRsTime
	preBookLog.SupplierRsHttpStatusCode = supplierRsHttpStatusCode
	preBookLog.SupplierRsLength = supplierRsLength
	preBookLog.SupplierErrorMessage = supplierErrorMessage
	preBookLog.RsInternal = rsInternal
	// RsProvider se asigna en ot_client.go con el XML que devuelve el proveedor
	// preBookLog.RsProvider = rsProvider
	preBookLog.ResResponseType = resResponseType
	preBookLog.ResStatus = resStatus
	preBookLog.SentToSupplier = true
	preBookLog.TaxesAmountAvail = taxesAmountAvail
	preBookLog.RetailAmountAvail = retailAmountAvail
	preBookLog.TotalTaxesAmountAvail = totalTaxesAmountAvail
	preBookLog.TaxesAmountBook = taxesAmountBook
	preBookLog.RetailAmountBook = retailAmountBook
	preBookLog.TotalTaxesAmountBook = totalTaxesAmountBook
	preBookLog.RsBookingCode = rsBookingCode

	// Mantener el campo success vacío en la respuesta
	domainResp.Success = ""

	return domainResp
}

// getMapKeys returns the keys of a map[string]string
func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
