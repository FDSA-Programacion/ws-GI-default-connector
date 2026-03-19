package domain

import common_domain "ws-int-httr/internal/domain/gi_response_common"

// ------------------------------------------------------------------
// Response: BookResponse (JSON Model)
// ------------------------------------------------------------------

type BookResponse struct {
	HotelReservation []BookHotelReservation `json:"hotelReservation"`
}

// ------------------------------------------------------------------
// Other Auxiliar
// ------------------------------------------------------------------

type ResGuest struct {
	Profiles          Profiles `json:"profiles"`
	AgeQualifyingCode string   `json:"ageQualifyingCode"`
}

type Profiles struct {
	ProfileInfo ProfileInfo `json:"profileInfo"`
}

type ProfileInfo struct {
	Profile []Profile `json:"profile"`
}

type Profile struct {
	Customer    Customer `json:"customer"`
	ProfileType string   `json:"profileType"`
	Rph         string   `json:"rph"`
}

type Customer struct {
	BirthDate  string     `json:"birthDate"`
	PersonName PersonName `json:"personName"`
}

type PersonName struct {
	GivenName string `json:"givenName"`
	Surname   string `json:"surname"`
}

type BookHotelReservation struct {
	CreatorID string `json:"creatorID"`
	RoomStays struct {
		RoomStay []BookRoomStay `json:"roomStay"`
	} `json:"roomStays"`
}

type BookRoomStay struct {
	CancelPenalties *BookCancelPenalties `json:"cancelPenalties"`
	Comments        *BookComments        `json:"comments"`
	GuestCounts     interface{}          `json:"guestCounts,omitempty"`
	MarketCode      *string              `json:"marketCode,omitempty"`
	ResGuestRPHs    interface{}          `json:"resGuestRPHs,omitempty"`
	RoomRates       struct {
		RoomRate struct {
			AvailabilityStatus *string `json:"availabilityStatus,omitempty"`
			BookingCode        *string `json:"bookingCode,omitempty"`
			DirectPayment      *string `json:"directPayment,omitempty"`
			InvBlockCode       *string `json:"invBlockCode,omitempty"`
			OpenBookingCode    string  `json:"openBookingCode,omitempty"`
			RoomTypeCode       *string `json:"roomTypeCode,omitempty"`
			Total              struct {
				NonRefundable *string `json:"nonRefundable,omitempty"`
				Taxes         struct {
					Amount       *float64 `json:"amount,omitempty"`
					RetailAmount *float64 `json:"retailAmount,omitempty"`
					CurrencyCode string   `json:"currencyCode"`
				} `json:"taxes"`
			} `json:"total"`
		} `json:"roomRate"`
	} `json:"roomRates"`
	RoomStayCandidateRPH interface{} `json:"roomStayCandidateRPH,omitempty"`
}

// BookCancelPenalties es específico para Book, permite cancelPenalty como null
type BookCancelPenalties struct {
	CancelPenalty *[]common_domain.CancelPenalty `json:"cancelPenalty"`
}

// BookComments es específico para Book, permite comment como null
type BookComments struct {
	Comment *[]Comment `json:"comment"`
}
