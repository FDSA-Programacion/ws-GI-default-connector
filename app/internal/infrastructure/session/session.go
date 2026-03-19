package session

import (
	"context"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"ws-int-httr/internal/domain/log_domain"
)

// SessionData contiene toda la información de la sesión de forma estructurada
type SessionData struct {
	// Request info
	RequestType    string
	RequestRawData []byte
	StartTime      time.Time
	Debug          string
	ProviderCode   int
	EchoToken      string

	// Context data
	ContextData     map[string]interface{}
	MapHotelSession map[string]interface{}

	// Logs
	AvailLog   *log_domain.AvailLog
	PreBookLog *log_domain.HotelResBookLog
	BookLog    *log_domain.HotelResCommitLog
	CancelLog  *log_domain.CancelLog

	// Log timing
	AvailLogStartTime   time.Time
	PreBookLogStartTime time.Time
	BookLogStartTime    time.Time
	CancelLogStartTime  time.Time

	// Supplier response metrics
	SupplierMetrics *EndpointMetrics
	AvailMetrics    *EndpointMetrics
	PreBookMetrics  *EndpointMetrics
	BookMetrics     *EndpointMetrics
	CancelMetrics   *EndpointMetrics

	// Special responses
	BookResponse interface{}

	// Panic recovery
	PanicValue interface{}
}

// EndpointMetrics contiene las métricas de respuesta del proveedor
type EndpointMetrics struct {
	RsTime         int64
	HttpStatusCode int
	RsLength       int
	ErrorMessage   string
}

type Session struct {
	data *SessionData
}

var sessions sync.Map

func New(ctx context.Context) *Session {
	s := &Session{
		data: &SessionData{
			ContextData:     make(map[string]interface{}),
			MapHotelSession: make(map[string]interface{}),
		},
	}
	sessions.Store(gid(), s)
	return s
}

func FromContext() *Session {
	if s, ok := sessions.Load(gid()); ok {
		return s.(*Session)
	}
	return &Session{
		data: &SessionData{
			ContextData:     make(map[string]interface{}),
			MapHotelSession: make(map[string]interface{}),
		},
	}
}

func Clear() {
	sessions.Delete(gid())
}

// Métodos de acceso directo al struct
func (s *Session) Data() *SessionData {
	return s.data
}

// Métodos de compatibilidad temporal con la API antigua
func (s *Session) Set(key string, value any) {
	switch key {
	case "requestType":
		s.data.RequestType = value.(string)
	case "requestRawData":
		s.data.RequestRawData = value.([]byte)
	case "startTime":
		s.data.StartTime = value.(time.Time)
	case "debug":
		s.data.Debug = value.(string)
	case "providerCode":
		s.data.ProviderCode = value.(int)
	case "echoToken":
		s.data.EchoToken = value.(string)
	case "contextData":
		s.data.ContextData = value.(map[string]interface{})
	case "mapHotelSession":
		s.data.MapHotelSession = value.(map[string]interface{})

	// Logs
	case "availLog":
		s.data.AvailLog = value.(*log_domain.AvailLog)
	case "preBookLog":
		s.data.PreBookLog = value.(*log_domain.HotelResBookLog)
	case "bookLog":
		s.data.BookLog = value.(*log_domain.HotelResCommitLog)
	case "cancelLog":
		s.data.CancelLog = value.(*log_domain.CancelLog)

	// Log timing
	case "availLogStartTime":
		s.data.AvailLogStartTime = value.(time.Time)
	case "preBookLogStartTime":
		s.data.PreBookLogStartTime = value.(time.Time)
	case "bookLogStartTime":
		s.data.BookLogStartTime = value.(time.Time)
	case "cancelLogStartTime":
		s.data.CancelLogStartTime = value.(time.Time)

	// Supplier metrics
	case "supplierMetrics":
		s.data.SupplierMetrics = value.(*EndpointMetrics)

	// Special responses
	case "bookResponse":
		s.data.BookResponse = value
	}
}

func (s *Session) Get(key string) (any, bool) {
	switch key {
	case "requestType":
		return s.data.RequestType, s.data.RequestType != ""
	case "requestRawData":
		return s.data.RequestRawData, len(s.data.RequestRawData) > 0
	case "startTime":
		return s.data.StartTime, !s.data.StartTime.IsZero()
	case "debug":
		return s.data.Debug, s.data.Debug != ""
	case "providerCode":
		return s.data.ProviderCode, s.data.ProviderCode != 0
	case "echoToken":
		return s.data.EchoToken, s.data.EchoToken != ""
	case "contextData":
		return s.data.ContextData, s.data.ContextData != nil
	case "mapHotelSession":
		return s.data.MapHotelSession, s.data.MapHotelSession != nil

	// Logs
	case "availLog":
		return s.data.AvailLog, s.data.AvailLog != nil
	case "preBookLog":
		return s.data.PreBookLog, s.data.PreBookLog != nil
	case "bookLog":
		return s.data.BookLog, s.data.BookLog != nil
	case "cancelLog":
		return s.data.CancelLog, s.data.CancelLog != nil

	// Log timing
	case "availLogStartTime":
		return s.data.AvailLogStartTime, !s.data.AvailLogStartTime.IsZero()
	case "preBookLogStartTime":
		return s.data.PreBookLogStartTime, !s.data.PreBookLogStartTime.IsZero()
	case "bookLogStartTime":
		return s.data.BookLogStartTime, !s.data.BookLogStartTime.IsZero()
	case "cancelLogStartTime":
		return s.data.CancelLogStartTime, !s.data.CancelLogStartTime.IsZero()

	// Supplier metrics
	case "supplierMetrics":
		return s.data.SupplierMetrics, s.data.SupplierMetrics != nil

	// Special responses
	case "bookResponse":
		return s.data.BookResponse, s.data.BookResponse != nil

	default:
		return nil, false
	}
}

func (s *Session) GetString(key string) string {
	if v, ok := s.Get(key); ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return ""
}

func gid() uint64 {
	b := make([]byte, 64)
	n := runtime.Stack(b, false)
	idField := strings.Fields(strings.TrimPrefix(string(b[:n]), "goroutine "))[0]
	id, _ := strconv.ParseUint(idField, 10, 64)
	return id
}
