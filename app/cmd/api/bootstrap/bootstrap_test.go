package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testAppConfig struct{}

func (testAppConfig) ServerPort() int                       { return 8080 }
func (testAppConfig) LogLevel() string                      { return "debug" }
func (testAppConfig) LogPath() string                       { return "/tmp/logs" }
func (testAppConfig) BookingCodeVersion() string            { return "1.2.3" }
func (testAppConfig) AvailTo() int                          { return 3000 }
func (testAppConfig) AdminUsername() string                 { return "admin" }
func (testAppConfig) AdminPassword() string                 { return "secret" }
func (testAppConfig) DefaultEmail() string                  { return "ops@example.test" }
func (testAppConfig) DefaultPhone() string                  { return "+34000000000" }
func (testAppConfig) DBHost() string                        { return "db.local" }
func (testAppConfig) DBPort() int                           { return 1521 }
func (testAppConfig) DBUser() string                        { return "user" }
func (testAppConfig) DBPass() string                        { return "pass" }
func (testAppConfig) DBDriver() string                      { return "godror" }
func (testAppConfig) DBSID() string                         { return "ora11g" }
func (testAppConfig) DBFechRowCount() int                   { return 1000 }
func (testAppConfig) ProviderName() string                  { return "hoteltrader" }
func (testAppConfig) ProviderCode() string                  { return "HTTR" }
func (testAppConfig) ProviderSearchURL() string             { return "https://example.test/search" }
func (testAppConfig) ProviderQuoteURL() string              { return "https://example.test/quote" }
func (testAppConfig) ProviderBookURL() string               { return "https://example.test/book" }
func (testAppConfig) ProviderCancelURL() string             { return "https://example.test/cancel" }
func (testAppConfig) ProviderAuthToken() string             { return "token" }
func (testAppConfig) ProviderAuthForChannel(string) string  { return "channel-token" }
func (testAppConfig) ProviderTimeoutMs() int                { return 5000 }
func (testAppConfig) ProviderIdList() []int                 { return []int{5542, 6298} }
func (testAppConfig) ProviderMaxRoomsPerOccupancy() int     { return 6 }

func TestMaskConfig(t *testing.T) {
	t.Parallel()

	got := maskConfig(testAppConfig{})

	require.Equal(t, map[string]interface{}{
		"serverPort":         8080,
		"logLevel":           "debug",
		"logPath":            "/tmp/logs",
		"bookingCodeVersion": "1.2.3",
		"providerSearchURL":  "https://example.test/search",
		"providerQuoteURL":   "https://example.test/quote",
		"providerBookURL":    "https://example.test/book",
		"providerCancelURL":  "https://example.test/cancel",
	}, got)
}

func TestGetHostName_ReturnsValue(t *testing.T) {
	t.Parallel()

	host := getHostName()
	require.NotEmpty(t, host)
}
