package domain

import common_domain "ws-int-httr/internal/domain/gi_response_common"

// Constantes para tipos de comentarios
const (
	CommentTypeProvider = "PROVIDER"
)

// ------------------------------------------------------------------
// GuestCounts
// ------------------------------------------------------------------

type HotelReservation struct {
	RoomStays struct {
		RoomStay []RoomStay `json:"roomStay"`
	} `json:"roomStays"`
	CreatorID string `json:"creatorID"`
}

type RoomStay struct {
	CancelPenalties *common_domain.CancelPenalties `json:"cancelPenalties"`
	Comments        struct {
		Comment []Comment `json:"comment"`
	} `json:"comments"`
	RoomRates struct {
		RoomRate struct {
			BookingCode     string `json:"bookingCode"`
			OpenBookingCode string `json:"openBookingCode,omitempty"`
			Total           struct {
				NonRefundable string `json:"nonRefundable,omitempty"`
				Taxes         struct {
					Amount       float64  `json:"amount"`
					RetailAmount *float64 `json:"retailAmount,omitempty"`
					CurrencyCode string   `json:"currencyCode"`
				} `json:"taxes"`
			} `json:"total"`
		} `json:"roomRate"`
	} `json:"roomRates"`
}

type Comment struct {
	Text string `json:"text"`
	Type string `json:"type,omitempty"`
}
