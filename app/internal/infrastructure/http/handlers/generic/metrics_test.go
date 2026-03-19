package generic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/require"
)

func TestIncrementCounter_IncrementsExistingAndNewKeys(t *testing.T) {
	countersMap = cmap.New()

	IncrementCounter("TOTAL", "AVAIL")
	IncrementCounter("TOTAL")
	IncrementCounter("AVAIL")

	totalRaw, ok := countersMap.Get("TOTAL")
	require.True(t, ok)
	totalCounter, ok := totalRaw.(metrics.Counter)
	require.True(t, ok)
	require.Equal(t, int64(2), totalCounter.Count())

	availRaw, ok := countersMap.Get("AVAIL")
	require.True(t, ok)
	availCounter, ok := availRaw.(metrics.Counter)
	require.True(t, ok)
	require.Equal(t, int64(2), availCounter.Count())
}

func TestProcessMetrics_ReturnsJSONWithCustomCounters(t *testing.T) {
	countersMap = cmap.New()
	IncrementCounter("TOTAL", "TOTAL", "AVAIL")

	gin.SetMode(gin.TestMode)
	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = httptest.NewRequest(http.MethodGet, "/metrics", nil)

	ProcessMetrics(c)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var payload map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &payload))

	customMetricsRaw, ok := payload["customMetrics"]
	require.True(t, ok)
	customMetrics, ok := customMetricsRaw.(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(2), customMetrics["TOTAL"])
	require.Equal(t, float64(1), customMetrics["AVAIL"])
	_, ok = payload["systemMetrics"]
	require.True(t, ok)
}
