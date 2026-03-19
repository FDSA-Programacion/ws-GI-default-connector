package session

import (
	"context"
	"testing"
	"time"

	"ws-int-httr/internal/domain/log_domain"

	"github.com/stretchr/testify/require"
)

func TestSessionLifecycle_NewFromContextClear(t *testing.T) {
	Clear()
	t.Cleanup(Clear)

	s0 := FromContext()
	require.NotNil(t, s0)
	require.NotNil(t, s0.Data())
	require.NotNil(t, s0.Data().ContextData)
	require.NotNil(t, s0.Data().MapHotelSession)

	s1 := New(context.Background())
	require.NotNil(t, s1)

	s2 := FromContext()
	require.Same(t, s1, s2)

	s1.Data().Debug = "debug-token"
	got, ok := s2.Get("debug")
	require.True(t, ok)
	require.Equal(t, "debug-token", got)

	Clear()

	s3 := FromContext()
	require.NotNil(t, s3)
	require.NotSame(t, s1, s3)
	_, ok = s3.Get("debug")
	require.False(t, ok)
}

func TestSessionSetGet_BasicFieldsLogsAndBookResponse(t *testing.T) {
	Clear()
	t.Cleanup(Clear)

	s := New(context.Background())

	now := time.Now().Truncate(time.Second)
	raw := []byte(`{"a":1}`)
	ctxData := map[string]interface{}{"k": "v"}
	hotelMap := map[string]interface{}{"h": 1}
	availLog := &log_domain.AvailLog{EchoToken: "e1"}
	preBookLog := &log_domain.HotelResBookLog{EchoToken: "e2"}
	bookLog := &log_domain.HotelResCommitLog{EchoToken: "e3"}
	cancelLog := &log_domain.CancelLog{RqType: "cancel"}

	s.Set("requestType", "GIOTAHotelAvail")
	s.Set("requestRawData", raw)
	s.Set("startTime", now)
	s.Set("debug", "dbg")
	s.Set("providerCode", 5542)
	s.Set("contextData", ctxData)
	s.Set("mapHotelSession", hotelMap)
	s.Set("availLog", availLog)
	s.Set("preBookLog", preBookLog)
	s.Set("bookLog", bookLog)
	s.Set("cancelLog", cancelLog)
	s.Set("availLogStartTime", now)
	s.Set("preBookLogStartTime", now)
	s.Set("bookLogStartTime", now)
	s.Set("cancelLogStartTime", now)
	s.Set("bookResponse", map[string]string{"status": "ok"})

	assertGet(t, s, "requestType", "GIOTAHotelAvail")
	assertGet(t, s, "requestRawData", raw)
	assertGet(t, s, "debug", "dbg")
	assertGet(t, s, "providerCode", 5542)
	assertGet(t, s, "contextData", ctxData)
	assertGet(t, s, "mapHotelSession", hotelMap)
	assertGet(t, s, "availLog", availLog)
	assertGet(t, s, "preBookLog", preBookLog)
	assertGet(t, s, "bookLog", bookLog)
	assertGet(t, s, "cancelLog", cancelLog)

	start, ok := s.Get("startTime")
	require.True(t, ok)
	require.Equal(t, now, start)

	assertGet(t, s, "bookResponse", map[string]string{"status": "ok"})

	require.Equal(t, "dbg", s.GetString("debug"))
	require.Equal(t, "", s.GetString("providerCode"))
	require.Equal(t, "", s.GetString("missing"))
}

func TestSessionSetGet_SupplierMetrics_Object(t *testing.T) {
	Clear()
	t.Cleanup(Clear)

	s := New(context.Background())
	metrics := &EndpointMetrics{
		RsTime:         321,
		HttpStatusCode: 503,
		RsLength:       789,
		ErrorMessage:   "provider timeout",
	}

	s.Set("supplierMetrics", metrics)

	v, ok := s.Get("supplierMetrics")
	require.True(t, ok)
	require.Same(t, metrics, v)
}

func TestSessionGet_UnknownKey(t *testing.T) {
	Clear()
	t.Cleanup(Clear)

	s := New(context.Background())
	v, ok := s.Get("doesNotExist")
	require.False(t, ok)
	require.Nil(t, v)
}

func assertGet(t *testing.T, s *Session, key string, expected any) {
	t.Helper()
	v, ok := s.Get(key)
	require.True(t, ok, "expected key %s to exist", key)
	require.Equal(t, expected, v)
}
