package provider_client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"ws-int-httr/internal/domain/log_domain"
	"ws-int-httr/internal/infrastructure/mapping/provider"
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

func (f fakeProviderConfig) ProviderName() string              { return "provider" }
func (f fakeProviderConfig) ProviderCode() string              { return "PROVIDERCODE" }
func (f fakeProviderConfig) ProviderSearchURL() string         { return f.searchURL }
func (f fakeProviderConfig) ProviderQuoteURL() string          { return f.quoteURL }
func (f fakeProviderConfig) ProviderBookURL() string           { return f.bookURL }
func (f fakeProviderConfig) ProviderCancelURL() string         { return f.cancelURL }
func (f fakeProviderConfig) ProviderAuthToken() string         { return "" }
func (f fakeProviderConfig) ProviderTimeoutMs() int            { return 1000 }
func (f fakeProviderConfig) ProviderIdList() []int             { return nil }
func (f fakeProviderConfig) ProviderMaxRoomsPerOccupancy() int { return 0 }
func (f fakeProviderConfig) DefaultEmail() string              { return "" }
func (f fakeProviderConfig) DefaultPhone() string              { return "" }
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

	// Sin debug: no debe guardar rqProvider en avail
	setExternalRequestResponseInLog("avail", `<rq>a</rq>`, `<ok>true</ok>`)

	require.Equal(t, "", availLog.RqProvider)
	require.Equal(t, `<ok>true</ok>`, availLog.RsProvider)

	// Con debug: debe guardar el request XML y mantener el response XML
	session.FromContext().Set("debug", "g2018i")
	setExternalRequestResponseInLog("avail", `<rq>b</rq>`, `<ok>false</ok>`)
	require.Equal(t, `<rq>b</rq>`, availLog.RqProvider)
	require.Equal(t, `<ok>false</ok>`, availLog.RsProvider)
}

func TestHttrClient_getProviderURL_SelectsEndpointAndFallback(t *testing.T) {
	t.Parallel()

	cfg := fakeProviderConfig{
		searchURL: "https://x/search",
		quoteURL:  "https://x/quote",
		bookURL:   "https://x/book",
		cancelURL: "https://x/cancel",
	}
	c := &ProviderClientImpl{config: cfg}

	require.Equal(t, cfg.searchURL, c.getProviderURL("avail"))
	require.Equal(t, cfg.quoteURL, c.getProviderURL("prebook"))
	require.Equal(t, cfg.bookURL, c.getProviderURL("book"))
	require.Equal(t, cfg.cancelURL, c.getProviderURL("cancel"))
	require.Equal(t, cfg.searchURL, c.getProviderURL("unknown"))
}

func TestExecuteProviderCall_Success_DecodesXMLAndSetsAuthHeader(t *testing.T) {
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
	client := &ProviderClientImpl{
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			gotAuth = r.Header.Get("Authorization")
			gotContentType = r.Header.Get("Content-Type")
			body, _ := ioReadAll(r)
			gotBody = string(body)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/xml"}},
				Body:       io.NopCloser(strings.NewReader(`<ProviderAvailResponse></ProviderAvailResponse>`)),
			}, nil
		})},
		serializer: serializer.NewGoSerializer(),
		config:     cfg,
	}

	var resp provider.ProviderAvailResponse
	err := client.executeProviderCall("prebook", []byte(`<ProviderPrebookRequest></ProviderPrebookRequest>`), &resp, "openb2b")
	require.NoError(t, err)
	require.Equal(t, "Basic abc123", gotAuth)
	require.Equal(t, "application/xml", gotContentType)
	require.Equal(t, `<ProviderPrebookRequest></ProviderPrebookRequest>`, gotBody)
	require.Equal(t, `<ProviderPrebookRequest></ProviderPrebookRequest>`, preBookLog.RqProvider)
	require.Contains(t, preBookLog.RsProvider, `<ProviderAvailResponse>`)
}

func TestExecuteProviderCall_ReturnsHTTPError(t *testing.T) {
	client := &ProviderClientImpl{
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

	var resp provider.ProviderAvailResponse
	err := client.executeProviderCall("avail", []byte(`<ProviderAvailRequest></ProviderAvailRequest>`), &resp, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "estado HTTP 502")
}

func TestExecuteProviderCall_ReturnsXMLDeserializationError(t *testing.T) {
	client := &ProviderClientImpl{
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/xml"}},
				Body:       io.NopCloser(strings.NewReader(`<ProviderAvailResponse>`)),
			}, nil
		})},
		serializer: serializer.NewGoSerializer(),
		config: fakeProviderConfig{
			searchURL: "http://provider.test/search",
			quoteURL:  "http://provider.test/quote",
			bookURL:   "http://provider.test/book",
			cancelURL: "http://provider.test/cancel",
		},
	}

	var resp provider.ProviderAvailResponse
	err := client.executeProviderCall("avail", []byte(`<ProviderAvailRequest></ProviderAvailRequest>`), &resp, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "error al deserializar XML")
}

func ioReadAll(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
