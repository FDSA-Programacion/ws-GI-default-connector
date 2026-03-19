package domain

import (
	"errors"
)

// Errores de validación
var (
	ErrMissingDates = errors.New("booking: fechas de check-in o check-out no especificadas")
)

// Errores de reserva/cancelación
var (
	ErrUniqueIDFormat   = errors.New("cancel: uniqueID debe contener ourId y externalId separados por |")
	ErrUniqueIDNotFound = errors.New("cancel: uniqueID con type 14 no encontrado")
)

// Errores de infraestructura
var (
	ErrCacheNotFound = errors.New("cache: no encontrada")
	ErrXMLSerialize  = errors.New("xml: error de serializacion")
	ErrXMLParse      = errors.New("xml: error de deserializacion")
	ErrHTTPRequest   = errors.New("http: error en peticion")
	ErrHTTPResponse  = errors.New("http: respuesta invalida del proveedor")
)

// Errores de autenticación para admin endpoints
var (
	ErrUnauthorized  = errors.New("invalid credentials")
	ErrNoCredentials = errors.New("missing credentials")
)

// Errores de código personalizados
var (
	ErrorConnectionSupplier        = CustomError{Message: "Connection error with supplier", ErrorCode: "CONNERR", Err: nil}
	ErrorConnectionSupplierTimeout = CustomError{Message: "Timeout with supplier", ErrorCode: "TIMEOUT", Err: nil}
	ErrorUnmarshalToStruct         = CustomError{Message: "Unmarshal error while cast to struct", ErrorCode: "0000001", Err: nil}
	ErrorMarshalToStruct           = CustomError{Message: "Marshal error while cast to struct", ErrorCode: "0000020", Err: nil}
	ErrorExternalUnmarshalToStruct = CustomError{Message: "Unmarshal external error while cast to struct", ErrorCode: "0000002", Err: nil}
	ErrorInvalidRequestType        = CustomError{Message: "Invalid request type", ErrorCode: "E402", Err: nil}
	ErrorInvalidJSON               = CustomError{Message: "Invalid JSON format", ErrorCode: "400", Err: nil}
	ErrorAvailCode                 = CustomError{Message: "Error on avail", ErrorCode: "0000003", Err: nil}
	ErrorPrebookCode               = CustomError{Message: "Error on prebook", ErrorCode: "23232", Err: nil}
	ErrorPrebookPriceDiff          = CustomError{Message: "Error on prebook. Price difference", ErrorCode: "19203", Err: nil}
	ErrorBookCode                  = CustomError{Message: "Error on book", ErrorCode: "9948032", Err: nil}
	ErrorBookingStatus             = CustomError{Message: "Error on book status", ErrorCode: "9948033", Err: nil}
	ErrorCancelCode                = CustomError{Message: "Error on cancel", ErrorCode: "389202", Err: nil}
	ErrorCancelBadRequest          = CustomError{Message: "Bad request on cancel", ErrorCode: "789327389", Err: nil}
	ErrorSoapError                 = CustomError{Message: "SOAP error", ErrorCode: "0000008", Err: nil}
	ErrorPanicInternal             = CustomError{Message: "Internal panic error", ErrorCode: "500", Err: nil}

	// Errores específicos de mappers
	ErrorRepositoryNotFound     = CustomError{Message: "Repository not found", ErrorCode: "0000010", Err: nil}
	ErrorConfigNotFound         = CustomError{Message: "Configuration not found", ErrorCode: "0000011", Err: nil}
	ErrorAvailLogNotFound       = CustomError{Message: "AvailLog not found in context", ErrorCode: "0000012", Err: nil}
	ErrorContextDataNotFound    = CustomError{Message: "ContextData not found", ErrorCode: "0000013", Err: nil}
	ErrorDateParsing            = CustomError{Message: "Error parsing date", ErrorCode: "0000014", Err: nil}
	ErrorXMLDeserialization     = CustomError{Message: "Error deserializing XML", ErrorCode: "0000015", Err: nil}
	ErrorXMLSerialization       = CustomError{Message: "Error serializing XML", ErrorCode: "0000016", Err: nil}
	ErrorDataConversion         = CustomError{Message: "Error converting data", ErrorCode: "0000017", Err: nil}
	ErrorRoomCandidatesNotFound = CustomError{Message: "RoomCandidates not found", ErrorCode: "0000018", Err: nil}
	ErrorPreBookLogNotFound     = CustomError{Message: "PreBookLog not found in context", ErrorCode: "0000019", Err: nil}
	ErrorRoomStaysNotFound      = CustomError{Message: "RoomStays not found in request context", ErrorCode: "0000021", Err: nil}
)

// CustomError representa un error personalizado con código
type CustomError struct {
	ErrorCode string
	Message   string
	Err       error
}

func (e CustomError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// ErrorType representa un error individual en la respuesta
type ErrorType struct {
	Error string `json:"error"`
	Type  int    `json:"type"`
}

// ErrorsContainer contiene la lista de errores
type ErrorsContainer struct {
	ErrorsType []ErrorType `json:"errorsType"`
}
