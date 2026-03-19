package httr_client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"ws-int-httr/internal/domain/log_domain"
	"ws-int-httr/internal/infrastructure/mapping/hoteltrader"
	"ws-int-httr/internal/infrastructure/serializer"
	"ws-int-httr/internal/infrastructure/session"

	"github.com/stretchr/testify/require"
)

type fakeProviderConfig struct {
	searchURL string
	quoteURL  string
	bookURL   string
	cancelURL string
	authByCh  map[string]string
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func (f fakeProviderConfig) ProviderName() string                         { return "httr" }
func (f fakeProviderConfig) ProviderCode() string                         { return "HTTR" }
func (f fakeProviderConfig) ProviderSearchURL() string                    { return f.searchURL }
func (f fakeProviderConfig) ProviderQuoteURL() string                     { return f.quoteURL }
func (f fakeProviderConfig) ProviderBookURL() string                      { return f.bookURL }
func (f fakeProviderConfig) ProviderCancelURL() string                    { return f.cancelURL }
func (f fakeProviderConfig) ProviderAuthToken() string                    { return "" }
func (f fakeProviderConfig) ProviderTimeoutMs() int                       { return 1000 }
func (f fakeProviderConfig) ProviderIdList() []int                        { return nil }
func (f fakeProviderConfig) ProviderMaxRoomsPerOccupancy() int            { return 0 }
func (f fakeProviderConfig) DefaultEmail() string                         { return "" }
func (f fakeProviderConfig) DefaultPhone() string                         { return "" }
func (f fakeProviderConfig) ProviderAuthForChannel(channelCode string) string {
	return f.authByCh[strings.TrimSpace(strings.ToUpper(channelCode))]
}

func TestGetSessionKeyPrefix(t *testing.T) {
	t.Parallel()
	require.Equal(t, "preBook", getSessionKeyPrefix("prebook"))
	require.Equal(t, "avail", getSessionKeyPrefix("avail"))
}

func TestSetExternalRequestResponseInLog_UpdatesEndpointLog(t *testing.T) {
	session.Clear()
	t.Cleanup(session.Clear)
	session.New(context.Background())

	availLog := &log_domain.AvailLog{}
	session.FromContext().Set("availLog", availLog)

	// Sin debug: no debe guardar ninguno en avail
	setExternalRequestResponseInLog("avail", `{"a":1}`, `{"ok":true}`)

	require.Equal(t, "", availLog.RqProvider)
	require.Equal(t, `{"ok":true}`, availLog.RsProvider)

	// Con debug: solo rqProvider, rsProvider nunca
	session.FromContext().Set("debug", "g2018i")
	setExternalRequestResponseInLog("avail", `{"a":2}`, `{"ok":false}`)
	require.Equal(t, `{"a":2}`, availLog.RqProvider)
	require.Equal(t, `{"ok":false}`, availLog.RsProvider)
}

func TestHttrClient_getProviderURL_SelectsEndpointAndFallback(t *testing.T) {
	t.Parallel()

	cfg := fakeProviderConfig{
		searchURL: "https://x/search",
		quoteURL:  "https://x/quote",
		bookURL:   "https://x/book",
		cancelURL: "https://x/cancel",
	}
	c := &HttrClientImpl{config: cfg}

	require.Equal(t, cfg.searchURL, c.getProviderURL("avail"))
	require.Equal(t, cfg.quoteURL, c.getProviderURL("prebook"))
	require.Equal(t, cfg.bookURL, c.getProviderURL("book"))
	require.Equal(t, cfg.cancelURL, c.getProviderURL("cancel"))
	require.Equal(t, cfg.searchURL, c.getProviderURL("unknown"))
}

func TestExecuteProviderCall_Success_DecodesJSONAndSetsAuthHeader(t *testing.T) {
	session.Clear()
	t.Cleanup(session.Clear)
	session.New(context.Background())
	preBookLog := &log_domain.HotelResBookLog{}
	session.FromContext().Set("preBookLog", preBookLog)

	var gotAuth, gotContentType, gotBody string

	cfg := fakeProviderConfig{
		searchURL: "http://provider.test/search",
		quoteURL:  "http://provider.test/quote",
		bookURL:   "http://provider.test/book",
		cancelURL: "http://provider.test/cancel",
		authByCh:  map[string]string{"OPENB2B": "abc123"},
	}
	client := &HttrClientImpl{
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			gotAuth = r.Header.Get("Authorization")
			gotContentType = r.Header.Get("Content-Type")
			body, _ := ioReadAll(r)
			gotBody = string(body)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"data":{"getPropertiesByIds":{"properties":[]}}}`)),
			}, nil
		})},
		serializer: serializer.NewGoSerializer(),
		config:     cfg,
	}

	var resp hoteltrader.ProviderAvailRS
	err := client.executeProviderCall("prebook", []byte(`{"query":"x"}`), &resp, "openb2b")
	require.NoError(t, err)
	require.Equal(t, "Basic abc123", gotAuth)
	require.Equal(t, "application/json", gotContentType)
	require.JSONEq(t, `{"query":"x"}`, gotBody)
	require.NotNil(t, resp.Data.GetPropertiesByIds.Properties)
	require.Equal(t, `{"query":"x"}`, preBookLog.RqProvider)
	require.Contains(t, preBookLog.RsProvider, `"data"`)
}

func TestExecuteProviderCall_ReturnsHTTPError(t *testing.T) {
	client := &HttrClientImpl{
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Body:       io.NopCloser(strings.NewReader("bad gateway\n")),
			}, nil
		})},
		config: fakeProviderConfig{
			searchURL: "http://provider.test/search",
			quoteURL:  "http://provider.test/quote",
			bookURL:   "http://provider.test/book",
			cancelURL: "http://provider.test/cancel",
		},
	}

	var resp hoteltrader.ProviderAvailRS
	err := client.executeProviderCall("avail", []byte(`{}`), &resp, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "estado HTTP 502")
}

func TestExecuteProviderCall_ReturnsGraphQLError(t *testing.T) {
	client := &HttrClientImpl{
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"errors":[{"message":"boom"}],"data":{"getPropertiesByIds":{"properties":[]}}}`)),
			}, nil
		})},
		config: fakeProviderConfig{
			searchURL: "http://provider.test/search",
			quoteURL:  "http://provider.test/quote",
			bookURL:   "http://provider.test/book",
			cancelURL: "http://provider.test/cancel",
		},
	}

	var resp hoteltrader.ProviderAvailRS
	err := client.executeProviderCall("avail", []byte(`{}`), &resp, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "errores GraphQL")
	require.Contains(t, err.Error(), "boom")
}

func ioReadAll(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
