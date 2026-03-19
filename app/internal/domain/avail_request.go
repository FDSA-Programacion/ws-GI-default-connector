package domain

// ------------------------------------------------------------------
// Request: AvailRequest (JSON Model)
// ------------------------------------------------------------------

type AvailRequest struct {
	AvailRequestSegments struct {
		AvailRequestSegment AvailRequestSegment `json:"availRequestSegment"`
	} `json:"availRequestSegments"`
	Debug             string `json:"debug"`
	InternalCondition struct {
		CallCondition struct {
			EchoToken       string `json:"echoToken"`
			OriginTimeStamp int64  `json:"originTimeStamp"`
		} `json:"callCondition"`
		Channels struct {
			Channel []struct {
				Code string `json:"code"`
				ID   int    `json:"id"`
			} `json:"channel"`
		} `json:"channels"`
		ClientCondition struct {
			DeltaMarkup     float64 `json:"deltaMarkup"`
			Markup          float64 `json:"markup"`
			SupportNoRefund bool    `json:"supportNoRefund"`
			Type            string  `json:"type"`
			Code            string  `json:"code"`
		} `json:"clientCondition"`
		Code              string `json:"code"`
		Forced            string `json:"forced"`
		ID                string `json:"id"`
		Name              string `json:"name"`
		ProviderCondition struct {
			Code     string `json:"code"`
			Password string `json:"password"`
			TimeOut  int    `json:"timeOut"`
			URL      string `json:"url"`
			User     string `json:"user"`
		} `json:"providerCondition"`
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
	PrimaryLangID string `json:"primaryLangID"`
	RqType        string `json:"rqType"`
	Timeout       int    `json:"timeout"`
	Valuation     bool   `json:"valuation"`
	Version       string `json:"version"`
}

type AvailRequestSegment struct {
	HotelSearchCriteria struct {
		Criterion struct {
			HotelRef []struct {
				AreaID        string `json:"areaID"`
				HotelCityCode string `json:"hotelCityCode"`
				HotelCode     string `json:"hotelCode"`
			} `json:"hotelRef"`
		} `json:"criterion"`
	} `json:"hotelSearchCriteria"`
	RoomStayCandidates struct {
		RoomStayCandidateList []RoomStayCandidate `json:"roomStayCandidate"`
	} `json:"roomStayCandidates"`
	StayDateRange struct {
		End   string `json:"end"`
		Start string `json:"start"`
	} `json:"stayDateRange"`
	Tpa struct {
		CancelPenalties bool   `json:"cancelPenalties"`
		Market          string `json:"market"`
		Nationality     string `json:"nationality"`
		Currency        string `json:"currency"`
	} `json:"tpa"`
	TpaExtension struct {
		CancelPenalties bool   `json:"cancelPenalties"`
		Market          string `json:"market"`
		Nationality     string `json:"nationality"`
		Currency        string `json:"currency"`
	} `json:"tpaExtension"`
}

type RoomStayCandidate struct {
	GuestCounts struct {
		GuestCountList []RQGuestCount `json:"guestCount"`
	} `json:"guestCounts"`
	Rph string `json:"rph"`
}

type RQGuestCount struct {
	Age               int    `json:"age"`
	AgeQualifyingCode string `json:"ageQualifyingCode"`
}
