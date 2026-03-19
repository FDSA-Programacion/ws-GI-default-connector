package hoteltrader

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"ws-int-httr/internal/domain"
	common_domain "ws-int-httr/internal/domain/gi_response_common"
	"ws-int-httr/internal/domain/log_domain"
	bookingcode "ws-int-httr/internal/infrastructure/booking_code"
	"ws-int-httr/internal/infrastructure/config"
	"ws-int-httr/internal/infrastructure/session"
)

// GIBookRequestToProvider convierte una petición del dominio a una petición GraphQL de Book
func GIBookRequestToProvider(req *domain.BookRequest, cfg config.ProviderConfig) *ProviderBookRQ {
	sessionCtx := session.FromContext()
	contextData := map[string]interface{}{}
	contextData["Request"] = map[string]interface{}{}

	bookLog := &log_domain.HotelResCommitLog{}
	sessionCtx.Data().BookLog = bookLog
	sessionCtx.Data().BookLogStartTime = time.Now()

	var bookingCodes []*bookingcode.InternalBookingCode
	var numResGuest int = 0
	effectiveTransactionIdentifier := req.TransactionIdentifier
	if req.RebTransactionIdentifier != "" {
		effectiveTransactionIdentifier = req.RebTransactionIdentifier
	}

	// ContextData para uso posterior
	requestContext := map[string]interface{}{
		"echoToken":             req.InternalCondition.CallCondition.EchoToken,
		"primaryLangID":         req.PrimaryLangID,
		"version":               req.Version,
		"transactionIdentifier": effectiveTransactionIdentifier,
	}

	giHotelReservation := req.HotelReservations.HotelReservation
	// Agregar creatorID si está disponible
	if len(giHotelReservation) > 0 {
		requestContext["creatorID"] = giHotelReservation[0].CreatorID
	}

	// Inicializar BookLog
	hostname, _ := os.Hostname()
	providerId := req.InternalCondition.Channels.Channel[0].ID
	providerCode := req.InternalCondition.Channels.Channel[0].Code

	bookLog.ProviderID = strconv.Itoa(providerId)
	bookLog.ProviderCode = providerCode
	bookLog.Integration = "WS-INT-HTTR"
	bookLog.Node = hostname
	bookLog.EchoToken = req.InternalCondition.CallCondition.EchoToken
	bookLog.RqType = req.RqType
	bookLog.RequestorID = req.Pos.Source.RequestorID.ID
	bookLog.RqTimestamp = int64(req.InternalCondition.CallCondition.OriginTimeStamp)
	bookLog.PrimaryLangID = req.PrimaryLangID
	bookLog.Version = req.Version
	bookLog.ClientName = req.Pos.Source.RequestorID.CompanyName.CompanyShortName
	bookLog.ClientCode = req.InternalCondition.ClientCondition.Code
	bookLog.BookingChannel = req.Pos.Source.BookingChannel.Code
	bookLog.RqInternal = getStringFromInterface(req.InternalCondition.ID)
	bookLog.RqProvider = strconv.Itoa(req.InternalCondition.Channels.Channel[0].ID)
	bookLog.SentToSupplier = false
	bookLog.IsRebook = req.InternalCondition.Rebook
	bookLog.RqBookingCode = []string{}
	if rqBytes, err := json.Marshal(req); err == nil {
		bookLog.RqInternal = string(rqBytes)
	}
	if effectiveTransactionIdentifier != "" {
		bookLog.TransactionIdentifier = effectiveTransactionIdentifier
	}
	if len(giHotelReservation) > 0 {
		bookLog.HotelReservationCreatorID = giHotelReservation[0].CreatorID
	}
	sessionCtx.Data().Debug = req.Debug

	// Nivel Book: clientConfirmationCode es siempre el transactionIdentifier (ej: "GI154591593")
	clientConfirmationCode := effectiveTransactionIdentifier
	otaConfirmationCode := effectiveTransactionIdentifier
	if otaConfirmationCode == "" {
		otaConfirmationCode = clientConfirmationCode
	}
	otaClientName := "GuestIncoming"

	// SpecialRequests a nivel reserva (comentarios generales)
	var specialRequests []string
	var rooms []BookRoomInput

	for hrIdx, hotelRes := range req.HotelReservations.HotelReservation {
		// Comentarios de reserva: primer room stay o vacío
		for _, roomStay := range hotelRes.RoomStays.RoomStay {
			for _, c := range roomStay.Comments.Comment {
				if c.Text != "" {
					specialRequests = append(specialRequests, c.Text)
					break
				}
			}
			break
		}

		// Edades de huéspedes para asignar por índice (guestAges del quote)
		guestAgesPerRoom := []string{} // se rellena por habitación desde quote

		for roomIdx, roomStay := range hotelRes.RoomStays.RoomStay {
			bookingCodeStr, ok := roomStay.RoomRates.RoomRate.OpenBookingCode.(string)
			if !ok || bookingCodeStr == "" {
				continue
			}

			bookingCode := &bookingcode.InternalBookingCode{}
			bookingCode.Deserialize(bookingCodeStr)
			bookingCodes = append(bookingCodes, bookingCode)
			bookLog.RqBookingCode = append(bookLog.RqBookingCode, bookingCodeStr)
			extraParams := bookingcode.StringToMap(bookingCode.ExtraParams)

			// --- Extraer quote desde ExtraParams (guardado en base64 en Avail/PreBook) ---
			quoteJSON := bookingcode.DecodeQuoteFromExtraParams(extraParams["quote"])
			if quoteJSON == "" {
				customErr := domain.ErrorInvalidJSON
				customErr.Err = fmt.Errorf("quote no encontrado en ExtraParams del booking code")
				panic(customErr)
			}

			var quoteData map[string]interface{}
			if err := json.Unmarshal([]byte(quoteJSON), &quoteData); err != nil {
				customErr := domain.ErrorInvalidJSON
				customErr.Err = fmt.Errorf("error parseando quote en ExtraParams: %w", err)
				panic(customErr)
			}

			htIdentifier := ""
			if htId, exists := extraParams["htIdentifier"]; exists && htId != "" {
				htIdentifier = htId
			} else if htId, ok := quoteData["htIdentifier"].(string); ok {
				htIdentifier = htId
			}
			if htIdentifier == "" {
				customErr := domain.ErrorInvalidJSON
				customErr.Err = fmt.Errorf("htIdentifier no encontrado en ExtraParams ni en quote")
				panic(customErr)
			}

			// Rates desde quote (igual que en PreBook)
			var rates *BookRatesInput
			if ratesMap, ok := quoteData["rates"].(map[string]interface{}); ok {
				rates = &BookRatesInput{}
				if v, ok := ratesMap["netPrice"].(float64); ok {
					rates.NetPrice = v
				}
				if v, ok := ratesMap["tax"].(float64); ok {
					rates.Tax = v
				}
				if v, ok := ratesMap["grossPrice"].(float64); ok {
					rates.GrossPrice = v
				}
				if v, ok := ratesMap["payAtProperty"].(float64); ok {
					rates.PayAtProperty = v
				}
				if arr, ok := ratesMap["dailyPrice"].([]interface{}); ok {
					for _, x := range arr {
						if f, ok := x.(float64); ok {
							rates.DailyPrice = append(rates.DailyPrice, f)
						}
					}
				}
				if arr, ok := ratesMap["dailyTax"].([]interface{}); ok {
					for _, x := range arr {
						if f, ok := x.(float64); ok {
							rates.DailyTax = append(rates.DailyTax, f)
						}
					}
				}
			}

			// Occupancy desde quote
			var occupancy *BookOccupancyInput
			if occMap, ok := quoteData["occupancy"].(map[string]interface{}); ok {
				if guestAges, ok := occMap["guestAges"].(string); ok && guestAges != "" {
					occupancy = &BookOccupancyInput{GuestAges: guestAges}
					guestAgesPerRoom = strings.Split(guestAges, ",")
				}
			}

			// Comentarios de la habitación
			var roomSpecialRequests []string
			for _, c := range roomStay.Comments.Comment {
				if c.Text != "" {
					roomSpecialRequests = append(roomSpecialRequests, c.Text)
				}
			}

			// clientRoomConfirmationCode
			clientRoomConfirmationCode := fmt.Sprintf("%s-%d-%d", clientConfirmationCode, hrIdx+1, roomIdx+1)

			// Guests desde ResGuests; edad por índice desde guestAges del quote
			guests := buildBookGuestsFromResGuests(hotelRes.ResGuests.ResGuest, guestAgesPerRoom)

			room := BookRoomInput{
				HTIdentifier:               htIdentifier,
				ClientRoomConfirmationCode: clientRoomConfirmationCode,
				RoomSpecialRequests:        roomSpecialRequests,
				Rates:                      rates,
				Occupancy:                  occupancy,
				Guests:                     guests,
			}
			rooms = append(rooms, room)
		}
	}

	if len(rooms) == 0 {
		rooms = []BookRoomInput{}
	}

	bookInput := &BookRequestInput{
		ClientConfirmationCode: clientConfirmationCode,
		OtaConfirmationCode:    otaConfirmationCode,
		OtaClientName:          otaClientName,
		SpecialRequests:        specialRequests,
		PaymentInformation:     nil,
		Rooms:                  rooms,
	}

	allBookingCodes := []interface{}{}
	for _, giHotelRes := range giHotelReservation {
		numResGuest += len(giHotelRes.ResGuests.ResGuest)
		for _, roomStay := range giHotelRes.RoomStays.RoomStay {
			if roomStay.RoomRates.RoomRate.OpenBookingCode != nil {
				bookingCodeStr := roomStay.RoomRates.RoomRate.OpenBookingCode.(string)
				allBookingCodes = append(allBookingCodes, bookingCodeStr)
			}
		}
	}

	if len(allBookingCodes) > 0 {
		requestContext["bookingCodes"] = allBookingCodes
	}

	// Guardar boardId (MealPlanCode) del primer bookingCode para usar en la respuesta
	if len(bookingCodes) > 0 {
		requestContext["boardId"] = bookingCodes[0].BoardId
	}

	// Llenar campos T* del log desde los bookingCodes
	if len(bookingCodes) > 0 {
		bc := bookingCodes[0]
		bookLog.TProviderID = strconv.Itoa(providerId)
		bookLog.TCheckInDate = bc.CheckInDate.Format("2006-01-02")
		bookLog.TCheckOutDate = bc.CheckOutDate.Format("2006-01-02")
		bookLog.TNumberOfRooms = bc.NumberOfRooms
		bookLog.TAdults = bc.Adults
		bookLog.TChildren = bc.Children
		bookLog.TChildrenAges = bc.ChildrenAges
		bookLog.TInfant = bc.Infant
		bookLog.TCustomerCountry = bc.CustomerCountry
		bookLog.TMarket = bc.Market
		bookLog.TGiHotelCode = bc.GiHotelCode
		bookLog.TPrvHotelCode = bc.PrvHotelCode
		bookLog.TDistribution = bc.Distribution
		bookLog.TAdultAges = bc.AdultAges
		bookLog.TNumberOfService = 1
		bookLog.NumResGuest = numResGuest

		// Arrays de datos por room
		bookLog.TRoomId = []string{}
		bookLog.TBoardId = []string{}
		bookLog.TBuyPrice = []float32{}
		bookLog.TSalePrice = []float32{}
		bookLog.TGiRoomID = []string{}
		bookLog.TGiRoomCode = []string{}
		bookLog.TGiRoomName = []string{}
		bookLog.TPrvRoomCode = []string{}
		bookLog.TPrvRoomName = []string{}
		bookLog.TIsDynamicRoom = []bool{}

		var totalBuyPrice float32 = 0
		var totalSalePrice float32 = 0
		for _, bc := range bookingCodes {
			bookLog.TRoomId = append(bookLog.TRoomId, bc.Id)
			bookLog.TBoardId = append(bookLog.TBoardId, bc.BoardId)
			bookLog.TBuyPrice = append(bookLog.TBuyPrice, bc.BuyPrice)
			bookLog.TSalePrice = append(bookLog.TSalePrice, bc.BuyPrice) // Por defecto igual a BuyPrice
			totalBuyPrice += bc.BuyTotalPrice
			totalSalePrice += bc.BuyTotalPrice // Por defecto igual
			bookLog.TGiRoomID = append(bookLog.TGiRoomID, bc.GIRoomID)
			bookLog.TGiRoomCode = append(bookLog.TGiRoomCode, bc.GIRoomCode)
			bookLog.TGiRoomName = append(bookLog.TGiRoomName, bc.GIRoomName)
			bookLog.TPrvRoomCode = append(bookLog.TPrvRoomCode, bc.PrvRoomCode)
			bookLog.TPrvRoomName = append(bookLog.TPrvRoomName, bc.PrvRoomName)
			bookLog.TIsDynamicRoom = append(bookLog.TIsDynamicRoom, bc.IsDynamicRoom == 1)
		}
		bookLog.TBuyTotalPrice = totalBuyPrice
		bookLog.TSaleTotalPrice = totalSalePrice
	}

	contextData["Request"] = requestContext
	sessionCtx.Data().ContextData = contextData

	return &ProviderBookRQ{
		Query: BookMutation,
		Variables: BookVariables{
			Book: bookInput,
		},
	}
}

// buildBookGuestsFromResGuests mapea ResGuest del dominio a BookGuestInput; agesByIndex son edades desde quote (guestAges)
func buildBookGuestsFromResGuests(resGuests []domain.ResGuest, agesByIndex []string) []BookGuestInput {
	out := make([]BookGuestInput, 0, len(resGuests)*2)
	ageIdx := 0
	for _, resGuest := range resGuests {
		adult := resGuest.AgeQualifyingCode == "10" || resGuest.AgeQualifyingCode == "ADULT"
		for _, profile := range resGuest.Profiles.ProfileInfo.Profile {
			age := 30
			if ageIdx < len(agesByIndex) {
				if a, err := strconv.Atoi(strings.TrimSpace(agesByIndex[ageIdx])); err == nil {
					age = a
				}
				ageIdx++
			}
			out = append(out, BookGuestInput{
				FirstName: profile.Customer.PersonName.GivenName,
				LastName:  profile.Customer.PersonName.Surname,
				Email:     "test@test.es",
				Adult:     adult,
				Age:       age,
				Phone:     "666666666",
				Primary:   profile.ProfileType == "18",
			})
		}
	}
	return out
}

// ProviderBookResponseToGI convierte una respuesta GraphQL a una respuesta del dominio
func ProviderBookResponseToGI(graphqlResp *ProviderBookRS, req *domain.BookRequest) *domain.BaseJsonRS[*domain.BookResponse] {
	sessionCtx := session.FromContext()
	bookLog := sessionCtx.Data().BookLog

	// Obtener contextData del request
	contextData := sessionCtx.Data().ContextData
	if contextData == nil {
		panic("contextData not found")
	}
	requestContext := contextData["Request"].(map[string]interface{})

	startTime := sessionCtx.Data().BookLogStartTime
	if startTime.IsZero() {
		startTime = time.Now()
	}

	var supplierRsTime int64 = 0
	var supplierRsHttpStatusCode int = 0
	var supplierRsLength int = 0
	supplierErrorMessage := ""
	if metrics := sessionCtx.Data().SupplierMetrics; metrics != nil {
		supplierRsTime = metrics.RsTime
		supplierRsHttpStatusCode = metrics.HttpStatusCode
		supplierRsLength = metrics.RsLength
		supplierErrorMessage = metrics.ErrorMessage
	}

	bookResp := graphqlResp.Data.Book

	// Comentarios consolidados del proveedor: texto plano, HTML, fecha de reserva y códigos
	var commentsList []domain.Comment
	if bookResp.BookingDate != "" || bookResp.ClientConfirmationCode != "" || bookResp.OTAConfirmationCode != "" {
		header := ""
		if bookResp.BookingDate != "" {
			header = "Booking date: " + bookResp.BookingDate
		}
		if bookResp.ClientConfirmationCode != "" {
			if header != "" {
				header += "\n"
			}
			header += "Client confirmation: " + bookResp.ClientConfirmationCode
		}
		if bookResp.OTAConfirmationCode != "" {
			if header != "" {
				header += "\n"
			}
			header += "OTA confirmation: " + bookResp.OTAConfirmationCode
		}
		if bookResp.ConsolidatedComments != "" {
			if header != "" {
				header += "\n\n"
			}
			header += bookResp.ConsolidatedComments
		}
		if header != "" {
			commentsList = append(commentsList, domain.Comment{Text: header})
		}
	} else if bookResp.ConsolidatedComments != "" {
		commentsList = append(commentsList, domain.Comment{Text: bookResp.ConsolidatedComments})
	}
	if bookResp.ConsolidatedHTMLComments != "" {
		commentsList = append(commentsList, domain.Comment{Text: bookResp.ConsolidatedHTMLComments, Type: "HTML"})
	}

	var consolidatedComments *domain.BookComments
	if len(commentsList) > 0 {
		consolidatedComments = &domain.BookComments{Comment: &commentsList}
	}

	// Por habitación: bookingCode y openBookingCode de entrada (para devolverlos en la respuesta)
	type reqRoomRate struct {
		BookingCode     string
		OpenBookingCode string
	}
	reqRoomStays := []reqRoomRate{}
	if len(req.HotelReservations.HotelReservation) > 0 {
		for _, rs := range req.HotelReservations.HotelReservation[0].RoomStays.RoomStay {
			obc := ""
			if s, ok := rs.RoomRates.RoomRate.OpenBookingCode.(string); ok {
				obc = s
			}
			reqRoomStays = append(reqRoomStays, reqRoomRate{
				BookingCode:     rs.RoomRates.RoomRate.BookingCode,
				OpenBookingCode: obc,
			})
		}
	}

	roomStays := []domain.BookRoomStay{}
	for roomIdx, room := range bookResp.Rooms {
		// Políticas de cancelación
		var cancelPenalties *domain.BookCancelPenalties
		if len(room.CancellationPolicies) > 0 {
			penalties := []common_domain.CancelPenalty{}
			for _, policy := range room.CancellationPolicies {
				startSpain, endSpain := convertPolicyWindowToSpain(policy)
				penalty := common_domain.CancelPenalty{
					Start: startSpain,
					End:   endSpain,
					AmountPercent: common_domain.AmountPercent{
						Amount:       fmt.Sprintf("%.2f", policy.CancellationCharge),
						CurrencyCode: policy.Currency,
					},
				}
				penalties = append(penalties, penalty)
			}
			cancelPenalties = &domain.BookCancelPenalties{
				CancelPenalty: &penalties,
			}
		}

		// Comentarios: usar los consolidados del book (htConfirmationCode, clientConfirmationCode, consolidatedComments, consolidatedHTMLComments, bookingDate)
		comments := consolidatedComments
		if len(room.RoomSpecialRequests) > 0 {
			roomComments := []domain.Comment{}
			for _, sr := range room.RoomSpecialRequests {
				if sr != "" {
					roomComments = append(roomComments, domain.Comment{Text: sr})
				}
			}
			if comments != nil && comments.Comment != nil {
				roomComments = append(roomComments, *comments.Comment...)
			}
			if len(roomComments) > 0 {
				comments = &domain.BookComments{Comment: &roomComments}
			}
		}

		nonRefund := "0"
		if !room.Refundable {
			nonRefund = "1"
		}

		roomStay := domain.BookRoomStay{
			CancelPenalties: cancelPenalties,
			Comments:        comments,
		}

		// bookingCode y openBookingCode = códigos de entrada (petición); referencia proveedor solo en tpaExtensions.service[].supplierReference
		localizador := bookResp.HTConfirmationCode
		var bookingCodeVal, openBookingCodeVal string
		if roomIdx < len(reqRoomStays) {
			r := reqRoomStays[roomIdx]
			openBookingCodeVal = r.OpenBookingCode
			if openBookingCodeVal != "" {
				bookingCodeVal = openBookingCodeVal
			} else {
				bookingCodeVal = r.BookingCode
			}
		}
		if bookingCodeVal == "" {
			bookingCodeVal = localizador
		}
		if openBookingCodeVal == "" {
			openBookingCodeVal = localizador
		}
		roomStay.RoomRates.RoomRate.BookingCode = &bookingCodeVal
		roomStay.RoomRates.RoomRate.OpenBookingCode = openBookingCodeVal
		roomStay.RoomRates.RoomRate.Total.NonRefundable = &nonRefund

		amount := room.Rates.GrossPrice
		roomStay.RoomRates.RoomRate.Total.Taxes.Amount = &amount
		roomStay.RoomRates.RoomRate.Total.Taxes.CurrencyCode = room.Rates.CurrencyCode

		roomStays = append(roomStays, roomStay)
	}

	// creatorID desde la petición (formato GI como en el mock)
	creatorID := ""
	if len(req.HotelReservations.HotelReservation) > 0 {
		creatorID = req.HotelReservations.HotelReservation[0].CreatorID
	}

	hotelReservation := domain.BookHotelReservation{
		CreatorID: creatorID,
	}
	hotelReservation.RoomStays.RoomStay = roomStays
	hotelReservations := []domain.BookHotelReservation{hotelReservation}

	transactionIdentifier := ""
	if transactionIdentifierInterface, ok := requestContext["transactionIdentifier"]; ok {
		if transactionIdentifierStr, ok := transactionIdentifierInterface.(string); ok {
			transactionIdentifier = transactionIdentifierStr
		}
	}

	// Formato de salida según gi_book_response_mock: echoToken, hotelReservations, internalConditionRS, primaryLangID, resResponseType, success, tpaExtensions, transactionIdentifier, version
	base := &domain.BaseJsonRS[*domain.BookResponse]{
		Success: "",
		HotelReservations: &domain.BookResponse{
			HotelReservation: hotelReservations,
		},
		TransactionIdentifier: transactionIdentifier,
		// TransactionIdentifier: firstNonEmpty(req.TransactionIdentifier, bookResp.ClientConfirmationCode, bookResp.HTConfirmationCode),
		ResResponseType: "Committed",
		PrimaryLangID:   req.PrimaryLangID,
		Version:         req.Version,
		InternalCondition: &common_domain.InternalCondition{
			Status:            "200",
			StatusDescription: "OK",
		},
	}

	if base.PrimaryLangID == "" {
		base.PrimaryLangID = "es"
	}
	if base.Version == "" {
		base.Version = "1.004"
	}
	if req.InternalCondition.CallCondition.EchoToken != "" {
		base.EchoToken = req.InternalCondition.CallCondition.EchoToken
	} else if req.EchoToken != "" {
		base.EchoToken = req.EchoToken
	}

	// tpaExtensions: reutilizar de la petición (PreBook/Book) y actualizar con datos del proveedor
	clientRef := creatorID
	if clientRef == "" {
		clientRef = req.TransactionIdentifier
	}
	base.TpaExtensions = buildBookTpaExtensions(req, bookResp, creatorID, clientRef, len(roomStays))

	if bookLog != nil {
		totalTime := int(time.Since(startTime).Milliseconds())

		rsInternal := ""
		if rsBytes, err := json.Marshal(base); err == nil {
			rsInternal = string(rsBytes)
		}

		success := "OK"
		errorValue := "false"
		errorCode := 0
		errorMessage := ""
		internalMessage := ""
		if supplierErrorMessage != "" {
			success = "KO"
			errorValue = "true"
			errorCode = 1
			errorMessage = supplierErrorMessage
		} else if supplierRsHttpStatusCode != 0 && supplierRsHttpStatusCode != 200 {
			success = "KO"
			errorValue = "true"
			errorCode = 1
			errorMessage = "HTTP error: " + strconv.Itoa(supplierRsHttpStatusCode)
		} else if base.InternalCondition != nil && base.InternalCondition.Status != "" && base.InternalCondition.Status != "200" {
			success = "KO"
			errorValue = "true"
			errorCode = 1
			internalMessage = base.InternalCondition.StatusDescription
			errorMessage = base.InternalCondition.StatusDescription
		}

		attachBookErrors(base, errorMessage)

		resStatus := ""
		if base.InternalCondition != nil {
			resStatus = base.InternalCondition.Status
		}

		taxesAmount := make([]float64, 0, len(bookResp.Rooms))
		retailAmount := []string{}
		totalTaxesAmount := 0.0
		bookedRoomName := make([]string, 0, len(bookResp.Rooms))
		rsBookingCode := []string{}

		for i := range bookResp.Rooms {
			room := bookResp.Rooms[i]
			taxesAmount = append(taxesAmount, room.Rates.GrossPrice)
			totalTaxesAmount += room.Rates.GrossPrice
			if room.RoomName != "" {
				bookedRoomName = append(bookedRoomName, room.RoomName)
			}
		}
		for i := range reqRoomStays {
			if reqRoomStays[i].OpenBookingCode != "" {
				rsBookingCode = append(rsBookingCode, reqRoomStays[i].OpenBookingCode)
			} else if reqRoomStays[i].BookingCode != "" {
				rsBookingCode = append(rsBookingCode, reqRoomStays[i].BookingCode)
			}
		}

		bookLog.Success = success
		bookLog.Error = errorValue
		bookLog.ErrorCode = errorCode
		bookLog.ErrorMessage = errorMessage
		bookLog.InternalMessage = internalMessage
		bookLog.TotalTime = totalTime
		bookLog.SupplierRsTime = supplierRsTime
		bookLog.SupplierRsHttpStatusCode = supplierRsHttpStatusCode
		bookLog.SupplierRsLength = supplierRsLength
		bookLog.SupplierErrorMessage = supplierErrorMessage
		bookLog.RsInternal = rsInternal
		bookLog.ResResponseType = base.ResResponseType
		bookLog.ResStatus = resStatus
		bookLog.ProviderReference = bookResp.HTConfirmationCode
		bookLog.SentToSupplier = true
		bookLog.TaxesAmount = taxesAmount
		bookLog.RetailAmount = retailAmount
		bookLog.TotalTaxesAmount = totalTaxesAmount
		bookLog.RsBookingCode = rsBookingCode
		bookLog.BookedRoomName = bookedRoomName
		bookLog.TransactionIdentifier = transactionIdentifier
	}

	return base
}

func attachBookErrors(base *domain.BaseJsonRS[*domain.BookResponse], errorMessage string) {
	if errorMessage == "" {
		return
	}

	base.Errors = BuildGIErrorContainer(errorMessage)
	if base.InternalCondition == nil {
		base.InternalCondition = &common_domain.InternalCondition{}
	}
	base.InternalCondition.ProviderStatus = "400"
	base.InternalCondition.ProviderStatusDescription = errorMessage
}

// buildBookTpaExtensions construye tpaExtensions para la respuesta Book: si la petición trae tpaExtension (p. ej. de PreBook), se reutiliza y se actualiza OwnReference, ClientReference y SupplierReference en servicios.
func buildBookTpaExtensions(req *domain.BookRequest, bookResp BookResponse, creatorID, clientRef string, roomStaysCount int) *common_domain.TpaExtensions {
	services, adultsTotal, childrenTotal, distribution := buildBookServicesFromRequest(req, bookResp)
	servicesCount := len(services)
	if servicesCount == 0 {
		servicesCount = roomStaysCount
	}

	var reqTpaExt []common_domain.TpaExtension
	if len(req.HotelReservations.HotelReservation) > 0 && len(req.HotelReservations.HotelReservation[0].RoomStays.RoomStay) > 0 {
		reqTpaExt = req.HotelReservations.HotelReservation[0].RoomStays.RoomStay[0].RoomRates.RoomRate.TpaExtensions.TpaExtension
	}
	if len(reqTpaExt) > 0 {
		// Copia profunda vía JSON para no modificar el request
		data, err := json.Marshal(reqTpaExt)
		if err == nil {
			var copied []common_domain.TpaExtension
			if err := json.Unmarshal(data, &copied); err == nil && len(copied) > 0 {
				ext := &copied[0]
				if ext.File == nil {
					ext.File = &common_domain.TpaFile{}
				}
				ext.File.OwnReference = bookResp.ClientConfirmationCode
				if bookResp.ClientConfirmationCode == "" {
					ext.File.OwnReference = bookResp.HTConfirmationCode
				}
				ext.File.ClientReference = &clientRef
				ext.File.RoomsTotal = servicesCount
				ext.File.ServiceTotal = servicesCount
				if adultsTotal > 0 {
					ext.File.AdultsTotal = adultsTotal
				}
				if childrenTotal > 0 {
					ext.File.ChildrenTotal = childrenTotal
				}
				if distribution != "" {
					ext.File.Distribution = distribution
				}
				if ext.File.SaleChannel == nil {
					ext.File.SaleChannel = &common_domain.TpaSaleChannel{
						ID:   req.Pos.Source.BookingChannel.Code,
						Name: req.Pos.Source.BookingChannel.Name,
					}
				}
				if (ext.File.Services == nil || len(ext.File.Services.Service) == 0) && len(services) > 0 {
					ext.File.Services = &common_domain.TpaServices{
						Service: services,
					}
				}
				supplierRef := bookResp.HTConfirmationCode
				if ext.File.Services != nil && len(ext.File.Services.Service) > 0 {
					for i := range ext.File.Services.Service {
						if ext.File.Services.Service[i].SupplierReference == nil || *ext.File.Services.Service[i].SupplierReference == "" {
							ext.File.Services.Service[i].SupplierReference = &supplierRef
						}
					}
				}
				return &common_domain.TpaExtensions{
					InvoiceReference: nil,
					TpaExtension:     copied,
				}
			}
		}
	}
	// Fallback: tpaExtensions mínima
	return &common_domain.TpaExtensions{
		InvoiceReference: nil,
		TpaExtension: []common_domain.TpaExtension{
			{
				Client: &common_domain.TpaClient{ID: strPtr(""), Name: strPtr("")},
				File: &common_domain.TpaFile{
					OwnReference:    bookResp.ClientConfirmationCode,
					ClientReference: &clientRef,
					AdultsTotal:     adultsTotal,
					ChildrenTotal:   childrenTotal,
					Distribution:    distribution,
					RoomsTotal:      servicesCount,
					ServiceTotal:    servicesCount,
					SaleChannel: &common_domain.TpaSaleChannel{
						ID:   req.Pos.Source.BookingChannel.Code,
						Name: req.Pos.Source.BookingChannel.Name,
					},
					Services: &common_domain.TpaServices{
						Service: services,
					},
				},
				InternalUser: &common_domain.TpaInternalUser{ID: strPtr(""), MessagePassword: strPtr("")},
			},
		},
	}
}

func buildBookServicesFromRequest(req *domain.BookRequest, bookResp BookResponse) ([]common_domain.TpaService, int, int, string) {
	services := []common_domain.TpaService{}
	totalAdults := 0
	totalChildren := 0
	distributionParts := []string{}
	roomIdx := 0

	for _, hr := range req.HotelReservations.HotelReservation {
		for _, rs := range hr.RoomStays.RoomStay {
			var bc *bookingcode.InternalBookingCode
			openBookingCode := ""
			if s, ok := rs.RoomRates.RoomRate.OpenBookingCode.(string); ok && s != "" {
				openBookingCode = s
				tmp := &bookingcode.InternalBookingCode{}
				tmp.Deserialize(s)
				bc = tmp
			}

			var room *BookRoomResponse
			if roomIdx < len(bookResp.Rooms) {
				room = &bookResp.Rooms[roomIdx]
			}

			tpaService := common_domain.TpaService{
				Identifier:    roomIdx,
				NumberOfRooms: 1,
			}

			if bc != nil {
				tpaService.AccommodationID = bc.GiHotelCode
				tpaService.SupplierAccommodationID = bc.PrvHotelCode
				tpaService.BookingCode = openBookingCode
				tpaService.Market = bc.Market
				tpaService.Nationality = bc.CustomerCountry
				tpaService.BuyCurrency = bc.BuyCurrency
				tpaService.SaleCurrency = bc.SaleCurrency
				tpaService.AdultAges = formatAges(bc.AdultAges)
				tpaService.ChildrensAges = formatAges(bc.ChildrenAges)
				tpaService.NumberOfAdults = bc.Adults
				tpaService.NumberOfChildren = bc.Children + bc.Infant
				tpaService.Distribution = bc.Distribution
				tpaService.StayDateRange = &common_domain.TpaStayRange{
					Start: bc.CheckInDate.Format("2006-01-02"),
					End:   bc.CheckOutDate.Format("2006-01-02"),
				}
				tpaService.BuyChannel = &common_domain.TpaSaleChannel{
					ID:   bc.ProviderCode,
					Name: "",
				}

				roomName := firstNonEmpty(bc.GIRoomName, bc.PrvRoomName)
				if room != nil && room.RoomName != "" {
					roomName = room.RoomName
				}
				tpaService.AccommodationName = roomName
				amount := float64(bc.BuyPrice)
				if amount <= 0 {
					amount = float64(bc.BuyTotalPrice)
				}
				if amount > 0 {
					tpaService.BuyAmount = &amount
					tpaService.SaleAmount = &amount
				}

				roomID := bc.GIRoomID
				if roomID == "" {
					roomID = strconv.Itoa(roomIdx + 1)
				}
				board := &common_domain.TpaBoard{
					ID:   bc.BoardId,
					Name: bc.BoardId,
				}
				tpaRoom := common_domain.TpaRoom{
					AdultsAges:   tpaService.AdultAges,
					Board:        board,
					ChildrenAges: tpaService.ChildrensAges,
					Distribution: tpaService.Distribution,
					// ID:               roomID,
					// Identifier:       roomIdx + 1,
					// Name:             roomName,
					NumberOfAdults:   tpaService.NumberOfAdults,
					NumberOfChildren: tpaService.NumberOfChildren,
					// SupplierName:     firstNonEmpty(bc.PrvRoomName, roomName),
					// SupplierRoomID:   bc.PrvRoomCode,
				}

				if len(bc.GIRoomID) > 0 {
					tpaRoom.ID = bc.GIRoomID
					rmID, _ := strconv.Atoi(bc.GIRoomID)
					tpaRoom.Identifier = rmID
				}
				tpaRoom.SupplierName = bc.PrvRoomName
				tpaRoom.SupplierRoomID = bc.PrvRoomCode
				tpaRoom.Code = bc.GIRoomCode
				tpaRoom.Name = bc.GIRoomName
				tpaRoom.Language = bc.Lang

				tpaService.Rooms = &common_domain.TpaRooms{Rooms: []common_domain.TpaRoom{tpaRoom}}
			}

			if room != nil {
				if room.RoomName != "" {
					tpaService.AccommodationName = room.RoomName
				}
				amount := room.Rates.GrossPrice
				if amount > 0 {
					tpaService.BuyAmount = &amount
					tpaService.SaleAmount = &amount
				}
				if room.Rates.CurrencyCode != "" {
					tpaService.BuyCurrency = room.Rates.CurrencyCode
					tpaService.SaleCurrency = bc.SaleCurrency
				}
				tpaService.NonRefundableRate = !room.Refundable
				supplierRef := bookResp.HTConfirmationCode
				if supplierRef != "" {
					tpaService.SupplierReference = &supplierRef
				}
			}

			if tpaService.Distribution != "" {
				distributionParts = append(distributionParts, tpaService.Distribution)
			}
			totalAdults += tpaService.NumberOfAdults
			totalChildren += tpaService.NumberOfChildren
			services = append(services, tpaService)
			roomIdx++
		}
	}

	return services, totalAdults, totalChildren, strings.Join(distributionParts, "")
}

func strPtr(s string) *string {
	return &s
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

func convertPolicyWindowToSpain(policy HtCancellationPolicy) (string, string) {
	locSpain, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		return normalizeDate(policy.StartWindowTime), normalizeDate(policy.EndWindowTime)
	}

	locProvider, err := time.LoadLocation(policy.TimeZone)
	if err != nil {
		locProvider = time.FixedZone(policy.TimeZone, parseOffsetSeconds(policy.TimeZoneUTC))
	}

	startProvider, err := time.ParseInLocation("2006-01-02 15:04:05", policy.StartWindowTime, locProvider)
	if err != nil {
		return normalizeDate(policy.StartWindowTime), normalizeDate(policy.EndWindowTime)
	}

	endProvider, err := time.ParseInLocation("2006-01-02 15:04:05", policy.EndWindowTime, locProvider)
	if err != nil {
		return normalizeDate(policy.StartWindowTime), normalizeDate(policy.EndWindowTime)
	}

	return startProvider.In(locSpain).Format("2006-01-02"), endProvider.In(locSpain).Format("2006-01-02")
}
