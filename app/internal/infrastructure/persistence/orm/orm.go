package orm

type DBRegimen struct {
	ID           string `mapstructure:"ID"`
	RegimenID    string `mapstructure:"REGIMENID"`
	Codigo       string `mapstructure:"CODIGO"`
	ProviderCode string `mapstructure:"PROVIDER_CODE"`
	Descripcion  string `mapstructure:"DESCRIPCION"`
}

type DBRegimenTraduccion struct {
	Key         string `mapstructure:"KEY"`
	IDRegimen   int    `mapstructure:"IDREGIMEN"`
	CodRegimen  string `mapstructure:"CODREGIMEN"`
	CodIdioma   string `mapstructure:"CODIDIOMA"`
	Descripcion string `mapstructure:"DESCRIPCION"`
}

type DBRoomMapping struct {
	GIRoomCode      string `mapstructure:"GIROOMCODE"`
	GIRoomName      string `mapstructure:"GIROOMNAME"`
	GIRoomID        string `mapstructure:"GIROOMID"`
	Id              string `mapstructure:"IDENTIFIER"`
	IntegrationCode string `mapstructure:"INTEGRATIONCODE"`
	IntegrationID   string `mapstructure:"INTEGRATIONID"`
	PrvRoomCode     string `mapstructure:"PRVROOMCODE"`
	PrvRoomName     string `mapstructure:"PRVROOMNAME"`
}

type DBRoomDescription struct {
	Key           string `mapstructure:"KEY"`
	Identifier    string `mapstructure:"IDENTIFIER"`
	CodHabitacion string `mapstructure:"CODHABITACION"`
	CodIdioma     string `mapstructure:"CODIDIOMA"`
	Descripcion   string `mapstructure:"DESCRIPCION"`
}

type DBAlojamiento struct {
	AreaCode          string `mapstructure:"AREACODE"`
	AreaName          string `mapstructure:"AREANAME"`
	Category          string `mapstructure:"CATEGORY"`
	CityID            int    `mapstructure:"CITYID"`
	CityName          string `mapstructure:"CITYNAME"`
	HotelCode         string `mapstructure:"HOTELCODE"`
	HotelID           string `mapstructure:"HOTELID"`
	HotelName         string `mapstructure:"HOTELNAME"`
	IntegrationCode   string `mapstructure:"INTEGRATIONCODE"`
	IntegrationID     int    `mapstructure:"INTEGRATIONID"`
	PropertyClassCode string `mapstructure:"PROPERTYCLASSCODE"`
	ProviderHotelID   string `mapstructure:"PROVIDERHOTELID"`
	Rating            string `mapstructure:"RATING"`
}
