package handlers

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"ws-int-httr/internal/infrastructure/registry"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type testAdminConfig struct {
	adminUser string
	adminPass string
}

func (c testAdminConfig) ServerPort() int                       { return 8080 }
func (c testAdminConfig) LogLevel() string                      { return "debug" }
func (c testAdminConfig) LogPath() string                       { return "/tmp" }
func (c testAdminConfig) BookingCodeVersion() string            { return "1.0.0" }
func (c testAdminConfig) AvailTo() int                          { return 3000 }
func (c testAdminConfig) AdminUsername() string                 { return c.adminUser }
func (c testAdminConfig) AdminPassword() string                 { return c.adminPass }
func (c testAdminConfig) DefaultEmail() string                  { return "" }
func (c testAdminConfig) DefaultPhone() string                  { return "" }
func (c testAdminConfig) DBHost() string                        { return "" }
func (c testAdminConfig) DBPort() int                           { return 0 }
func (c testAdminConfig) DBUser() string                        { return "" }
func (c testAdminConfig) DBPass() string                        { return "" }
func (c testAdminConfig) DBDriver() string                      { return "" }
func (c testAdminConfig) DBSID() string                         { return "" }
func (c testAdminConfig) DBFechRowCount() int                   { return 0 }
func (c testAdminConfig) ProviderName() string                  { return "" }
func (c testAdminConfig) ProviderCode() string                  { return "" }
func (c testAdminConfig) ProviderSearchURL() string             { return "" }
func (c testAdminConfig) ProviderQuoteURL() string              { return "" }
func (c testAdminConfig) ProviderBookURL() string               { return "" }
func (c testAdminConfig) ProviderCancelURL() string             { return "" }
func (c testAdminConfig) ProviderAuthToken() string             { return "" }
func (c testAdminConfig) ProviderAuthForChannel(string) string  { return "" }
func (c testAdminConfig) ProviderTimeoutMs() int                { return 0 }
func (c testAdminConfig) ProviderIdList() []int                 { return []int{5542} }
func (c testAdminConfig) ProviderMaxRoomsPerOccupancy() int     { return 0 }

func TestReloadCacheHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 500 when config is missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ws-int-httr/admin/cache/reload", nil)
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		c.Request = req

		ReloadCacheHandler(c)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		require.JSONEq(t, `{"status":"KO"}`, rr.Body.String())
	})

	registry.Register("config", testAdminConfig{adminUser: "admin", adminPass: "secret"})

	t.Run("returns 401 and challenge when basic auth is missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ws-int-httr/admin/cache/reload", nil)
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		c.Request = req

		ReloadCacheHandler(c)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
		require.Equal(t, `Basic realm="Admin Area"`, rr.Header().Get("WWW-Authenticate"))
		require.JSONEq(t, `{"status":"KO"}`, rr.Body.String())
	})

	t.Run("returns 401 when credentials are invalid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ws-int-httr/admin/cache/reload", nil)
		req.Header.Set("Authorization", basicAuthHeader("bad-user", "bad-pass"))
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		c.Request = req

		ReloadCacheHandler(c)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
		require.JSONEq(t, `{"status":"KO"}`, rr.Body.String())
	})

	t.Run("returns 500 when repository is missing after valid auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ws-int-httr/admin/cache/reload", nil)
		req.Header.Set("Authorization", basicAuthHeader("admin", "secret"))
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		c.Request = req

		ReloadCacheHandler(c)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		require.JSONEq(t, `{"status":"KO"}`, rr.Body.String())
	})
}

func basicAuthHeader(user, pass string) string {
	token := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	return "Basic " + token
}
