package log_domain

import (
	"encoding/json"
)

type AvailLog struct {
	ProviderID     string `json:"providerID"`
	ProviderCode   string `json:"providerCode"`
	SentToSupplier bool   `json:"sentToProvider"`
	Node           string `json:"node"`
	Integration    string `json:"integration"`

	//Request params
	EchoToken          string   `json:"echoToken"`
	RqType             string   `json:"rqType"`
	ClientName         string   `json:"clientName"`
	ClientCode         string   `json:"clientCode"`
	RqCity             string   `json:"rqCity"`
	RqHotelCodeList    []string `json:"rqHotelCodeList"`
	RqPrvHotelCodeList []string `json:"rqPrvHotelCodeList"`
	RqZone             string   `json:"rqZone"`
	Market             string   `json:"market"`
	RequestorID        string   `json:"requestorID"`
	BookingChannel     string   `json:"bookingChannel"`
	RqTimestamp        int64    `json:"rqTimestamp"`
	Nationality        string   `json:"nationality"`
	PrimaryLangID      string   `json:"primaryLangID"`
	StayDateRangeEnd   string   `json:"stayDateRangeEnd"`
	StayDateRangeStart string   `json:"stayDateRangeStart"`
	Version            string   `json:"version"`
	RqNumRooms         int      `json:"rqNumRoomStays"`
	RqNumGuests        int      `json:"rqNumGuests"`
	RqDistribution     string   `json:"rqDistribution"`
	RqInternal         string   `json:"rqInternal"`
	RqProvider         string   `json:"rqProvider"`
	IsRebook           bool     `json:"isRebook"`

	//Response params
	Error                    string `json:"error"`
	Success                  string `json:"success"`
	Summary                  string `json:"summary"`
	ErrorCode                int    `json:"errorCode"`
	ErrorMessage             string `json:"errorMessage"`
	InternalMessage          string `json:"internalMessage"`
	RsLength                 int    `json:"rsLength"`
	RsNumHotels              int    `json:"rsNumHotels"`
	RsNumRoomStay            int    `json:"rsNumRoomStays"`
	RsTime                   int    `json:"rsTime"`
	SupplierRsTime           int64  `json:"providerRsTime"`
	SupplierRsHttpStatusCode int    `json:"providerRsHttpStatusCode"`
	SupplierRsLength         int    `json:"providerRsLength"`
	SupplierNumHotels        int    `json:"providerNumHotels"`
	SupplierNumRooms         int    `json:"providerNumRooms"`
	SupplierNumRates         int    `json:"providerNumRates"`
	RsInternal               string `json:"rsInternal"`
	RsProvider               string `json:"rsProvider"`
	SupplierErrorMessage     string `json:"providerErrorMessage"`
	CachedProviderResponse   bool   `json:"cachedProviderResponse"`
}

type HotelResBookLog struct {
	EchoToken                string    `json:"echoToken"`
	Error                    string    `json:"error"`
	PrimaryLangID            string    `json:"primaryLangID"`
	ProviderID               string    `json:"providerID"`
	ProviderCode             string    `json:"providerCode"`
	RequestorID              string    `json:"requestorID"`
	ResResponseType          string    `json:"resResponseType"`
	ResStatus                string    `json:"resStatus"`
	RqInternal               string    `json:"rqInternal"`
	RqProvider               string    `json:"rqProvider"`
	RqTimestamp              int64     `json:"rqTimestamp"`
	RqType                   string    `json:"rqType"`
	RsInternal               string    `json:"rsInternal"`
	RsProvider               string    `json:"rsProvider"`
	Success                  string    `json:"success"`
	TaxesAmountAvail         []float64 `json:"taxesAmountAvail"`
	RetailAmountAvail        []string  `json:"retailAmountAvail"`
	TotalTaxesAmountAvail    float64   `json:"totalTaxesAmountAvail"`
	TaxesAmountBook          []float64 `json:"taxesAmountBook"`
	RetailAmountBook         []string  `json:"retailAmountBook"`
	TotalTaxesAmountBook     float64   `json:"totalTaxesAmountBook"`
	TotalTime                int       `json:"totalTime"`
	Version                  string    `json:"version"`
	SupplierRsTime           int64     `json:"providerRsTime"`
	SupplierRsHttpStatusCode int       `json:"providerRsHttpStatusCode"`
	SupplierRsLength         int       `json:"providerRsLength"`
	ErrorCode                int       `json:"errorCode"`
	ErrorMessage             string    `json:"errorMessage"`
	InternalMessage          string    `json:"internalMessage"`
	SupplierErrorMessage     string    `json:"providerErrorMessage"`
	SentToSupplier           bool      `json:"sentToProvider"`
	Node                     string    `json:"node"`
	Integration              string    `json:"integration"`
	ClientIp                 string    `json:"clientIp"`
	ClientName               string    `json:"clientName"`
	ClientCode               string    `json:"clientCode"`
	BookingChannel           string    `json:"bookingChannel"`
	RqBookingCode            []string  `json:"rqBookingCode"`
	TProviderID              string    `json:"bcProviderID"`
	TCheckInDate             string    `json:"bcCheckInDate"`
	TCheckOutDate            string    `json:"bcCheckOutDate"`
	TNumberOfRooms           int       `json:"bcNumberOfRooms"`
	TAdults                  int       `json:"bcAdults"`
	TChildren                int       `json:"bcChildren"`
	TChildrenAges            string    `json:"bcChildrenAges"`
	TInfant                  int       `json:"bcInfant"`
	TCustomerCountry         string    `json:"bcCustomerCountry"`
	TRoomId                  []string  `json:"bcRoomId"`
	TBoardId                 []string  `json:"bcBoardId"`
	TBuyTotalPriceAvail      float32   `json:"bcBuyTotalPriceAvail"`
	TBuyPriceAvail           []float32 `json:"bcBuyPriceAvail"`
	TBuyTotalPriceBook       float32   `json:"bcBuyTotalPriceBook"`
	TBuyPriceBook            []float32 `json:"bcBuyPriceBook"`
	TDistribution            string    `json:"bcDistribution"`
	TAdultAges               string    `json:"bcAdultAges"`
	TNumberOfService         int       `json:"bcNumberOfService"`

	RsBookingCode []string `json:"rsBookingCode"`
	ErrorCodeSt   string   `json:"errorCodeSt"`

	//New
	TMarket        string   `json:"bcMarket"`
	TGiRoomID      []string `json:"bcGiRoomID"`
	TGiRoomCode    []string `json:"bcGiRoomCode"`
	TGiRoomName    []string `json:"bcGiRoomName"`
	TPrvRoomCode   []string `json:"bcPrvRoomCode"`
	TPrvRoomName   []string `json:"bcPrvRoomName"`
	TIsDynamicRoom []bool   `json:"bcDynamicRoom"`
	TIsMappedRoom  []bool   `json:"bcMappedRoom"`
	TGiHotelCode   string   `json:"bcGiHotelCode"`
	TPrvHotelCode  string   `json:"bcPrvHotelCode"`
	BookedRoomName []string `json:"bookedRoomName"`
	IsRebook       bool     `json:"isRebook"`
}

type HotelResCommitLog struct {
	EchoToken                 string    `json:"echoToken"`
	Error                     string    `json:"error"`
	HotelReservationCreatorID string    `json:"hotelReservationCreatorID"`
	PrimaryLangID             string    `json:"primaryLangID"`
	ProviderID                string    `json:"providerID"`
	ProviderCode              string    `json:"providerCode"`
	ProviderReference         string    `json:"providerReference"`
	RequestorID               string    `json:"requestorID"`
	ResResponseType           string    `json:"resResponseType"`
	ResStatus                 string    `json:"resStatus"`
	RqInternal                string    `json:"rqInternal"`
	RqPrebookSupplier         string    `json:"rqPrebookSupplier"`
	RqProvider                string    `json:"rqProvider"`
	RqTimestamp               int64     `json:"rqTimestamp"`
	RqType                    string    `json:"rqType"`
	RsInternal                string    `json:"rsInternal"`
	RsPrebookSupplier         string    `json:"rsPrebookSupplier"`
	RsProvider                string    `json:"rsProvider"`
	Success                   string    `json:"success"`
	TaxesAmount               []float64 `json:"taxesAmount"`
	RetailAmount              []string  `json:"retailAmount"`
	TotalTaxesAmount          float64   `json:"totalTaxesAmount"`
	TotalTime                 int       `json:"totalTime"`
	TransactionIdentifier     string    `json:"transactionIdentifier"`
	Version                   string    `json:"version"`
	SupplierRsTime            int64     `json:"providerRsTime"`
	SupplierRsHttpStatusCode  int       `json:"providerRsHttpStatusCode"`
	SupplierRsLength          int       `json:"providerRsLength"`
	ErrorCode                 int       `json:"errorCode"`
	ErrorMessage              string    `json:"errorMessage"`
	InternalMessage           string    `json:"internalMessage"`
	SupplierErrorMessage      string    `json:"providerErrorMessage"`
	SentToSupplier            bool      `json:"sentToProvider"`
	Node                      string    `json:"node"`
	Integration               string    `json:"integration"`
	ClientIp                  string    `json:"clientIp"`
	ClientName                string    `json:"clientName"`
	ClientCode                string    `json:"clientCode"`
	BookingChannel            string    `json:"bookingChannel"`
	NumResGuest               int       `json:"numResGuest"`
	RqBookingCode             []string  `json:"rqBookingCode"`
	TProviderID               string    `json:"bcProviderID"`
	TCheckInDate              string    `json:"bcCheckInDate"`
	TCheckOutDate             string    `json:"bcCheckOutDate"`
	TNumberOfRooms            int       `json:"bcNumberOfRooms"`
	TAdults                   int       `json:"bcAdults"`
	TChildren                 int       `json:"bcChildren"`
	TChildrenAges             string    `json:"bcChildrenAges"`
	TInfant                   int       `json:"bcInfant"`
	TCustomerCountry          string    `json:"bcCustomerCountry"`
	TRoomId                   []string  `json:"bcRoomId"`
	TBoardId                  []string  `json:"bcBoardId"`
	TBuyTotalPrice            float32   `json:"bcBuyTotalPrice"`
	TBuyPrice                 []float32 `json:"bcBuyPrice"`
	TSaleTotalPrice           float32   `json:"bcSaleTotalPrice"`
	TSalePrice                []float32 `json:"bcSalePrice"`
	TDistribution             string    `json:"bcDistribution"`
	TAdultAges                string    `json:"bcAdultAges"`
	TNumberOfService          int       `json:"bcNumberOfService"`

	RsBookingCode []string `json:"rsBookingCode"`
	ErrorCodeSt   string   `json:"errorCodeSt"`

	//New
	TMarket        string   `json:"bcMarket"`
	TGiRoomID      []string `json:"bcGiRoomID"`
	TGiRoomCode    []string `json:"bcGiRoomCode"`
	TGiRoomName    []string `json:"bcGiRoomName"`
	TPrvRoomCode   []string `json:"bcPrvRoomCode"`
	TPrvRoomName   []string `json:"bcPrvRoomName"`
	TIsDynamicRoom []bool   `json:"bcDynamicRoom"`
	TIsMappedRoom  []bool   `json:"bcMappedRoom"`
	TGiHotelCode   string   `json:"bcGiHotelCode"`
	TPrvHotelCode  string   `json:"bcPrvHotelCode"`
	BookedRoomName []string `json:"bookedRoomName"`
	IsRebook       bool     `json:"isRebook"`
}

type CancelLog struct {
	CancelType    string `json:"cancelType"`
	Error         string `json:"error"`
	PrimaryLangID string `json:"primaryLangID"`
	ProviderID    string `json:"providerID"`
	ProviderCode  string `json:"providerCode"`
	RequestorID   string `json:"requestorID"`
	Integration   string `json:"integration"`

	RqInternal  string `json:"rqInternal"`
	RqProvider  string `json:"rqProvider"`
	RqTimestamp int64  `json:"rqTimestamp"`
	RqType      string `json:"rqType"`
	ClientName  string `json:"clientName"`
	ClientCode  string `json:"clientCode"`

	RsInternal            string `json:"rsInternal"`
	RsProvider            string `json:"rsProvider"`
	Status                string `json:"status"`
	Success               string `json:"success"`
	TotalTime             int    `json:"totalTime"`
	TransactionIdentifier string `json:"transactionIdentifier"`
	Version               string `json:"version"`

	SupplierRsTime           int64 `json:"providerRsTime"`
	SupplierRsHttpStatusCode int   `json:"providerRsHttpStatusCode"`
	SupplierRsLength         int   `json:"providerRsLength"`

	ErrorCode       int    `json:"errorCode"`
	ErrorMessage    string `json:"errorMessage"`
	InternalMessage string `json:"internalMessage"`

	SupplierErrorMessage string `json:"providerErrorMessage"`
	SentToSupplier       bool   `json:"sentToProvider"`
	IsRebook             bool   `json:"isRebook"`
}

func (al AvailLog) ToJsonString() string {
	content, _ := json.Marshal(al)
	return string(content)
}

func (al AvailLog) CallType() string {
	return "Avail"
}

func (al *AvailLog) SetRqProvider(rqProvider string) {
	al.RqProvider = rqProvider
}

func (al *AvailLog) SetRsProvider(rsProvider string) {
	al.RsProvider = rsProvider
}

func (prl HotelResBookLog) ToJsonString() string {
	content, _ := json.Marshal(prl)
	return string(content)
}

func (prl HotelResBookLog) CallType() string {
	return "HotelResBook"
}

func (prl *HotelResBookLog) SetRqProvider(rqProvider string) {
	prl.RqProvider = rqProvider
}

func (prl *HotelResBookLog) SetRsProvider(rsProvider string) {
	prl.RsProvider = rsProvider
}

func (prl *HotelResBookLog) SetRsTime(ms int) {
	prl.TotalTime = ms
}

func (al *AvailLog) SetRsTime(ms int) {
	al.RsTime = ms
}

func (cfl HotelResCommitLog) ToJsonString() string {
	content, _ := json.Marshal(cfl)
	return string(content)
}

func (cfl HotelResCommitLog) CallType() string {
	return "HotelResCommit"
}

func (cfl *HotelResCommitLog) SetRqProvider(rqProvider string) {
	cfl.RqProvider = rqProvider
}

func (cfl *HotelResCommitLog) SetRsProvider(rsProvider string) {
	cfl.RsProvider = rsProvider
}

func (cfl *HotelResCommitLog) SetRsTime(ms int) {
	cfl.TotalTime = ms
}

func (cl CancelLog) ToJsonString() string {
	content, _ := json.Marshal(cl)
	return string(content)
}

func (cl CancelLog) CallType() string {
	return "Cancel"
}

func (cl *CancelLog) SetRqProvider(rqProvider string) {
	cl.RqProvider = rqProvider
}

func (cl *CancelLog) SetRsProvider(rsProvider string) {
	cl.RsProvider = rsProvider
}

func (cl *CancelLog) SetRsTime(ms int) {
	cl.TotalTime = ms
}

type GenericCallLog interface {
	ToJsonString() string
	CallType() string
	SetRsTime(int)
}

// ExternalXMLLog define la interfaz para establecer XMLs externos en los logs
type ExternalXMLLog interface {
	SetRqProvider(rqProvider string)
	SetRsProvider(rsProvider string)
}
