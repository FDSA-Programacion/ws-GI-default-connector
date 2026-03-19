package hoteltrader

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
	"ws-int-httr/internal/domain"
	common_domain "ws-int-httr/internal/domain/gi_response_common"
	log_domain "ws-int-httr/internal/domain/log_domain"
	"ws-int-httr/internal/infrastructure"
	bookingcode "ws-int-httr/internal/infrastructure/booking_code"
	"ws-int-httr/internal/infrastructure/config"
	"ws-int-httr/internal/infrastructure/persistence"
	"ws-int-httr/internal/infrastructure/persistence/orm"
	"ws-int-httr/internal/infrastructure/registry"
	"ws-int-httr/internal/infrastructure/session"
)

// ===================================================================
// TYPES (structs auxiliares del mapper)
// ===================================================================

// HotelMappingResult contiene el resultado de mapear hoteles Proveedor a GI
type HotelMappingResult struct {
	RoomStayGroups []domain.GiRoomStayGroup
	AvailList      string
	NoAvailList    string
}

// occupancyInfo contiene información de ocupación de una habitación
type occupancyInfo struct {
	Adults           int
	Children         int
	Infants          int
	NumberOfRooms    int
	NumberOfService  int
	Distribution     string
	AdultAges        string
	ChildrenAges     string
	RateDistribution string
	RateChildAges    string
}

// optionMappingContext contiene el contexto necesario para mapear una option
type optionMappingContext struct {
	Hotel            PropertyResponseEntity
	HotelCount       int
	MealPlan         string
	Room             RoomResponse // Room de HTTR (incluye mealplan y rate)
	CachedHotel      *orm.DBAlojamiento
	GiBoard          *orm.DBRegimen
	GiHotelCode      string
	InvBlockCode     string
	NonRefundable    string
	NonRefundableInt int
	IntId            int
	PrimaryLangID    string
	MarketCode       string
	CheckInDate      string
	BookingCodeBase  *bookingcode.InternalBookingCode
	GuestCountsMap   map[string][]domain.RSGuestCount
	OccupancyInfoMap map[string]occupancyInfo
	RequestContext   map[string]interface{}
	RoomCounter      *int
}

// ===================================================================
// FUNCIONES PRINCIPALES - RQ (Domain -> GraphQL)
// ===================================================================

// GIAvailRQToOTAvailRQ transforma una petición de disponibilidad GI a formato Proveedor
func GIAvailRequestToProvider(giDomainReq *domain.AvailRequest, cfg config.ProviderConfig) ProviderAvailRQ {
	sessionCtx := session.FromContext()

	// Inicializar context data en la sesión
	if sessionCtx != nil {
		sessionCtx.Set("ContextData", make(map[string]interface{}))
	}

	// Inicializar AvailLog
	providerId := giDomainReq.InternalCondition.Channels.Channel[0].ID
	providerCode := giDomainReq.InternalCondition.Channels.Channel[0].Code
	availLog := &log_domain.AvailLog{
		ProviderID:         strconv.Itoa(providerId),
		ProviderCode:       providerCode,
		Integration:        "WS-INT-HTTR",
		Node:               "",
		EchoToken:          giDomainReq.InternalCondition.CallCondition.EchoToken,
		RqType:             "OTA_HotelAvail",
		ClientName:         giDomainReq.Pos.Source.RequestorID.CompanyName.CompanyShortName,
		ClientCode:         giDomainReq.InternalCondition.ClientCondition.Code,
		RequestorID:        giDomainReq.Pos.Source.RequestorID.ID,
		BookingChannel:     giDomainReq.Pos.Source.BookingChannel.Code,
		RqTimestamp:        giDomainReq.InternalCondition.CallCondition.OriginTimeStamp,
		PrimaryLangID:      giDomainReq.PrimaryLangID,
		StayDateRangeStart: giDomainReq.AvailRequestSegments.AvailRequestSegment.StayDateRange.Start,
		StayDateRangeEnd:   giDomainReq.AvailRequestSegments.AvailRequestSegment.StayDateRange.End,
		Version:            giDomainReq.Version,
		Market:             giDomainReq.AvailRequestSegments.AvailRequestSegment.Tpa.Market,
		Nationality:        giDomainReq.AvailRequestSegments.AvailRequestSegment.Tpa.Nationality,
		RqInternal:         "",
		RqProvider:         "",
		IsRebook:           false,
		RqHotelCodeList:    []string{},
		RqPrvHotelCodeList: []string{},
		SentToSupplier:     false,
	}

	// Guardar RqInternal solo si debug == "g2018i"
	if strings.EqualFold(giDomainReq.Debug, infrastructure.DEBUG_PASSWORD) {
		if rqBytes, err := json.Marshal(giDomainReq); err == nil {
			availLog.RqInternal = string(rqBytes)
		}
	}
	// Si no hay debug, dejar RqInternal vacío
	// availLog.RqInternal = ""
	if availLog.Market == "" {
		availLog.Market = giDomainReq.AvailRequestSegments.AvailRequestSegment.TpaExtension.Market
	}
	if availLog.Nationality == "" {
		availLog.Nationality = giDomainReq.AvailRequestSegments.AvailRequestSegment.TpaExtension.Nationality
	}

	giReqAvailRequestSegment := giDomainReq.AvailRequestSegments.AvailRequestSegment

	// generalConf, ok := registry.Get[*config.YamlConfig]("config")
	// if !ok {
	// 	log.Fatal("config not found")
	// }

	sessionCtx.Data().ProviderCode = giDomainReq.InternalCondition.Channels.Channel[0].ID
	sessionCtx.Data().Debug = giDomainReq.Debug

	checkInDate, err := time.Parse("2006-01-02", giReqAvailRequestSegment.StayDateRange.Start)
	if err != nil {
		customErr := domain.ErrorDateParsing
		customErr.Err = fmt.Errorf("fallo al parsear CheckInDate (%s): %v", giReqAvailRequestSegment.StayDateRange.Start, err)
		panic(customErr)
	}
	checkOutDate, err := time.Parse("2006-01-02", giReqAvailRequestSegment.StayDateRange.End)
	if err != nil {
		customErr := domain.ErrorDateParsing
		customErr.Err = fmt.Errorf("fallo al parsear CheckOutDate (%s): %v", giReqAvailRequestSegment.StayDateRange.End, err)
		panic(customErr)
	}

	otherData := map[string]interface{}{}
	otherData["CheckInDate"] = giReqAvailRequestSegment.StayDateRange.Start
	otherData["CheckOutDate"] = giReqAvailRequestSegment.StayDateRange.End
	listCandidates := giReqAvailRequestSegment.RoomStayCandidates.RoomStayCandidateList
	guestCounts, occupancies, guestInfoData := processGuestData(listCandidates, otherData)

	// Mapear destinos y room candidates
	propertyIds := mapDestinations(giReqAvailRequestSegment, availLog)

	// Guardar log en sesión
	sessionCtx.Data().AvailLog = availLog
	sessionCtx.Data().AvailLogStartTime = time.Now()

	config, ok := registry.Get[*config.YamlConfig]("config")
	if !ok {
		panic(domain.ErrorConfigNotFound)
	}

	bookingCodeBase := &bookingcode.InternalBookingCode{
		EchoToken:         giDomainReq.InternalCondition.CallCondition.EchoToken,
		Version:           config.BookingCodeVersion(),
		ProviderCode:      strconv.Itoa(giDomainReq.InternalCondition.Channels.Channel[0].ID),
		CheckInDate:       checkInDate,
		CheckOutDate:      checkOutDate,
		CustomerCountry:   giReqAvailRequestSegment.Tpa.Nationality,
		Market:            giReqAvailRequestSegment.Tpa.Market,
		SaleCurrency:      giReqAvailRequestSegment.Tpa.Currency,
		ExchangeRateID:    0,
		ExchangeRateValue: 0.0,
		MarkupID:          0,
		Markup:            0.0,
	}

	// Construir contexto de sesión
	requestContextData := map[string]interface{}{}
	requestContextData["guestInfoData"] = guestInfoData
	requestContextData["guestCounts"] = guestCounts
	requestContextData["occupancies"] = occupancies

	requestContextData["echoToken"] = giDomainReq.InternalCondition.CallCondition.EchoToken
	requestContextData["checkIn"] = giReqAvailRequestSegment.StayDateRange.Start
	requestContextData["checkOut"] = giReqAvailRequestSegment.StayDateRange.End
	requestContextData["providerCode"] = giDomainReq.InternalCondition.Channels.Channel[0].ID
	requestContextData["tpaCurrency"] = giReqAvailRequestSegment.Tpa.Currency
	requestContextData["tpaNationality"] = giReqAvailRequestSegment.Tpa.Nationality
	requestContextData["tpaMarket"] = giReqAvailRequestSegment.Tpa.Market
	requestContextData["primaryLangID"] = giDomainReq.PrimaryLangID
	requestContextData["occupancyInfo"] = _buildOccupancyInfo(giReqAvailRequestSegment.RoomStayCandidates.RoomStayCandidateList)

	requestContextData["provider"] = map[string]interface{}{
		"id":   giDomainReq.InternalCondition.Channels.Channel[0].ID,
		"code": giDomainReq.InternalCondition.Channels.Channel[0].Code,
	}

	contextData := map[string]interface{}{}
	contextData["Request"] = requestContextData
	contextData["bookingCodeBase"] = bookingCodeBase
	contextData["numOccupancies"] = strconv.Itoa(len(occupancies))

	sessionCtx.Data().ContextData = contextData

	return ProviderAvailRQ{
		Query: GetPropertiesByIdsQuery,
		Variables: SearchCriteriaVariables{
			SearchCriteriaByIds: &SearchCriteriaByIdsInput{
				PropertyIds: propertyIds,
				Occupancies: occupancies,
			},
		},
	}
}

// ===================================================================
// FUNCIONES PRINCIPALES - RS (XML -> Domain)
// ===================================================================

// OTAvailRSToGIAvailRS transforma una respuesta de disponibilidad Proveedor a formato GI
func ProviderAvailResponseToGI(graphqlResp *ProviderAvailRS, req *domain.AvailRequest) *domain.BaseJsonRS[*domain.AvailResponse] {
	sessionCtx := session.FromContext()
	contextData := sessionCtx.Data().ContextData
	requestContext := contextData["Request"].(map[string]interface{})

	intId := sessionCtx.Data().ProviderCode

	// Obtener log de la sesión
	availLog := sessionCtx.Data().AvailLog
	if availLog == nil {
		panic(domain.ErrorAvailLogNotFound)
	}
	// Obtener tiempo de inicio (desde que entra la petición)
	startTime := sessionCtx.Data().StartTime
	if startTime.IsZero() {
		// Fallback: si no hay StartTime, usar AvailLogStartTime
		startTime = sessionCtx.Data().AvailLogStartTime
		if startTime.IsZero() {
			startTime = time.Now()
		}
	}

	// Obtener datos del HTTP response del proveedor desde métricas de sesión
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

	// Procesar respuesta proveedor
	hotels := graphqlResp.Data.GetPropertiesByIds.Properties

	// Calcular estadísticas del proveedor
	supplierNumHotels := len(hotels)
	supplierNumRooms := 0
	supplierNumRates := 0
	for _, hotel := range hotels {
		supplierNumRooms += len(hotel.Rooms)
		supplierNumRates += len(hotel.Rooms) // Cada room es una rate
	}

	// Obtener configuración para pasar límite de habitaciones
	cfgService, ok := registry.Get[config.AppConfig]("config")
	if !ok {
		panic(domain.ErrorConfigNotFound)
	}

	primaryLangID := requestContext["primaryLangID"].(string)
	marketCode := requestContext["tpaMarket"].(string)
	checkInDate := requestContext["checkIn"].(string)

	bookingCodeBase := contextData["bookingCodeBase"].(*bookingcode.InternalBookingCode)

	// Crear contexto para mapear la option
	ctx := &optionMappingContext{
		IntId:           intId,
		PrimaryLangID:   primaryLangID,
		MarketCode:      marketCode,
		CheckInDate:     checkInDate,
		BookingCodeBase: bookingCodeBase,
		RequestContext:  requestContext,
	}

	// Mapear hoteles a GI
	availResponse := &domain.AvailResponse{}
	mappingResult := mapHotelsToGI(hotels, cfgService, ctx)
	availResponse.GiRoomStayGroup = mappingResult.RoomStayGroups

	// Construir RsList
	rsList := map[string]string{}
	if mappingResult.AvailList != "" {
		rsList["AvailList"] = mappingResult.AvailList
	}
	if mappingResult.NoAvailList != "" {
		rsList["NoAvailList"] = mappingResult.NoAvailList
	}

	domainResp := &domain.BaseJsonRS[*domain.AvailResponse]{
		// EchoToken:      OTAvailResp.ProviderRSs.ProviderRS.RefId,
		PrimaryLangID:  "es",
		SchemaLocation: "http://www.opentravel.org/OTA/2003/05 OTA_HotelAvailRS.xsd",
		Success:        "",
		Version:        "1.004",
		Xsi:            "http://www.w3.org/2001/XMLSchema-instance",
		GiRoomStays:    availResponse,
		InternalCondition: &common_domain.InternalCondition{
			RsList:            rsList,
			Status:            "200",
			StatusDescription: "OK",
		},
	}

	// Completar log y escribirlo
	completeAvailLog(availLog, domainResp, mappingResult, intId, startTime,
		supplierRsTime, supplierRsHttpStatusCode, supplierRsLength, supplierErrorMessage,
		supplierNumHotels, supplierNumRooms, supplierNumRates)
	attachAvailErrors(domainResp, availLog)

	domainResp.Success = ""

	return domainResp
}

// ===================================================================
// HELPERS RQ - Mapeo de Request
// ===================================================================

func mapDestinations(giReqAvailRequestSegment domain.AvailRequestSegment, availLog *log_domain.AvailLog) []string {
	repositoryService, ok := registry.Get[persistence.RepositoryService]("repository")
	if !ok {
		panic(domain.ErrorRepositoryNotFound)
	}

	if availLog == nil {
		panic(domain.ErrorAvailLogNotFound)
	}

	destinations := []string{}
	mapHotelSession := map[string]string{}

	sessionCtx := session.FromContext()
	intId := sessionCtx.Data().ProviderCode

	for i := 0; i < len(giReqAvailRequestSegment.HotelSearchCriteria.Criterion.HotelRef); i++ {
		destRef := giReqAvailRequestSegment.HotelSearchCriteria.Criterion.HotelRef[i]

		if destRef.HotelCode != "" {
			availLog.RqHotelCodeList = append(availLog.RqHotelCodeList, destRef.HotelCode)
			externalCodeStr := repositoryService.GIHotelCodeToExternalCode(destRef.HotelCode, intId)

			if externalCodeStr != "" {
				mapHotelSession[externalCodeStr] = destRef.HotelCode
				destinations = append(destinations, externalCodeStr)
				availLog.RqPrvHotelCodeList = append(availLog.RqPrvHotelCodeList, externalCodeStr)
			}

		} else if destRef.HotelCityCode != "" {
			if availLog.RqCity == "" {
				availLog.RqCity = destRef.HotelCityCode
			}
			externalCodeStrs := repositoryService.GICityCodeToExternalCode(destRef.HotelCityCode, intId)
			for j := 0; j < len(externalCodeStrs); j++ {
				mapHotelSession[externalCodeStrs[j]] = destRef.HotelCityCode
				destinations = append(destinations, externalCodeStrs[j])
				availLog.RqPrvHotelCodeList = append(availLog.RqPrvHotelCodeList, externalCodeStrs[j])
			}
		} else if destRef.AreaID != "" {
			if availLog.RqZone == "" {
				availLog.RqZone = destRef.AreaID
			}
			externalCodeStrs := repositoryService.GIAreaIdToExternalCode(destRef.AreaID, intId)
			for j := 0; j < len(externalCodeStrs); j++ {
				mapHotelSession[externalCodeStrs[j]] = destRef.AreaID
				destinations = append(destinations, externalCodeStrs[j])
				availLog.RqPrvHotelCodeList = append(availLog.RqPrvHotelCodeList, externalCodeStrs[j])
			}
		}
	}

	sessionCtx.Data().MapHotelSession["mapHotelSession"] = mapHotelSession

	return destinations
}

// ===================================================================
// HELPERS RS - Mapeo de Response (Hoteles)
// ===================================================================

type RoomsByOccupancy map[int][]RoomResponse
type RoomsByMealplan map[string]RoomsByOccupancy

type RoomClassification struct {
	Refundables    RoomsByMealplan
	NonRefundables RoomsByMealplan
}

// mapHotelsToGI mapea la lista de hoteles Proveedor a grupos de RoomStay GI
func mapHotelsToGI(hotels []PropertyResponseEntity, cfg config.ProviderConfig, ctx *optionMappingContext) HotelMappingResult {
	sessionCtx := session.FromContext()
	giRoomStayGroups := []domain.GiRoomStayGroup{}
	availList := ""
	noAvailList := ""

	repositoryService, ok := registry.Get[persistence.RepositoryService]("repository")
	if !ok {
		panic(domain.ErrorRepositoryNotFound)
	}

	mapHotelSession := map[string]string{}
	if storedMap, ok := sessionCtx.Data().MapHotelSession["mapHotelSession"]; ok {
		if casted, ok := storedMap.(map[string]string); ok {
			mapHotelSession = casted
		}
	}

	hotelCount := len(hotels)
	ctx.HotelCount = hotelCount

	for _, hotel := range hotels {
		cachedHotel := repositoryService.GiHotelFromExternalCode(strconv.Itoa(hotel.PropertyID), ctx.IntId)

		// Hotel no mapeado - excluir
		if cachedHotel == nil {
			noAvailList = noAvailList + strconv.Itoa(hotel.PropertyID) + ","
			continue
		}

		availList = availList + strconv.Itoa(hotel.PropertyID) + ","

		roomClassification := RoomClassification{
			Refundables:    make(RoomsByMealplan),
			NonRefundables: make(RoomsByMealplan),
		}

		// Obtener GI hotel code
		giHotelCode := resolveGiHotelCode(cachedHotel, mapHotelSession, strconv.Itoa(hotel.PropertyID))

		// Crear contexto para mapear la option
		ctx.Hotel = hotel
		ctx.CachedHotel = cachedHotel
		ctx.GiHotelCode = giHotelCode

		for _, room := range hotel.Rooms {
			mealplan := room.MealplanOptions.MealplanCode

			// Excluir si el board no está mapeado
			giBoard := repositoryService.GetBoardFromExternalCode(mealplan)
			if giBoard == nil {
				continue
			}

			occ := room.OccupancyRefID

			var target RoomsByMealplan
			if room.Refundable {
				target = roomClassification.Refundables
			} else {
				target = roomClassification.NonRefundables
			}

			if _, ok := target[giBoard.Codigo]; !ok {
				target[giBoard.Codigo] = make(RoomsByOccupancy)
			}

			target[giBoard.Codigo][occ] = append(
				target[giBoard.Codigo][occ],
				room,
			)
		}

		// Limpiamos las combinaciones que NO tengan una de las occupaciones.
		cleanEmptyCombinations(roomClassification)
		// Reducimos las combinaciones a las más baratas.
		reduceToCheapestRooms(roomClassification, cfg)

		giRoomStayGroup := domain.GiRoomStayGroup{
			HotelCode: ctx.GiHotelCode,
			Supplier:  strconv.Itoa(ctx.IntId),
		}

		basicPropertyInfo := createBasicPropertyInfo(cachedHotel)

		buildCombinations(roomClassification, &giRoomStayGroup, &giRoomStayGroups, basicPropertyInfo, ctx)

	}

	// Obtener contexto de sesión
	contextData := sessionCtx.Data().ContextData
	if contextData == nil {
		panic(domain.ErrorContextDataNotFound)
	}

	return HotelMappingResult{
		RoomStayGroups: giRoomStayGroups,
		AvailList:      "",
		NoAvailList:    "",
	}
}

func reduceToCheapestRooms(roomClassification RoomClassification, cfg config.ProviderConfig) {
	maxRoomsPerOccupancy := cfg.ProviderMaxRoomsPerOccupancy()
	if maxRoomsPerOccupancy <= 0 {
		maxRoomsPerOccupancy = 4 // Por defecto serán 4 si no se especifica
	}

	for _, byMealplan := range []RoomsByMealplan{
		roomClassification.Refundables,
		roomClassification.NonRefundables,
	} {
		for mealplan, byOccupancy := range byMealplan {
			for occ, rooms := range byOccupancy {
				byOccupancy[occ] = cheapestRooms(rooms, maxRoomsPerOccupancy)
			}
			byMealplan[mealplan] = byOccupancy
		}
	}
}

func cheapestRooms(rooms []RoomResponse, limit int) []RoomResponse {
	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i].RateInfo.GrossPrice < rooms[j].RateInfo.GrossPrice
	})

	if len(rooms) > limit {
		return rooms[:limit]
	}

	return rooms
}

func cleanEmptyCombinations(roomClassification RoomClassification) {
	sessionCtx := session.FromContext()
	data := sessionCtx.Data().ContextData

	numOccupanciesStr, ok := data["numOccupancies"].(string)
	if !ok {
		customErr := domain.ErrorDataConversion
		customErr.Err = fmt.Errorf("numOccupancies not found in cleanEmptyCombinations")
		panic(customErr)
	}

	numOccupancies, err := strconv.Atoi(numOccupanciesStr)
	if err != nil || numOccupancies == 0 {
		customErr := domain.ErrorDataConversion
		customErr.Err = fmt.Errorf("numOccupancies is not a valid integer in cleanEmptyCombinations: %v", err)
		panic(customErr)
	}

	// Limpia refundables
	for mealplan, occMap := range roomClassification.Refundables {
		if len(occMap) != numOccupancies {
			delete(roomClassification.Refundables, mealplan)
		}
	}

	// Limpia non-refundables
	for mealplan, occMap := range roomClassification.NonRefundables {
		if len(occMap) != numOccupancies {
			delete(roomClassification.NonRefundables, mealplan)
		}
	}
}

func createBasicPropertyInfo(hotel *orm.DBAlojamiento) domain.BasicPropertyInfo {
	return domain.BasicPropertyInfo{
		HotelCode:     hotel.HotelCode,
		HotelName:     hotel.HotelName,
		HotelCityCode: strconv.Itoa(hotel.CityID),

		Address: domain.Address{
			CityName: hotel.CityName,
		},
		Award: domain.Award{
			HotelCategory:     hotel.Category,
			PropertyClassCode: hotel.PropertyClassCode,
			Rating:            hotel.Rating,
		},
	}
}

type RoomCombination []RoomResponse

func buildCombinations(
	roomClassification RoomClassification,
	giRoomStayGroup *domain.GiRoomStayGroup,
	giRoomStayGroups *[]domain.GiRoomStayGroup,
	basicPropertyInfo domain.BasicPropertyInfo,
	ctx *optionMappingContext,
) {
	// Obtener contexto de sesión
	sessionCtx := session.FromContext()
	contextData := sessionCtx.Data().ContextData

	if contextData == nil {
		panic(domain.ErrorContextDataNotFound)
	}

	for refundable, byMealplan := range map[bool]RoomsByMealplan{
		true:  roomClassification.Refundables,
		false: roomClassification.NonRefundables,
	} {

		for giMealplan, byOccupancy := range byMealplan {
			ctx.MealPlan = giMealplan
			if !refundable {
				ctx.NonRefundable = "1"
				ctx.NonRefundableInt = 1
			} else {
				ctx.NonRefundable = "0"
				ctx.NonRefundableInt = 0
			}

			// ordenamos las occupancies
			occupancies := make([]int, 0, len(byOccupancy))
			for occ := range byOccupancy {
				occupancies = append(occupancies, occ)
			}
			sort.Ints(occupancies)

			// generamos las combinaciones
			combinations := []RoomCombination{}
			combineByOccupancy(
				occupancies,
				byOccupancy,
				0,
				[]RoomResponse{},
				&combinations,
			)

			giRoomStayGroup.BoardCode = giMealplan
			if refundable {
				giRoomStayGroup.NonRefund = "0"
			} else {
				giRoomStayGroup.NonRefund = "1"
			}

			// mapeamos a GiRoomStayGroup
			for _, combo := range combinations {
				*giRoomStayGroups = append(
					*giRoomStayGroups,
					completeGiRoomStayGroup(giRoomStayGroup, combo, basicPropertyInfo, ctx),
				)
			}
		}
	}
}

func combineByOccupancy(
	occupancies []int,
	roomsByOcc map[int][]RoomResponse,
	index int,
	current []RoomResponse,
	result *[]RoomCombination,
) {
	if index == len(occupancies) {
		comb := make(RoomCombination, len(current))
		copy(comb, current)
		*result = append(*result, comb)
		return
	}

	occ := occupancies[index]
	for _, room := range roomsByOcc[occ] {
		combineByOccupancy(
			occupancies,
			roomsByOcc,
			index+1,
			append(current, room),
			result,
		)
	}
}

func completeGiRoomStayGroup(
	giRoomStayGroup *domain.GiRoomStayGroup,
	rooms []RoomResponse,
	basicPropertyInfo domain.BasicPropertyInfo,
	ctx *optionMappingContext,
) domain.GiRoomStayGroup {

	var total float64
	stays := make([]*domain.GiRoomStay, 0, len(rooms))

	sessionCtx := session.FromContext()
	contextData := sessionCtx.Data().ContextData
	requestContext := contextData["Request"].(map[string]interface{})
	marketCode := requestContext["tpaMarket"].(string)

	bookingCodeBase := contextData["bookingCodeBase"].(*bookingcode.InternalBookingCode)
	guestCountsMap := requestContext["guestCounts"].(map[string]domain.GuestCounts)
	guestInfoData := requestContext["guestInfoData"].(GuestInfoData)

	repositoryService, ok := registry.Get[persistence.RepositoryService]("repository")
	if !ok {
		panic(domain.ErrorRepositoryNotFound)
	}

	// Primera room
	firstRoom := rooms[0]

	// Para calcular el total.
	for _, room := range rooms {
		total += float64(room.RateInfo.GrossPrice)
	}

	// Importante para multi-habitacion:
	// todas las habitaciones de la misma combinacion deben compartir InvBlockCode.
	groupInvBlockCode := getInvBlockCode(1, strconv.Itoa(ctx.IntId), ctx.GiHotelCode, giRoomStayGroup.BoardCode, ctx.NonRefundable)

	for _, room := range rooms {
		giRoom := repositoryService.GetRoomFromExternalCode(room.RoomCode, room.RoomName, ctx.PrimaryLangID, ctx.IntId)
		giBoard := repositoryService.GetBoardFromExternalCode(room.MealplanOptions.MealplanCode)
		if giBoard == nil {
			continue
		}

		bookingCode := *bookingCodeBase

		bookingCode.GiHotelCode = ctx.GiHotelCode
		bookingCode.PrvHotelCode = strconv.Itoa(ctx.Hotel.PropertyID)
		bookingCode.GIRoomID = giRoom.GIRoomID
		bookingCode.GIRoomCode = giRoom.GIRoomCode
		bookingCode.GIRoomName = giRoom.GIRoomName

		// currentPrvRoomCode, curentPrvRoomName := getPrvRoomCodeAndName(provRoom.Description)
		bookingCode.PrvRoomCode = room.RoomCode
		bookingCode.PrvRoomName = room.RoomName
		// bookingCode.IsDynamicRoom = 0
		bookingCode.BuyCurrency = room.RateInfo.Currency
		bookingCode.BuyPrice = float32(room.RateInfo.GrossPrice)
		bookingCode.BuyTotalPrice = float32(total)
		bookingCode.NrRate = ctx.NonRefundableInt
		bookingCode.BoardId = giBoard.Codigo
		bookingCode.Lang = strings.ToUpper(ctx.PrimaryLangID)

		// Guardar el RoomCandidateRefID en el campo Id para usarlo en Book
		bookingCode.Id = strconv.Itoa(room.OccupancyRefID)

		roomsCount := len(rooms)

		bookingCode.NumberOfRooms = roomsCount
		bookingCode.NumberOfService = roomsCount

		// Usar ocupación agregada
		guestRoomInfoData := guestInfoData.infoByRoom[strconv.Itoa(room.OccupancyRefID)]

		// ExtraParams - Construir objeto Quote para PreBook (guardar en base64 para evitar que : o @@@ rompan MapToString)
		quoteData := map[string]interface{}{
			"htIdentifier": room.HTIdentifier,
			"occupancy": map[string]interface{}{
				"guestAges": guestRoomInfoData.roomGuestAges, // Formato: "30,30"
			},
			"rates": map[string]interface{}{
				"netPrice":      room.RateInfo.NetPrice,
				"tax":           room.RateInfo.Tax,
				"grossPrice":    room.RateInfo.GrossPrice,
				"payAtProperty": room.RateInfo.PayAtProperty,
			},
		}

		quoteJSON, err := json.Marshal(quoteData)
		if err != nil {
		}

		extraParams := map[string]string{
			"quote": base64.StdEncoding.EncodeToString(quoteJSON),
		}
		bookingCode.ExtraParams = bookingcode.MapToString(extraParams)

		bookingCode.Distribution = guestInfoData.distribution
		bookingCode.Adults = guestInfoData.totalNumAdults
		bookingCode.Children = guestInfoData.totalNumChildren
		bookingCode.Infant = guestRoomInfoData.roomNumInfants

		bookingCode.AdultAges = guestInfoData.totalAdultAges
		bookingCode.ChildrenAges = guestInfoData.totalChildrenAges

		bookingCode.RateDistribution = fmt.Sprintf("%d~%d", guestRoomInfoData.roomNumAdults, guestRoomInfoData.roomTotalNonAdults)
		bookingCode.RateChildAges = splitAgesByTilde(guestRoomInfoData.roomChildrenAges)
		bookingCode.InvBlockCode = groupInvBlockCode

		nonRefundable := "1"
		if room.Refundable {
			nonRefundable = "0"
		}

		descriptionTexts := []string{room.RoomName, giBoard.Codigo}
		if consolidatedComments := strings.TrimSpace(room.ConsolidatedComments); consolidatedComments != "" {
			descriptionTexts = append(descriptionTexts, consolidatedComments)
		}

		roomRates := domain.RoomRates{
			RoomRate: domain.RoomRate{
				AvailabilityStatus: infrastructure.AVAILABILITY_STATUS,
				DirectPayment:      infrastructure.DIRECT_PAYMENT,
				InvBlockCode:       groupInvBlockCode,
				OpenBookingCode:    bookingCode.Serialize(),
				RoomRateDescription: domain.RoomRateDescription{
					Text: descriptionTexts,
				},
				RoomTypeCode: room.RoomCode,
				Total: domain.Total{
					NonRefundable: nonRefundable,
					Taxes: domain.Taxes{
						Amount:       roundTo2Decimals(room.RateInfo.GrossPrice),
						CurrencyCode: room.RateInfo.Currency,
					},
				},
			},
		}

		// Construir GiRoomStay y se añade a la lista
		stays = append(stays, &domain.GiRoomStay{
			BasicPropertyInfo:    basicPropertyInfo,
			CancelPenalties:      buildCancelPenalties(room.CancellationPolicies),
			RoomStayCandidateRPH: room.OccupancyRefID,
			GuestCounts:          guestCountsMap[strconv.Itoa(room.OccupancyRefID)],
			MarketCode:           marketCode,
			RoomRates:            roomRates,
		})
	}

	giRoomStayGroup.GiRoomStay = stays
	giRoomStayGroup.NettPrice = roundTo2Decimals(total)
	giRoomStayGroup.Price = roundTo2Decimals(total)
	giRoomStayGroup.RoomCode = firstRoom.RoomCode

	giRoomStayGroup.Key = composeKey(
		giRoomStayGroup.HotelCode,
		giRoomStayGroup.RoomCode,
		giRoomStayGroup.BoardCode,
		giRoomStayGroup.NonRefund,
	)

	return *giRoomStayGroup
}

func buildCancelPenalties(provCancelPolicies []HtCancellationPolicy) common_domain.CancelPenalties {
	cancelPenaltiesArr := []common_domain.CancelPenalty{}

	for _, provCancelPolicy := range provCancelPolicies {
		startSpain, endSpain := convertPolicyWindowToSpain(provCancelPolicy)

		cancelPenaltiesArr = append(cancelPenaltiesArr, common_domain.CancelPenalty{
			Start: startSpain,
			End:   endSpain,
			AmountPercent: common_domain.AmountPercent{
				CurrencyCode: provCancelPolicy.Currency,
				Amount:       fmt.Sprintf("%.2f", provCancelPolicy.CancellationCharge),
			},
		})
	}

	return common_domain.CancelPenalties{
		CancelPenalty: cancelPenaltiesArr,
	}
}

func splitAgesByTilde(s string) string {
	if len(s)%2 != 0 {
		return s // o manejar error si quieres
	}

	parts := make([]string, 0, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		parts = append(parts, s[i:i+2])
	}

	return strings.Join(parts, "~")
}

// ===================================================================
// HELPERS RS - Ocupación
// ===================================================================

type GuestInfoData struct {
	totalNumAdults    int
	totalNumChildren  int
	totalNumInfants   int
	totalNonAdults    int
	totalGuest        int
	totalAges         string
	totalAdultAges    string
	totalChildrenAges string
	distribution      string
	infoByRoom        map[string]GuestRoomInfoData
}
type GuestRoomInfoData struct {
	roomNumAdults      int
	roomNumChildren    int
	roomNumInfants     int
	roomTotalNonAdults int
	roomTotalGuest     int
	roomTotalAges      string
	roomAdultAges      string
	roomChildrenAges   string
	roomDistribution   string
	roomGuestAges      string // Formato: "30,30" (con comas)
}

func processGuestData(listCandidates []domain.RoomStayCandidate, otherData map[string]interface{}) (map[string]domain.GuestCounts, []OccupancyInput, GuestInfoData) {
	occupancies := []OccupancyInput{}
	guestCounts := map[string]domain.GuestCounts{}

	// Extraer datos genéricos que cada proveedor puede pasar de forma diferente
	checkInDate, _ := otherData["CheckInDate"].(string)
	checkOutDate, _ := otherData["CheckOutDate"].(string)

	guestInfoData := GuestInfoData{
		totalNumAdults:    0,
		totalNumChildren:  0,
		totalNumInfants:   0,
		totalNonAdults:    0,
		totalGuest:        0,
		totalAges:         "",
		totalAdultAges:    "",
		totalChildrenAges: "",
		infoByRoom:        map[string]GuestRoomInfoData{},
	}

	for _, listGuest := range listCandidates {
		responseCandidateCount := []domain.RSGuestCount{}

		guestRoomInfoData := GuestRoomInfoData{
			roomNumAdults:      0,
			roomNumChildren:    0,
			roomNumInfants:     0,
			roomTotalNonAdults: 0,
			roomTotalGuest:     0,
			roomTotalAges:      "",
			roomAdultAges:      "",
			roomChildrenAges:   "",
			roomDistribution:   "",
			roomGuestAges:      "",
		}

		// Se genera el GuestCount que se retorna a GI.
		// sum := map[string]int{"10": 0, "8": 0, "7": 0}
		sum := map[string]int{}
		allAges := map[string]string{}
		// Se genera para el proveedor.
		guestAges := []string{}

		for _, guest := range listGuest.GuestCounts.GuestCountList {
			sum[guest.AgeQualifyingCode]++
			allAges[guest.AgeQualifyingCode] = allAges[guest.AgeQualifyingCode] + CastTo2Digits(strconv.Itoa(guest.Age))
			guestAges = append(guestAges, CastTo2Digits(strconv.Itoa(guest.Age)))
		}

		for ageQualifyingCode, count := range sum {
			guestCount := domain.RSGuestCount{
				AgeQualifyingCode: ageQualifyingCode,
				Count:             count,
			}

			responseCandidateCount = append(responseCandidateCount, guestCount)
		}

		guestRoomInfoData.roomAdultAges = allAges["10"]
		guestRoomInfoData.roomChildrenAges = allAges["8"] + allAges["7"]
		guestRoomInfoData.roomTotalAges = allAges["10"] + allAges["8"] + allAges["7"]

		guestInfoData.totalAdultAges += allAges["10"]
		guestInfoData.totalChildrenAges += allAges["8"] + allAges["7"]
		guestInfoData.totalAges += allAges["10"] + allAges["8"] + allAges["7"]

		guestCounts[listGuest.Rph] = domain.GuestCounts{
			GuestCount: responseCandidateCount,
		}

		guestRoomInfoData.roomNumAdults = sum["10"]
		guestRoomInfoData.roomNumChildren = sum["8"]
		guestRoomInfoData.roomNumInfants = sum["7"]
		guestRoomInfoData.roomTotalNonAdults = guestRoomInfoData.roomNumChildren + guestRoomInfoData.roomNumInfants
		guestRoomInfoData.roomTotalGuest = guestRoomInfoData.roomNumAdults + guestRoomInfoData.roomTotalNonAdults
		guestRoomInfoData.roomDistribution = fmt.Sprintf("%d%d", guestRoomInfoData.roomNumAdults, guestRoomInfoData.roomTotalNonAdults)
		guestRoomInfoData.roomTotalAges = strings.Join(guestAges, "")
		guestRoomInfoData.roomGuestAges = strings.Join(guestAges, ",") // Para PreBook: "30,30"

		guestInfoData.infoByRoom[listGuest.Rph] = guestRoomInfoData

		refId, err := strconv.Atoi(listGuest.Rph)
		if err != nil {
			customErr := domain.ErrorDataConversion
			customErr.Err = fmt.Errorf("error converting OccupancyRefId to int: %v", err)
			panic(customErr)
		}
		// Se genera el occupancy a proveedor.
		occupancies = append(occupancies, OccupancyInput{
			OccupancyRefId: refId,
			CheckInDate:    checkInDate,
			CheckOutDate:   checkOutDate,
			GuestAges:      strings.Join(guestAges, ","),
		})
	}

	// Sumamos lo de todas para tener un resumen general
	for _, roomInfo := range guestInfoData.infoByRoom {
		guestInfoData.totalNumAdults += roomInfo.roomNumAdults
		guestInfoData.totalNumChildren += roomInfo.roomNumChildren
		guestInfoData.totalNumInfants += roomInfo.roomNumInfants
		guestInfoData.totalNonAdults += roomInfo.roomTotalNonAdults
		guestInfoData.totalGuest += roomInfo.roomTotalGuest
		guestInfoData.distribution += roomInfo.roomDistribution
	}

	// guestCounts := domain.GuestCounts{
	// 	GuestCount: ,
	// }

	return guestCounts, occupancies, guestInfoData
}

func _buildOccupancyInfo(list []domain.RoomStayCandidate) map[string]occupancyInfo {
	result := make(map[string]occupancyInfo)

	for _, roomStayCandidate := range list {
		info := occupancyInfo{
			NumberOfRooms:   1,
			NumberOfService: 1,
		}
		adultAgesBuilder := strings.Builder{}
		childAges := []string{}

		for _, guest := range roomStayCandidate.GuestCounts.GuestCountList {
			// Siempre formatear la edad, incluso si es 0 (queda "00")
			ageStr := CastTo2Digits(strconv.Itoa(guest.Age))

			switch guest.AgeQualifyingCode {
			case "10": // Adulto
				info.Adults++
				adultAgesBuilder.WriteString(ageStr)
			case "7": // Infante
				info.Infants++
				childAges = append(childAges, ageStr)
			default: // Niño u otros (incluye "8")
				info.Children++
				childAges = append(childAges, ageStr)
			}
		}

		totalNonAdults := info.Children + info.Infants
		info.Distribution = fmt.Sprintf("%d%d", info.Adults, totalNonAdults)
		info.RateDistribution = fmt.Sprintf("%d~%d", info.Adults, info.Children)
		info.AdultAges = adultAgesBuilder.String()
		info.ChildrenAges = strings.Join(childAges, "")
		info.RateChildAges = strings.Join(childAges, "~")

		result[roomStayCandidate.Rph] = info
	}

	return result
}

// completeAvailLog completa y escribe el log de disponibilidad
func completeAvailLog(availLog *log_domain.AvailLog, domainResp *domain.BaseJsonRS[*domain.AvailResponse],
	mappingResult HotelMappingResult, intId int, startTime time.Time,
	supplierRsTime int64, supplierRsHttpStatusCode int, supplierRsLength int, supplierErrorMessage string,
	supplierNumHotels int, supplierNumRooms int, supplierNumRates int) {

	rsLength := 0

	if jsonBytes, err := json.Marshal(domainResp); err == nil {
		rsLength = len(jsonBytes)
	}

	rsTime := int(time.Since(startTime).Milliseconds())

	// Contar hoteles y roomStays
	hotelMap := make(map[string]bool)
	roomStayCount := 0
	for _, group := range mappingResult.RoomStayGroups {
		if !hotelMap[group.HotelCode] {
			hotelMap[group.HotelCode] = true
		}
		roomStayCount += len(group.GiRoomStay)
	}

	// Completar campos del log
	availLog.RsNumHotels = len(hotelMap)
	availLog.RsNumRoomStay = roomStayCount

	// Determinar Success/Summary/Error según el patrón requerido:
	// - Si hay ERROR real (parsing, HTTP error, etc.): success="KO", summary="KO"
	// - Si todo va bien y hay resultados: success="OK", summary="OK:Result", errorCode=0
	// - Si todo va bien pero NO hay resultados: success="OK", summary="OK:NoResult", errorCode=3
	hasResults := len(hotelMap) > 0

	// Detectar si hay un ERROR real (HTTP error, parsing error, etc.)
	// NO es un error real si:
	// - HTTP status es 200 Y
	// - El mensaje contiene "NO_AVAIL_FOUND" o "PR#204" o "returns 0 results" (casos de no disponibilidad)
	isNoAvailabilityCase := supplierErrorMessage != "" && (strings.Contains(supplierErrorMessage, "NO_AVAIL_FOUND") ||
		strings.Contains(supplierErrorMessage, "PR#204") ||
		strings.Contains(strings.ToLower(supplierErrorMessage), "returns 0 results") ||
		strings.Contains(strings.ToLower(supplierErrorMessage), "no availability"))
	hasRealError := supplierRsHttpStatusCode != 200 || (supplierErrorMessage != "" && !isNoAvailabilityCase)

	var summary string
	var success string
	var errorCode int
	var errorMessage string
	var providerErrorMessage string

	if hasRealError {
		// ERROR real: parsing, HTTP error, etc.
		success = "KO"
		summary = "KO"
		errorCode = 3
		errorMessage = "Error on avail"
		providerErrorMessage = supplierErrorMessage
	} else if hasResults {
		// Todo va bien y hay resultados
		success = "OK"
		summary = "OK:Result"
		errorCode = 0
		errorMessage = ""
		providerErrorMessage = ""
	} else {
		// Todo va bien pero no hay resultados
		success = "OK"
		summary = "OK:NoResult"
		errorCode = 3
		errorMessage = "Error on avail"
		// Si hay mensaje del proveedor con NO_AVAIL_FOUND, usarlo; si no, usar genérico
		if supplierErrorMessage != "" {
			providerErrorMessage = supplierErrorMessage
		} else {
			providerErrorMessage = "PR#NO_AVAIL_FOUND# No availability was found"
		}
	}

	availLog.Success = success
	availLog.Error = "false"
	availLog.ErrorCode = errorCode
	availLog.ErrorMessage = errorMessage
	availLog.InternalMessage = ""
	availLog.RsTime = rsTime
	availLog.SupplierRsTime = supplierRsTime
	availLog.SupplierRsHttpStatusCode = supplierRsHttpStatusCode
	availLog.SupplierRsLength = supplierRsLength
	availLog.SupplierNumHotels = supplierNumHotels
	availLog.SupplierNumRooms = supplierNumRooms
	availLog.SupplierNumRates = supplierNumRates

	// Habilitado para pruebas: guardar RsInternal (respuesta GI completa)
	if rsBytes, err := json.Marshal(domainResp); err == nil {
		availLog.RsInternal = string(rsBytes)
	}

	availLog.SupplierErrorMessage = providerErrorMessage
	availLog.RsLength = rsLength
	availLog.SentToSupplier = true
	availLog.Summary = summary
	availLog.CachedProviderResponse = false

}

func attachAvailErrors(domainResp *domain.BaseJsonRS[*domain.AvailResponse], availLog *log_domain.AvailLog) {
	if availLog == nil || availLog.Success != "KO" {
		return
	}

	errorMessage := availLog.SupplierErrorMessage
	if errorMessage == "" {
		errorMessage = availLog.ErrorMessage
	}
	if errorMessage == "" {
		return
	}

	domainResp.Errors = BuildGIErrorContainer(errorMessage)
	if domainResp.InternalCondition == nil {
		domainResp.InternalCondition = &common_domain.InternalCondition{}
	}
	domainResp.InternalCondition.ProviderStatus = "400"
	domainResp.InternalCondition.ProviderStatusDescription = errorMessage
}

// ===================================================================
// UTILIDADES GENERALES
// ===================================================================

func composeKey(hotelCode string, roomCode string, boardCode string, nonRefundable string) string {
	b := bytes.Buffer{}
	b.WriteString("hc")
	b.WriteString(hotelCode)
	b.WriteString("rc")
	b.WriteString(roomCode)
	b.WriteString("bc")
	b.WriteString(boardCode)
	b.WriteString("nr")
	b.WriteString(nonRefundable)
	return b.String()
}

func normalizeDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	layouts := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t.Format("2006-01-02")
		}
	}

	if len(dateStr) >= 10 {
		return dateStr[:10]
	}

	return dateStr
}

func getInvBlockCode(providerHotelOrder int, providerCode string, hotelCode string, boardCode string, nr string) string {
	randCode := generateRandomString()

	b := bytes.Buffer{}
	b.WriteString(strconv.Itoa(providerHotelOrder))
	b.WriteString(bookingcode.HYPHEN)
	b.WriteString(providerCode)
	b.WriteString(bookingcode.NEGATION)
	b.WriteString(hotelCode)
	b.WriteString(bookingcode.NEGATION)
	b.WriteString(boardCode)
	b.WriteString(bookingcode.NEGATION)
	b.WriteString(nr)
	b.WriteString(bookingcode.NEGATION)
	b.WriteString(randCode)
	return b.String()
}

func generateRandomString() string {
	letterBytes := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	stringLength := 5

	b := make([]byte, stringLength)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func CastTo2Digits(input string) string {
	if len(input) < 2 {
		return "0" + input
	}
	return input
}

// resolveGiHotelCode obtiene el código de hotel GI
func resolveGiHotelCode(cachedHotel *orm.DBAlojamiento, mapHotelSession map[string]string, hotelCode string) string {
	giHotelCode := ""
	if cachedHotel != nil && cachedHotel.HotelCode != "" {
		giHotelCode = cachedHotel.HotelCode
	}
	if giHotelCode == "" {
		if mappedCode, ok := mapHotelSession[hotelCode]; ok && mappedCode != "" {
			giHotelCode = mappedCode
		}
	}
	if giHotelCode == "" {
		giHotelCode = hotelCode
	}
	return giHotelCode
}
