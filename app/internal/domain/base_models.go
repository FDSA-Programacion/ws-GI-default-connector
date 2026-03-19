package domain

import common_domain "ws-int-httr/internal/domain/gi_response_common"

type BaseJsonRS[T any] struct {
	EchoToken             string `json:"echoToken,omitempty"`
	PrimaryLangID         string `json:"primaryLangID,omitempty"`
	SchemaLocation        string `json:"schemaLocation,omitempty"`
	Success               string `json:"success"`
	Version               string `json:"version,omitempty"`
	Xsi                   string `json:"xsi,omitempty"`
	ResResponseType       string `json:"resResponseType,omitempty"`
	TransactionIdentifier string `json:"transactionIdentifier,omitempty"`

	Errors            *ErrorsContainer                 `json:"errors,omitempty"`
	TpaExtensions     *common_domain.TpaExtensions     `json:"tpaExtensions,omitempty"`
	InternalCondition *common_domain.InternalCondition `json:"internalConditionRS,omitempty"`

	// Cancel
	UniqueID []UniqueID    `json:"uniqueID,omitempty"`
	Segment  []interface{} `json:"segment,omitempty"`
	Status   string        `json:"status,omitempty"`

	GiRoomStays       T `json:"giRoomStays,omitempty"`
	HotelReservations T `json:"hotelReservations,omitempty"`
	CancelInfoRS      T `json:"cancelInfoRS,omitempty"`
}

// BaseJsonRSBook especifica para Book
type BaseJsonRSBook struct {
	EchoToken             string `json:"echoToken,omitempty"`
	PrimaryLangID         string `json:"primaryLangID,omitempty"`
	SchemaLocation        string `json:"schemaLocation,omitempty"`
	Success               string `json:"success"`
	Version               string `json:"version,omitempty"`
	Xsi                   string `json:"xsi,omitempty"`
	ResResponseType       string `json:"resResponseType,omitempty"`
	TransactionIdentifier string `json:"transactionIdentifier,omitempty"`

	Errors            *ErrorsContainer                     `json:"errors,omitempty"`
	TpaExtensions     *common_domain.TpaExtensions         `json:"tpaExtensions,omitempty"`
	InternalCondition *common_domain.InternalConditionBook `json:"internalConditionRS,omitempty"`

	HotelReservations BookResponse `json:"hotelReservations,omitempty"`
}
