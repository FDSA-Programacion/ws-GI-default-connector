package provider_client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"ws-int-httr/internal/domain"
	"ws-int-httr/internal/domain/log_domain"
	"ws-int-httr/internal/infrastructure"
	"ws-int-httr/internal/infrastructure/config"
	"ws-int-httr/internal/infrastructure/mapping/provider"

	"ws-int-httr/internal/infrastructure/serializer"
	"ws-int-httr/internal/infrastructure/session"
)

type ProviderClientImpl struct {
	httpClient *http.Client
	serializer serializer.Serializer
	config     config.ProviderConfig
}

func NewProviderClientImpl(cfg config.ProviderConfig, ser serializer.Serializer) domain.BookingProvider {
	client := &http.Client{Timeout: time.Duration(cfg.ProviderTimeoutMs()) * time.Millisecond}

	return &ProviderClientImpl{
		httpClient: client,
		serializer: ser,
		config:     cfg,
	}
}

var _ domain.BookingProvider = (*ProviderClientImpl)(nil)

// SendAvail gestiona el flujo de Disponibilidad (Avail) usando XML.
func (c *ProviderClientImpl) SendAvail(req *domain.AvailRequest) (*domain.BaseJsonRS[*domain.AvailResponse], error) {
	// Marcar que se va a enviar al proveedor
	if sessionContext := session.FromContext(); sessionContext != nil {
		if availLogVal, ok := sessionContext.Get("availLog"); ok {
			if availLog, ok := availLogVal.(*log_domain.AvailLog); ok {
				availLog.SentToSupplier = true
			}
		}
	}

	// Mapeo Canónico -> XML Struct
	providerReq := provider.GIAvailRequestToProvider(req, c.config)
	var providerResp provider.ProviderAvailResponse

	// Extraer channel code para autenticación
	channelCode := ""
	if len(req.InternalCondition.Channels.Channel) > 0 {
		channelCode = req.InternalCondition.Channels.Channel[0].Code
	}

	// Serializar XML antes de enviar para el log
	xmlBytes, err := c.serializer.ToXML(providerReq)
	if err != nil {
		return nil, fmt.Errorf("provider_client: error al serializar XML para avail: %w", err)
	}

	if err := c.executeProviderCall("avail", xmlBytes, &providerResp, channelCode); err != nil {
		return nil, err
	}

	// Traducción XML Struct -> Canónico (pasando el request original)
	domainResp := provider.ProviderAvailResponseToGI(&providerResp, req)
	return &domainResp, nil
}

func (c *ProviderClientImpl) SendPreBook(req *domain.PreBookRequest) (*domain.BaseJsonRS[*domain.PreBookResponse], error) {
	// Marcar que se va a enviar al proveedor
	if sessionContext := session.FromContext(); sessionContext != nil {
		if preBookLogVal, ok := sessionContext.Get("preBookLog"); ok {
			if preBookLog, ok := preBookLogVal.(*log_domain.HotelResBookLog); ok {
				preBookLog.SentToSupplier = true
			}
		}
	}

	// Mapeo Canónico -> XML Struct
	providerReq := provider.GIPrebookRequestToProvider(req, c.config)
	var providerResp provider.ProviderPrebookResponse

	// Extraer channel code para autenticación
	channelCode := ""
	if len(req.InternalCondition.Channels.Channel) > 0 {
		channelCode = req.InternalCondition.Channels.Channel[0].Code
	}

	// Serializar XML antes de enviar para el log
	xmlBytes, err := c.serializer.ToXML(providerReq)
	if err != nil {
		return nil, fmt.Errorf("provider_client: error al serializar XML para prebook: %w", err)
	}

	if err := c.executeProviderCall("prebook", xmlBytes, &providerResp, channelCode); err != nil {
		return nil, err
	}

	// Traducción XML Struct -> Canónico (pasando el request original)
	domainResp := provider.ProviderPrebookResponseToGI(&providerResp, req)
	return &domainResp, nil
}

func (c *ProviderClientImpl) SendBook(req *domain.BookRequest) (*domain.BaseJsonRS[*domain.BookResponse], error) {
	// Marcar que se va a enviar al proveedor
	if sessionContext := session.FromContext(); sessionContext != nil {
		if bookLogVal, ok := sessionContext.Get("bookLog"); ok {
			if bookLog, ok := bookLogVal.(*log_domain.HotelResCommitLog); ok {
				bookLog.SentToSupplier = true
			}
		}
	}

	// Mapeo Canónico -> XML Struct
	providerReq := provider.GIBookRequestToProvider(req, c.config)
	var providerResp provider.ProviderBookResponse

	// Extraer channel code para autenticación
	channelCode := ""
	if len(req.InternalCondition.Channels.Channel) > 0 {
		channelCode = req.InternalCondition.Channels.Channel[0].Code
	}

	// Serializar XML antes de enviar para el log
	xmlBytes, err := c.serializer.ToXML(providerReq)
	if err != nil {
		return nil, fmt.Errorf("provider_client: error al serializar XML para book: %w", err)
	}

	if err := c.executeProviderCall("book", xmlBytes, &providerResp, channelCode); err != nil {
		return nil, err
	}

	// Traducción XML Struct -> Canónico (pasando el request original)
	domainResp := provider.ProviderBookResponseToGI(&providerResp, req)
	return &domainResp, nil
}

func (c *ProviderClientImpl) SendCancel(req *domain.CancelRequest) (*domain.BaseJsonRS[*domain.CancelResponse], error) {
	// Marcar que se va a enviar al proveedor
	if sessionContext := session.FromContext(); sessionContext != nil {
		if cancelLogVal, ok := sessionContext.Get("cancelLog"); ok {
			if cancelLog, ok := cancelLogVal.(*log_domain.CancelLog); ok {
				cancelLog.SentToSupplier = true
			}
		}
	}

	// Mapeo Canónico -> XML Struct
	providerReq := provider.GICancelRequestToProvider(req, c.config)
	var providerResp provider.ProviderCancelResponse

	// Extraer channel code para autenticación
	channelCode := ""
	if len(req.InternalCondition.Channels.Channel) > 0 {
		channelCode = req.InternalCondition.Channels.Channel[0].Code
	}

	// Serializar XML antes de enviar para el log
	xmlBytes, err := c.serializer.ToXML(providerReq)
	if err != nil {
		return nil, fmt.Errorf("provider_client: error al serializar XML para cancel: %w", err)
	}

	if err := c.executeProviderCall("cancel", xmlBytes, &providerResp, channelCode); err != nil {
		return nil, err
	}

	// Traducción XML Struct -> Canónico (pasando el request original)
	domainResp := provider.ProviderCancelResponseToGI(&providerResp, req)
	return &domainResp, nil
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
// Para el proveedor, guarda XML externo en los logs
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
func (c *ProviderClientImpl) getProviderURL(endpoint string) string {
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

// executeProviderCall realiza una llamada XML genérica al proveedor
func (c *ProviderClientImpl) executeProviderCall(endpoint string, xmlBytes []byte, respData interface{}, channelCode string) error {
	url := c.getProviderURL(endpoint)

	// Establecer XML request en el log antes de hacer la petición
	setExternalRequestResponseInLog(endpoint, string(xmlBytes), "")

	req, err := http.NewRequest("POST", url, bytes.NewReader(xmlBytes))
	if err != nil {
		return fmt.Errorf("provider_client: error al crear petición HTTP para %s: %w", endpoint, err)
	}

	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Accept", "application/xml")

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
		return fmt.Errorf("provider_client: error de red al llamar a %s en %s: %w", endpoint, url, err)
	}

	defer resp.Body.Close()

	// Leer el cuerpo de la respuesta como []byte
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("provider_client: error al leer respuesta de %s: %w", endpoint, err)
	}

	// Establecer XML response en el log
	setExternalRequestResponseInLog(endpoint, "", string(bodyBytes))

	supplierRsHttpStatusCode := resp.StatusCode
	supplierRsLength := len(bodyBytes)
	supplierErrorMessage := ""

	if resp.StatusCode != http.StatusOK {
		supplierErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))

		// Guardar error en sesión antes de retornar
		saveSessionMetrics(supplierRsTime, supplierRsHttpStatusCode, supplierRsLength, supplierErrorMessage)

		return fmt.Errorf("provider_client: proveedor %s devolvió estado HTTP %d. Body: %s", endpoint, resp.StatusCode, string(bodyBytes))
	}

	// Guardar datos del HTTP response en sesión para logging
	saveSessionMetrics(supplierRsTime, supplierRsHttpStatusCode, supplierRsLength, supplierErrorMessage)

	if respData != nil {
		if err := c.serializer.FromXML(bodyBytes, respData); err != nil {
			return fmt.Errorf("provider_client: error al deserializar XML de %s: %w", endpoint, err)
		}
	}

	return nil
}
