package serializer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type serializerSample struct {
	Name  string `json:"name" xml:"name"`
	Count int    `json:"count" xml:"count"`
}

type serializerXMLRoot struct {
	XMLName struct{}         `xml:"sample"`
	Payload serializerSample `xml:",inline"`
}

func TestGoSerializer_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	s := NewGoSerializer()
	in := serializerSample{Name: "alpha", Count: 3}

	data, err := s.ToJSON(in)
	require.NoError(t, err)
	require.JSONEq(t, `{"name":"alpha","count":3}`, string(data))

	var out serializerSample
	err = s.FromJSON(data, &out)
	require.NoError(t, err)
	require.Equal(t, in, out)
}

func TestGoSerializer_XMLRoundTrip(t *testing.T) {
	t.Parallel()

	s := NewGoSerializer()
	in := serializerXMLRoot{Payload: serializerSample{Name: "beta", Count: 7}}

	data, err := s.ToXML(in)
	require.NoError(t, err)
	require.Contains(t, string(data), "<sample>")
	require.Contains(t, string(data), "<name>beta</name>")
	require.Contains(t, string(data), "<count>7</count>")

	var out serializerXMLRoot
	err = s.FromXML(data, &out)
	require.NoError(t, err)
	require.Equal(t, in.Payload, out.Payload)
}

func TestGoSerializer_InvalidJSONAndXML(t *testing.T) {
	t.Parallel()

	s := NewGoSerializer()

	var j serializerSample
	require.Error(t, s.FromJSON([]byte(`{"name":`), &j))

	var x serializerXMLRoot
	require.Error(t, s.FromXML([]byte(`<sample><name>x</name>`), &x))
}
