package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"ws-int-httr/internal/domain/log_domain"
	"ws-int-httr/internal/infrastructure/session"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestNewFileStructuredLogger_CreatesLogstashFiles(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()

	sl, err := NewFileStructuredLogger(baseDir)
	require.NoError(t, err)
	require.NotNil(t, sl)

	logstashDir := filepath.Join(baseDir, "logstash")
	require.DirExists(t, logstashDir)
	require.FileExists(t, filepath.Join(logstashDir, "HotelAvailJSON.log"))
	require.FileExists(t, filepath.Join(logstashDir, "HotelResBookJSON.log"))
	require.FileExists(t, filepath.Join(logstashDir, "HotelResCommitJSON.log"))
	require.FileExists(t, filepath.Join(logstashDir, "CancelJSON.log"))
}

func TestFileStructuredLogger_LogAvail_WritesJSONLine(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()

	sl, err := NewFileStructuredLogger(baseDir)
	require.NoError(t, err)

	fileLogger, ok := sl.(*FileStructuredLogger)
	require.True(t, ok)

	entry := &log_domain.AvailLog{
		EchoToken:    "echo-123",
		RqType:       "GIOTAHotelAvailRQ",
		ProviderCode: "OPENB2B",
		Success:      "true",
	}

	fileLogger.LogAvail(entry)

	line := readLastLine(t, filepath.Join(baseDir, "logstash", "HotelAvailJSON.log"))
	require.NotEmpty(t, line)

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(line), &payload))
	require.Equal(t, "echo-123", payload["echoToken"])
	require.Equal(t, "GIOTAHotelAvailRQ", payload["rqType"])
	require.Equal(t, "OPENB2B", payload["providerCode"])
	require.Equal(t, "true", payload["success"])
}

func TestFileStructuredLogger_LogCall_RoutesByType(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()

	sl, err := NewFileStructuredLogger(baseDir)
	require.NoError(t, err)

	fileLogger, ok := sl.(*FileStructuredLogger)
	require.True(t, ok)

	fileLogger.LogCall(&log_domain.CancelLog{
		RqType:       "GIOTAHotelCancelRQ",
		ProviderCode: "OPENB2C",
		Success:      "OK",
	})

	line := readLastLine(t, filepath.Join(baseDir, "logstash", "CancelJSON.log"))
	require.NotEmpty(t, line)
	require.Contains(t, line, `"rqType":"GIOTAHotelCancelRQ"`)
	require.Contains(t, line, `"providerCode":"OPENB2C"`)
	require.Contains(t, line, `"success":"OK"`)
}

func TestNormalizeRequestType(t *testing.T) {
	t.Parallel()

	require.Equal(t, "GIOTAHotelAvail", normalizeRequestType("GIOTAHotelAvailRQ"))
	require.Equal(t, "GIOTAHotelResBook", normalizeRequestType("GIOTAHotelResRQ_Book"))
	require.Equal(t, "GIOTAHotelResCommit", normalizeRequestType("GIOTAHotelResRQ_COMMIT"))
	require.Equal(t, "GIOTACancel", normalizeRequestType("GIOTACancelRQ"))
	require.Equal(t, "custom", normalizeRequestType("custom"))
}

func TestLoggerMiddleware_WritesRequestLogAndSkipsHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	session.Clear()
	t.Cleanup(session.Clear)

	var buf bytes.Buffer
	originalLog := Log
	Log = log.New(&buf, "", 0)
	t.Cleanup(func() { Log = originalLog })

	r := gin.New()
	r.Use(func(c *gin.Context) {
		session.New(c.Request.Context())
		session.FromContext().Data().EchoToken = "echo-1"
		session.FromContext().Set("requestType", "GIOTAHotelAvailRQ")
		c.Next()
		session.Clear()
	})
	r.Use(Logger())
	r.GET("/ws-int-httr/ws", func(c *gin.Context) { c.String(http.StatusCreated, "ok") })
	r.GET("/ws-int-httr/health", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ws-int-httr/ws", nil))
	require.Equal(t, http.StatusCreated, rr.Code)
	require.Contains(t, buf.String(), `EchoToken=echo-1`)
	require.Contains(t, buf.String(), `path=/ws-int-httr/ws`)
	require.Contains(t, buf.String(), `requestType=GIOTAHotelAvail`)
	require.Contains(t, buf.String(), `statusCode=201`)

	before := buf.Len()
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ws-int-httr/health", nil))
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, before, buf.Len())
}

func readLastLine(t *testing.T, path string) string {
	t.Helper()

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	last := ""
	for scanner.Scan() {
		last = scanner.Text()
	}
	require.NoError(t, scanner.Err())
	return last
}
