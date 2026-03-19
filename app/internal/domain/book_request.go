package domain

import giresponsecommon "ws-int-httr/internal/domain/gi_response_common"

// ------------------------------------------------------------------
// Request: BookRequest (JSON Model)
// ------------------------------------------------------------------

type BookRequest struct {
	Debug             string `json:"debug"`
	EchoToken         string `json:"echoToken"`
	HotelReservations struct {
		HotelReservation []struct {
			CreatorID string `json:"creatorID"`
			ResGuests struct {
				ResGuest []ResGuest `json:"resGuest"`
			} `json:"resGuests"`
			RoomStays struct {
				RoomStay []struct {
					BasicPropertyInfo interface{} `json:"basicPropertyInfo"`
					CancelPenalties   interface{} `json:"cancelPenalties"`
					Comments          struct {
						Comment []struct {
							Text string `json:"text"`
						} `json:"comment"`
					} `json:"comments"`
					GuestCounts  interface{} `json:"guestCounts"`
					MarketCode   interface{} `json:"marketCode"`
					ResGuestRPHs struct {
						ResGuestRPH []struct {
							Rph string `json:"rph"`
						} `json:"resGuestRPH"`
					} `json:"resGuestRPHs"`
					RoomRates struct {
						RoomRate struct {
							AvailabilityStatus  interface{} `json:"availabilityStatus"`
							BookingCode         string      `json:"bookingCode"`
							DirectPayment       interface{} `json:"directPayment"`
							InvBlockCode        interface{} `json:"invBlockCode"`
							OpenBookingCode     interface{} `json:"openBookingCode"`
							RoomRateDescription interface{} `json:"roomRateDescription"`
							RoomTypeCode        interface{} `json:"roomTypeCode"`
							Total               interface{} `json:"total"`
							TpaExtensions       struct {
								TpaExtension []giresponsecommon.TpaExtension `json:"tpaExtension"`
							} `json:"tpaExtensions"`
						} `json:"roomRate"`
					} `json:"roomRates"`
					RoomStayCandidateRPH interface{} `json:"roomStayCandidateRPH"`
					TimeSpan             interface{} `json:"timeSpan"`
				} `json:"roomStay"`
			} `json:"roomStays"`
		} `json:"hotelReservation"`
	} `json:"hotelReservations"`
	InternalCondition struct {
		CallCondition struct {
			EchoToken       string `json:"echoToken"`
			OriginTimeStamp int64  `json:"originTimeStamp"`
		} `json:"callCondition"`
		Channels struct {
			Channel []struct {
				ID   int    `json:"id"`
				Code string `json:"code"`
			} `json:"channel"`
		} `json:"channels"`
		ClientCondition struct {
			DeltaMarkup     float64 `json:"deltaMarkup"`
			Markup          float64 `json:"markup"`
			SupportNoRefund bool    `json:"supportNoRefund"`
			Type            string  `json:"type"`
			Code            string  `json:"code"`
		} `json:"clientCondition"`
		Code              interface{} `json:"code"`
		Forced            interface{} `json:"forced"`
		ID                interface{} `json:"id"`
		Name              string      `json:"name"`
		ProviderCondition struct {
			Code     string `json:"code"`
			Password string `json:"password"`
			TimeOut  int    `json:"timeOut"`
			URL      string `json:"url"`
			User     string `json:"user"`
		} `json:"providerCondition"`
		Rebook bool `json:"rebook"`
	} `json:"internalCondition"`
	Pos struct {
		Source struct {
			RequestorID struct {
				CompanyName struct {
					CompanyShortName string `json:"CompanyShortName"`
					TravelSector     string `json:"TravelSector"`
				} `json:"companyName"`
				ID              string `json:"id"`
				MessagePassword string `json:"messagePassword"`
				Type            string `json:"type"`
			} `json:"requestorID"`
			BookingChannel struct {
				ID   int    `json:"id"`
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"bookingChannel"`
		} `json:"source"`
	} `json:"pos"`
	PrimaryLangID         string `json:"primaryLangID"`
	ResStatus             string `json:"resStatus"`
	RebTransactionIdentifier string `json:"rebTransactionIdentifier"`
	RqType                string `json:"rqType"`
	TransactionIdentifier string `json:"transactionIdentifier"`
	Version               string `json:"version"`
}
