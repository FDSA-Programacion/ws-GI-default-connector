package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ws-int-httr/internal/infrastructure/session"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSessionMiddleware_InitializesAndClearsSession(t *testing.T) {
	t.Cleanup(session.Clear)
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(SessionMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		s := session.FromContext()
		require.NotNil(t, s)
		require.NotNil(t, s.Data())
		require.NotNil(t, s.Data().ContextData)
		require.NotNil(t, s.Data().MapHotelSession)

		s.Data().Debug = "in-request"
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
	_, ok := session.FromContext().Get("debug")
	require.False(t, ok)
}

func TestCustomMiddleware_CapturesRawBodyRestoresBodyAndSetsStartTime(t *testing.T) {
	t.Cleanup(session.Clear)
	gin.SetMode(gin.TestMode)

	const body = `{"rqType":"GIOTAHotelAvailRQ","echoToken":"abc"}`

	r := gin.New()
	r.Use(SessionMiddleware(), CustomMiddleware())
	r.POST("/ws-int-httr/ws", func(c *gin.Context) {
		s := session.FromContext()
		require.NotNil(t, s)
		require.Equal(t, body, string(s.Data().RequestRawData))
		require.False(t, s.Data().StartTime.IsZero())

		downstreamBody, err := c.GetRawData()
		require.NoError(t, err)
		require.Equal(t, body, string(downstreamBody))

		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/ws-int-httr/ws", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}
