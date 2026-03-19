package domain

// ------------------------------------------------------------------
// Request: CancelRequest (JSON Model)
// ------------------------------------------------------------------

type CancelRequest struct {
	Debug             string  `json:"debug"`
	CancelType        string  `json:"cancelType"`
	EchoToken         string  `json:"echoToken"`
	PrimaryLangID     string  `json:"primaryLangID"`
	RqType            string  `json:"rqType"`
	Version           float64 `json:"version"`
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
			Code            string  `json:"code"`
			DeltaMarkup     float64 `json:"deltaMarkup"`
			Markup          float64 `json:"markup"`
			SupportNoRefund bool    `json:"supportNoRefund"`
			Type            string  `json:"type"`
		} `json:"clientCondition"`
		Code              interface{} `json:"code"`
		Forced            interface{} `json:"forced"`
		ID                interface{} `json:"id"`
		Name              interface{} `json:"name"`
		Rebook            bool        `json:"rebook"`
		ProviderCondition struct {
			Code     string      `json:"code"`
			Password interface{} `json:"password"`
			TimeOut  int         `json:"timeOut"`
			URL      interface{} `json:"url"`
			User     interface{} `json:"user"`
		} `json:"providerCondition"`
	} `json:"internalCondition"`
	Pos struct {
		Source struct {
			RequestorID struct {
				CompanyName struct {
					CompanyShortName interface{} `json:"companyShortName"`
					TravelSector     interface{} `json:"travelSector"`
				} `json:"companyName"`
				ID              string      `json:"id"`
				MessagePassword string      `json:"messagePassword"`
				Type            interface{} `json:"type"`
			} `json:"requestorID"`
			BookingChannel struct {
				ID   int    `json:"id"`
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"bookingChannel"`
		} `json:"source"`
	} `json:"pos"`
	UniqueID []CancelUniqueID `json:"uniqueID"`
}

// UniqueID para Cancel (estructura diferente a otros endpoints)
type CancelUniqueID struct {
	CompanyName struct {
		CompanyShortName interface{} `json:"companyShortName"`
		TravelSector     interface{} `json:"travelSector"`
	} `json:"companyName"`
	ID   string `json:"id"`
	Type string `json:"type"`
}

// UniqueID genérico (usado en mappers para contexto)
type UniqueID struct {
	CompanyName string `json:"companyName"`
	ID          string `json:"id"`
	Type        string `json:"type"`
}
