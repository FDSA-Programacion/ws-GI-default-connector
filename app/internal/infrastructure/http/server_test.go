package httpserver

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"ws-int-httr/internal/infrastructure/http/handlers"
	"ws-int-httr/internal/infrastructure/registry"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type testServerConfig struct {
	adminUser string
	adminPass string
}

func (c testServerConfig) ServerPort() int                      { return 8080 }
func (c testServerConfig) LogLevel() string                     { return "debug" }
func (c testServerConfig) LogPath() string                      { return "/tmp" }
func (c testServerConfig) BookingCodeVersion() string           { return "1.0.0" }
func (c testServerConfig) AvailTo() int                         { return 3000 }
func (c testServerConfig) AdminUsername() string                { return c.adminUser }
func (c testServerConfig) AdminPassword() string                { return c.adminPass }
func (c testServerConfig) DefaultEmail() string                 { return "" }
func (c testServerConfig) DefaultPhone() string                 { return "" }
func (c testServerConfig) DBHost() string                       { return "" }
func (c testServerConfig) DBPort() int                          { return 0 }
func (c testServerConfig) DBUser() string                       { return "" }
func (c testServerConfig) DBPass() string                       { return "" }
func (c testServerConfig) DBDriver() string                     { return "" }
func (c testServerConfig) DBSID() string                        { return "" }
func (c testServerConfig) DBFechRowCount() int                  { return 0 }
func (c testServerConfig) ProviderName() string                 { return "" }
func (c testServerConfig) ProviderCode() string                 { return "" }
func (c testServerConfig) ProviderSearchURL() string            { return "" }
func (c testServerConfig) ProviderQuoteURL() string             { return "" }
func (c testServerConfig) ProviderBookURL() string              { return "" }
func (c testServerConfig) ProviderCancelURL() string            { return "" }
func (c testServerConfig) ProviderAuthToken() string            { return "" }
func (c testServerConfig) ProviderAuthForChannel(string) string { return "" }
func (c testServerConfig) ProviderTimeoutMs() int               { return 0 }
func (c testServerConfig) ProviderIdList() []int                { return []int{5542} }
func (c testServerConfig) ProviderMaxRoomsPerOccupancy() int    { return 0 }

func TestServerRegistersAdminCacheReloadRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := testServerConfig{adminUser: "admin", adminPass: "secret"}
	registry.Register("config", cfg)

	srv := New(cfg, &handlers.BookingHTTPHandler{})

	req := httptest.NewRequest(http.MethodGet, "/ws-int-httr/admin/cache/reload", nil)
	req.Header.Set("Authorization", basicAuthHeader("admin", "secret"))
	rr := httptest.NewRecorder()

	srv.engine.ServeHTTP(rr, req)

	require.NotEqual(t, http.StatusNotFound, rr.Code)
	require.Equal(t, http.StatusInternalServerError, rr.Code)
	require.JSONEq(t, `{"status":"KO"}`, rr.Body.String())
}

func basicAuthHeader(user, pass string) string {
	token := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	return "Basic " + token
}
