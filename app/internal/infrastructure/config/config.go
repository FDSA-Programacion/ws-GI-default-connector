package config

type AppConfig interface {
	ServerPort() int
	LogLevel() string
	LogPath() string
	DBConfig
	ProviderConfig
	BookingCodeVersion() string
	AvailTo() int
	AdminUsername() string
	AdminPassword() string
	DefaultEmail() string
	DefaultPhone() string
}

type DBConfig interface {
	DBHost() string
	DBPort() int
	DBUser() string
	DBPass() string
	DBDriver() string
	DBSID() string
	DBFechRowCount() int
}

type ProviderConfig interface {
	ProviderName() string
	ProviderCode() string
	ProviderSearchURL() string
	ProviderQuoteURL() string
	ProviderBookURL() string
	ProviderCancelURL() string
	ProviderAuthToken() string
	ProviderAuthForChannel(channelCode string) string
	ProviderTimeoutMs() int
	ProviderIdList() []int
	ProviderMaxRoomsPerOccupancy() int
	DefaultEmail() string
	DefaultPhone() string
}
