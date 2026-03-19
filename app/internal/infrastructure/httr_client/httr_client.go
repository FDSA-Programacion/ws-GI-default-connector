package httr_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"ws-int-httr/internal/domain"
	"ws-int-httr/internal/domain/log_domain"
	"ws-int-httr/internal/infrastructure"
	"ws-int-httr/internal/infrastructure/config"
	"ws-int-httr/internal/infrastructure/mapping/hoteltrader"

	"ws-int-httr/internal/infrastructure/serializer"
	"ws-int-httr/internal/infrastructure/session"
)

type HttrClientImpl struct {
	httpClient *http.Client
	serializer serializer.Serializer
	config     config.ProviderConfig
}

func NewHttrClientImpl(cfg config.ProviderConfig, ser serializer.Serializer) domain.BookingProvider {
	client := &http.Client{Timeout: time.Duration(cfg.ProviderTimeoutMs()) * time.Millisecond}

	return &HttrClientImpl{
		httpClient: client,
		serializer: ser,
		config:     cfg,
	}
}

var _ domain.BookingProvider = (*HttrClientImpl)(nil)

// SendAvail gestiona el flujo de Disponibilidad (Avail) usando GraphQL.
func (c *HttrClientImpl) SendAvail(req *domain.AvailRequest) (*domain.BaseJsonRS[*domain.AvailResponse], error) {
	// Marcar que se va a enviar al proveedor
	if sessionContext := session.FromContext(); sessionContext != nil {
		if availLogVal, ok := sessionContext.Get("availLog"); ok {
			if availLog, ok := availLogVal.(*log_domain.AvailLog); ok {
				availLog.SentToSupplier = true
			}
		}
	}

	// Mapeo Canónico -> GraphQL Struct
	graphqlReq := hoteltrader.GIAvailRequestToProvider(req, c.config)

	// Serialización, Comunicación HTTP, y Deserialización
	var graphqlResp hoteltrader.ProviderAvailRS

	// Extraer channel code para autenticación
	channelCode := ""
	if len(req.InternalCondition.Channels.Channel) > 0 {
		channelCode = req.InternalCondition.Channels.Channel[0].Code
	}

	// Serializar JSON antes de enviar para el log
	jsonBytes, err := json.Marshal(graphqlReq)
	if err != nil {
		return nil, fmt.Errorf("httr_client: error al serializar JSON para avail: %w", err)
	}

	if err := c.executeProviderCall("avail", jsonBytes, &graphqlResp, channelCode); err != nil {
		return nil, err
	}

	// Traducción GraphQL Struct -> Canónico (pasando el request original)
	domainResp := hoteltrader.ProviderAvailResponseToGI(&graphqlResp, req)
	return domainResp, nil
}

func (c *HttrClientImpl) SendPreBook(req *domain.PreBookRequest) (*domain.BaseJsonRS[*domain.PreBookResponse], error) {
	// Marcar que se va a enviar al proveedor
	if sessionContext := session.FromContext(); sessionContext != nil {
		if preBookLogVal, ok := sessionContext.Get("preBookLog"); ok {
			if preBookLog, ok := preBookLogVal.(*log_domain.HotelResBookLog); ok {
				preBookLog.SentToSupplier = true
			}
		}
	}

	// Mapeo Canónico -> GraphQL Struct
	graphqlReq := hoteltrader.GIPrebookRequestToProvider(req, c.config)
	var graphqlResp hoteltrader.ProviderPrebookRS

	// Extraer channel code para autenticación
	channelCode := ""
	if len(req.InternalCondition.Channels.Channel) > 0 {
		channelCode = req.InternalCondition.Channels.Channel[0].Code
	}

	// Serializar JSON antes de enviar para el log
	jsonBytes, err := json.Marshal(graphqlReq)
	if err != nil {
		return nil, fmt.Errorf("httr_client: error al serializar JSON para prebook: %w", err)
	}

	if err := c.executeProviderCall("prebook", jsonBytes, &graphqlResp, channelCode); err != nil {
		return nil, err
	}

	// Traducción GraphQL Struct -> Canónico (pasando el request original)
	domainResp := hoteltrader.ProviderPrebookResponseToGI(&graphqlResp, req)
	return domainResp, nil
}

func (c *HttrClientImpl) SendBook(req *domain.BookRequest) (*domain.BaseJsonRS[*domain.BookResponse], error) {
	// Marcar que se va a enviar al proveedor
	if sessionContext := session.FromContext(); sessionContext != nil {
		if bookLogVal, ok := sessionContext.Get("bookLog"); ok {
			if bookLog, ok := bookLogVal.(*log_domain.HotelResCommitLog); ok {
				bookLog.SentToSupplier = true
			}
		}
	}

	// Mapeo Canónico -> GraphQL Struct
	graphqlReq := hoteltrader.GIBookRequestToProvider(req, c.config)
	var graphqlResp hoteltrader.ProviderBookRS

	// Extraer channel code para autenticación
	channelCode := ""
	if len(req.InternalCondition.Channels.Channel) > 0 {
		channelCode = req.InternalCondition.Channels.Channel[0].Code
	}

	// Serializar JSON antes de enviar para el log
	jsonBytes, err := json.Marshal(graphqlReq)
	if err != nil {
		return nil, fmt.Errorf("httr_client: error al serializar JSON para book: %w", err)
	}

	if err := c.executeProviderCall("book", jsonBytes, &graphqlResp, channelCode); err != nil {
		return nil, err
	}

	// Traducción GraphQL Struct -> Canónico (pasando el request original)
	domainResp := hoteltrader.ProviderBookResponseToGI(&graphqlResp, req)
	return domainResp, nil
}

func (c *HttrClientImpl) SendCancel(req *domain.CancelRequest) (*domain.BaseJsonRS[*domain.CancelResponse], error) {
	// Marcar que se va a enviar al proveedor
	if sessionContext := session.FromContext(); sessionContext != nil {
		if cancelLogVal, ok := sessionContext.Get("cancelLog"); ok {
			if cancelLog, ok := cancelLogVal.(*log_domain.CancelLog); ok {
				cancelLog.SentToSupplier = true
			}
		}
	}

	// Mapeo Canónico -> GraphQL Struct
	graphqlReq := hoteltrader.GICancelRequestToProvider(req, c.config)
	var graphqlResp hoteltrader.ProviderCancelRS

	// Extraer channel code para autenticación
	channelCode := ""
	if len(req.InternalCondition.Channels.Channel) > 0 {
		channelCode = req.InternalCondition.Channels.Channel[0].Code
	}

	// Serializar JSON antes de enviar para el log
	jsonBytes, err := json.Marshal(graphqlReq)
	if err != nil {
		return nil, fmt.Errorf("httr_client: error al serializar JSON para cancel: %w", err)
	}

	if err := c.executeProviderCall("cancel", jsonBytes, &graphqlResp, channelCode); err != nil {
		return nil, err
	}

	// Traducción GraphQL Struct -> Canónico (pasando el request original)
	domainResp := hoteltrader.ProviderCancelResponseToGI(&graphqlResp, req)
	return domainResp, nil
}

// getSessionKeyPrefix normaliza el nombre del endpoint para las claves de sesión
// "prebook" se convierte a "preBook" para mantener compatibilidad con los mappers existentes
func getSessionKeyPrefix(endpoint string) string {
	if endpoint == "prebook" {
		return "preBook"
	}
	return endpoint
}

// setExternalRequestResponseInLog establece el request y response externos en el log desde la sesión
// Para Hotel Trader, guarda JSON (GraphQL), no XML
// Solo establece los campos que no están vacíos, preservando los valores existentes
func setExternalRequestResponseInLog(endpoint string, rqExternal string, rsExternal string) {
	sessionContext := session.FromContext()
	if sessionContext == nil {
		return
	}

	if endpoint == "avail" {
		availLog := sessionContext.Data().AvailLog
		if availLog == nil {
			return
		}

		// En avail: rqProvider solo con debug activo
		if rqExternal != "" && strings.EqualFold(sessionContext.Data().Debug, infrastructure.DEBUG_PASSWORD) {
			availLog.RqProvider = rqExternal
		}
		// Habilitado para pruebas: guardar rsProvider en avail
		if rsExternal != "" {
			availLog.RsProvider = rsExternal
		}
		return
	}

	prefix := getSessionKeyPrefix(endpoint)
	logKey := prefix + "Log"

	// Intentar obtener el log específico del endpoint y actualizar directamente los campos
	if logInterface, ok := sessionContext.Get(logKey); ok && logInterface != nil {
		switch log := logInterface.(type) {
		case *log_domain.AvailLog:
			if rqExternal != "" {
				log.RqProvider = rqExternal
			}
			if rsExternal != "" {
				log.RsProvider = rsExternal
			}
		case *log_domain.HotelResBookLog:
			if rqExternal != "" {
				log.RqProvider = rqExternal
			}
			if rsExternal != "" {
				log.RsProvider = rsExternal
			}
		case *log_domain.HotelResCommitLog:
			if rqExternal != "" {
				log.RqProvider = rqExternal
			}
			if rsExternal != "" {
				log.RsProvider = rsExternal
			}
		case *log_domain.CancelLog:
			if rqExternal != "" {
				log.RqProvider = rqExternal
			}
			if rsExternal != "" {
				log.RsProvider = rsExternal
			}
		}
	}
}

// getProviderURL obtiene la URL correcta según el endpoint
func (c *HttrClientImpl) getProviderURL(endpoint string) string {
	switch endpoint {
	case "avail":
		return c.config.ProviderSearchURL()
	case "prebook":
		return c.config.ProviderQuoteURL()
	case "book":
		return c.config.ProviderBookURL()
	case "cancel":
		return c.config.ProviderCancelURL()
	default:
		// Si el endpoint no es reconocido, usar SearchURL como fallback
		return c.config.ProviderSearchURL()
	}
}

// saveSessionMetrics guarda las métricas genéricas en la sesión (sin prefijo de endpoint)
func saveSessionMetrics(supplierRsTime int64, supplierRsHttpStatusCode int, supplierRsLength int, supplierErrorMessage string) {
	sessionContext := session.FromContext()
	if sessionContext == nil {
		return
	}

	// Guardar un único objeto de métricas genéricas en sesión
	sessionContext.Set("supplierMetrics", &session.EndpointMetrics{
		RsTime:         supplierRsTime,
		HttpStatusCode: supplierRsHttpStatusCode,
		RsLength:       supplierRsLength,
		ErrorMessage:   supplierErrorMessage,
	})
}

// GraphQLResponse es una interfaz para las respuestas GraphQL que tienen errores
type GraphQLResponse interface {
	GetErrors() []hoteltrader.GraphQLError
}

// executeProviderCall realiza una llamada GraphQL genérica al proveedor
func (c *HttrClientImpl) executeProviderCall(endpoint string, jsonBytes []byte, respData interface{}, channelCode string) error {
	url := c.getProviderURL(endpoint)

	// Establecer JSON request en el log antes de hacer la petición
	setExternalRequestResponseInLog(endpoint, string(jsonBytes), "")

	// Crear la petición HTTP con headers para GraphQL
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return fmt.Errorf("httr_client: error al crear petición HTTP para %s: %w", endpoint, err)
	}

	// Configurar headers GraphQL
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Autenticación con token para PreBook, Book, Cancel (Avail no lo necesita)
	authToken := c.config.ProviderAuthForChannel(channelCode)
	if authToken != "" {
		req.Header.Set("Authorization", "Basic "+authToken)
	}

	// Realizar la petición y medir tiempo
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	supplierRsTime := time.Since(startTime).Milliseconds()

	if err != nil {
		return fmt.Errorf("httr_client: error de red al llamar a %s en %s: %w", endpoint, url, err)
	}

	defer resp.Body.Close()

	// Leer el cuerpo de la respuesta como []byte
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("httr_client: error al leer respuesta de %s: %w", endpoint, err)
	}

	// Establecer JSON response en el log
	setExternalRequestResponseInLog(endpoint, "", string(bodyBytes))

	supplierRsHttpStatusCode := resp.StatusCode
	supplierRsLength := len(bodyBytes)
	supplierErrorMessage := ""

	if resp.StatusCode != http.StatusOK {
		supplierErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))

		// Guardar error en sesión antes de retornar
		saveSessionMetrics(supplierRsTime, supplierRsHttpStatusCode, supplierRsLength, supplierErrorMessage)

		return fmt.Errorf("httr_client: proveedor %s devolvió estado HTTP %d. Body: %s", endpoint, resp.StatusCode, string(bodyBytes))
	}

	// Guardar datos del HTTP response en sesión para logging
	saveSessionMetrics(supplierRsTime, supplierRsHttpStatusCode, supplierRsLength, supplierErrorMessage)

	// Deserializar JSON a la estructura de respuesta GraphQL
	if err := json.Unmarshal(bodyBytes, respData); err != nil {
		return fmt.Errorf("httr_client: error al deserializar JSON de %s: %w", endpoint, err)
	}

	// Verificar si hay errores en la respuesta GraphQL usando type assertion
	if graphqlResp, ok := respData.(GraphQLResponse); ok {
		if errors := graphqlResp.GetErrors(); len(errors) > 0 {
			errorMessages := ""
			for _, gqlErr := range errors {
				errorMessages += gqlErr.Message + "; "
			}
			return fmt.Errorf("httr_client: errores GraphQL en %s: %s", endpoint, errorMessages)
		}
	}

	return nil
}
