package domain

import common_domain "ws-int-httr/internal/domain/gi_response_common"

// ------------------------------------------------------------------
// Response: AvailResponse (JSON Model)
// ------------------------------------------------------------------

type AvailResponse struct {
	GiRoomStayGroup []GiRoomStayGroup `json:"giRoomStayGroup"`
}

type GiRoomStayGroup struct {
	BoardCode  string        `json:"boardCode,omitempty"`
	GiRoomStay []*GiRoomStay `json:"giRoomStay"`
	HotelCode  string        `json:"hotelCode"`
	Key        string        `json:"key"`
	NettPrice  float64       `json:"nettPrice"`
	NonRefund  string        `json:"nonRefund"`
	Price      float64       `json:"price"`
	RoomCode   string        `json:"roomCode"`
	Supplier   string        `json:"supplier"`
}

type GiRoomStay struct {
	BasicPropertyInfo    BasicPropertyInfo             `json:"basicPropertyInfo"`
	CancelPenalties      common_domain.CancelPenalties `json:"cancelPenalties"`
	GuestCounts          GuestCounts                   `json:"guestCounts"`
	MarketCode           string                        `json:"marketCode"`
	RoomRates            RoomRates                     `json:"roomRates"`
	RoomStayCandidateRPH int                           `json:"roomStayCandidateRPH"`
}

type BasicPropertyInfo struct {
	Address       Address `json:"address"`
	Award         Award   `json:"award"`
	HotelCityCode string  `json:"hotelCityCode"`
	HotelCode     string  `json:"hotelCode"`
	HotelName     string  `json:"hotelName"`
}

type Address struct {
	CityName string `json:"cityName"`
}

type Award struct {
	HotelCategory     string `json:"hotelCategory"`
	PropertyClassCode string `json:"propertyClassCode"`
	Rating            string `json:"rating"`
}

type GuestCounts struct {
	GuestCount []RSGuestCount `json:"guestCount"`
}

type RSGuestCount struct {
	AgeQualifyingCode string `json:"ageQualifyingCode"`
	Count             int    `json:"count"`
}

type RoomRates struct {
	RoomRate RoomRate `json:"roomRate"`
}

type RoomRate struct {
	AvailabilityStatus  string              `json:"availabilityStatus"`
	BookingCode         string              `json:"bookingCode"`
	DirectPayment       string              `json:"directPayment"`
	InvBlockCode        string              `json:"invBlockCode"`
	OpenBookingCode     string              `json:"openBookingCode"`
	RoomRateDescription RoomRateDescription `json:"roomRateDescription"`
	RoomTypeCode        string              `json:"roomTypeCode"`
	Total               Total               `json:"total"`
}

type RoomRateDescription struct {
	Text []string `json:"text"`
}

type Total struct {
	NonRefundable string `json:"nonRefundable"`
	Taxes         Taxes  `json:"taxes"`
}

type Taxes struct {
	Amount       float64 `json:"amount"`
	CurrencyCode string  `json:"currencyCode"`
}
