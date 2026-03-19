package giresponsecommon

type TpaExtensions struct {
	InvoiceReference *string        `json:"invoiceReference,omitempty"`
	TpaExtension     []TpaExtension `json:"tpaExtension,omitempty"`
}

type TpaExtension struct {
	Client       *TpaClient       `json:"client,omitempty"`
	File         *TpaFile         `json:"file,omitempty"`
	InternalUser *TpaInternalUser `json:"internalUser,omitempty"`
}

type TpaClient struct {
	ID   *string `json:"iD,omitempty"`
	Name *string `json:"name,omitempty"`
}

type TpaInternalUser struct {
	ID              *string `json:"iD,omitempty"`
	MessagePassword *string `json:"messagePassword,omitempty"`
}

type TpaFile struct {
	AdultsTotal       int             `json:"adultsTotal,omitempty"`
	ChildrenTotal     int             `json:"childrenTotal,omitempty"`
	ClientReference   *string         `json:"clientReference,omitempty"`
	Distribution      string          `json:"distribution,omitempty"`
	Documentation     bool            `json:"documentation,omitempty"`
	OwnReference      string          `json:"ownReference,omitempty"`
	Rebook            bool            `json:"rebook,omitempty"`
	RebookProviderRef *string         `json:"rebookProviderOwnReference,omitempty"`
	RoomsTotal        int             `json:"roomsTotal,omitempty"`
	SaleChannel       *TpaSaleChannel `json:"saleChannel,omitempty"`
	ServiceTotal      int             `json:"serviceTotal,omitempty"`
	Services          *TpaServices    `json:"services,omitempty"`
}

type TpaSaleChannel struct {
	ID   string `json:"iD,omitempty"`
	Name string `json:"name,omitempty"`
}

type TpaServices struct {
	Service []TpaService `json:"service,omitempty"`
}

type TpaService struct {
	AccommodationID         string             `json:"accommodationID,omitempty"`
	AccommodationName       string             `json:"accommodationName,omitempty"`
	AdultAges               string             `json:"adultAges,omitempty"`
	BillingProviderID       *string            `json:"billingProviderID,omitempty"`
	BookingCode             string             `json:"bookingCode,omitempty"`
	BuyAmount               *float64           `json:"buyAmount,omitempty"`
	BuyChannel              *TpaSaleChannel    `json:"buyChannel,omitempty"`
	BuyCurrency             string             `json:"buyCurrency,omitempty"`
	CancelPolicies          *TpaCancelPolicies `json:"cancelPolicies,omitempty"`
	ChildrensAges           string             `json:"childrensAges,omitempty"`
	Comments                *TpaComments       `json:"comments,omitempty"`
	Currency                *string            `json:"currency,omitempty"`
	DecreaseQuota           bool               `json:"decreaseQuota,omitempty"`
	Distribution            string             `json:"distribution,omitempty"`
	ExchangeRate            *float64           `json:"exchangeRate,omitempty"`
	Identifier              int                `json:"identifier,omitempty"`
	Market                  string             `json:"market,omitempty"`
	Markup                  *float64           `json:"markup,omitempty"`
	Nationality             string             `json:"nationality,omitempty"`
	NonRefundableRate       bool               `json:"nonRefundableRate,omitempty"`
	NumberOfAdults          int                `json:"numberOfAdults,omitempty"`
	NumberOfChildren        int                `json:"numberOfChildren,omitempty"`
	NumberOfRooms           int                `json:"numberOfRooms,omitempty"`
	RebookProserID          *string            `json:"rebookProserId,omitempty"`
	Remarks                 *string            `json:"remarks,omitempty"`
	RetailAmount            *float64           `json:"retailAmount,omitempty"`
	Rooms                   *TpaRooms          `json:"rooms,omitempty"`
	SaleAmount              *float64           `json:"saleAmount,omitempty"`
	SaleCurrency            string             `json:"saleCurrency,omitempty"`
	Status                  *string            `json:"status,omitempty"`
	StayDateRange           *TpaStayRange      `json:"stayDateRange,omitempty"`
	SupplierAccommodationID string             `json:"supplierAccommodationID,omitempty"`
	SupplierReference       *string            `json:"supplierReference,omitempty"`
}

type TpaCancelPolicies struct {
	DateRange  []TpaCancelDate `json:"dateRange,omitempty"`
	Identifier int             `json:"identifier,omitempty"`
}

type TpaCancelDate struct {
	Amount   *float64 `json:"amount,omitempty"`
	Currency string   `json:"currency,omitempty"`
	End      string   `json:"end,omitempty"`
	Source   *string  `json:"source,omitempty"`
	Start    string   `json:"start,omitempty"`
	Type     string   `json:"type,omitempty"`
	Value    int      `json:"value,omitempty"`
}

type TpaComments struct {
	Comment []TpaComment `json:"comment,omitempty"`
}

type TpaComment struct {
	Text string `json:"text,omitempty"`
	Type string `json:"type,omitempty"`
}

type TpaRooms struct {
	Rooms []TpaRoom `json:"room,omitempty"`
}

type TpaRoom struct {
	AdultsAges       string             `json:"adultsAges,omitempty"`
	Board            *TpaBoard          `json:"board,omitempty"`
	ChildrenAges     string             `json:"childrenAges,omitempty"`
	Distribution     string             `json:"distribution,omitempty"`
	ID               string             `json:"iD,omitempty"`
	Identifier       int                `json:"identifier,omitempty"`
	Name             string             `json:"name,omitempty"`
	NumberOfAdults   int                `json:"numberOfAdults,omitempty"`
	NumberOfChildren int                `json:"numberOfChildren,omitempty"`
	Status           *string            `json:"status,omitempty"`
	SumUpDetails     *TpaSumUpContainer `json:"sumUpDetails,omitempty"`
	SupplierName     string             `json:"supplierName,omitempty"`
	SupplierRoomID   string             `json:"supplierRoomID,omitempty"`
	Code             string             `json:"code"`
	Language         string             `json:"language"`
}

type TpaBoard struct {
	ID   string `json:"iD,omitempty"`
	Name string `json:"name,omitempty"`
}

type TpaSumUpContainer struct {
	SumUpDetails []TpaSumUp `json:"sumUpDetails,omitempty"`
}

type TpaSumUp struct {
	BuyAmount    *float64 `json:"buyAmount,omitempty"`
	BuyCurrency  string   `json:"buyCurrency,omitempty"`
	Concept      string   `json:"concept,omitempty"`
	Quantity     int      `json:"quantity,omitempty"`
	SaleAmount   *float64 `json:"saleAmount,omitempty"`
	SaleCurrency string   `json:"saleCurrency,omitempty"`
	Type         string   `json:"type,omitempty"`
}

type TpaStayRange struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}
