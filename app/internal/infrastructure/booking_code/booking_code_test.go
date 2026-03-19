package bookingcode

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInternalBookingCode_SerializeDeserialize_RoundTrip(t *testing.T) {
	t.Parallel()

	checkIn := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 2, 5, 0, 0, 0, 0, time.UTC)

	original := &InternalBookingCode{
		RoomCounter:        "01",
		EchoToken:          "echo|token",
		Version:            "1.0.0",
		ProviderCode:       "5542",
		GiHotelCode:        "GI123",
		PrvHotelCode:       "PRV456",
		CheckInDate:        checkIn,
		CheckOutDate:       checkOut,
		GIRoomID:           "RID|01",
		GIRoomCode:         "RC01",
		GIRoomName:         "Room|Name",
		PrvRoomCode:        "P01",
		PrvRoomName:        "Provider|Room",
		IsDynamicRoom:      1,
		NumberOfService:    2,
		NumberOfRooms:      1,
		Distribution:       "2-0",
		Adults:             2,
		Children:           1,
		AdultAges:          "30,35",
		ChildrenAges:       "7",
		Infant:             0,
		CustomerCountry:    "ES",
		Market:             "ES",
		BoardId:            "BB",
		BuyTotalPrice:      123.45,
		BuyPrice:           120.00,
		BuyCurrency:        "EUR",
		SaleCurrency:       "USD",
		ExchangeRateID:     10,
		ExchangeRateValue:  1.08,
		MarkupID:           20,
		Markup:             3.21,
		AdditionalMarkupID: 21,
		AdditionalMarkup:   1.11,
		RetailAmount:       "130.00",
		RateDistribution:   "A",
		RateChildAges:      "7",
		RateCode:           "RATE1",
		IsSpecificRate:     1,
		NrRate:             7,
		InvBlockCode:       "INV|BLOCK",
		ExtraParams:        "rIdx<:=:>1",
		PreBookCode:        "PB|123",
		Lang:               "es",
		Id:                 "ID|789",
	}

	serialized := original.Serialize()
	require.NotEmpty(t, serialized)

	var decoded InternalBookingCode
	decoded.Deserialize(serialized)

	require.Equal(t, original.RoomCounter, decoded.RoomCounter)
	require.Equal(t, original.EchoToken, decoded.EchoToken)
	require.Equal(t, original.Version, decoded.Version)
	require.Equal(t, original.ProviderCode, decoded.ProviderCode)
	require.Equal(t, original.GiHotelCode, decoded.GiHotelCode)
	require.Equal(t, original.PrvHotelCode, decoded.PrvHotelCode)
	require.True(t, original.CheckInDate.Equal(decoded.CheckInDate))
	require.True(t, original.CheckOutDate.Equal(decoded.CheckOutDate))
	require.Equal(t, original.GIRoomID, decoded.GIRoomID)
	require.Equal(t, original.GIRoomName, decoded.GIRoomName)
	require.Equal(t, original.PrvRoomName, decoded.PrvRoomName)
	require.Equal(t, original.NumberOfService, decoded.NumberOfService)
	require.Equal(t, original.NumberOfRooms, decoded.NumberOfRooms)
	require.Equal(t, original.Adults, decoded.Adults)
	require.Equal(t, original.Children, decoded.Children)
	require.Equal(t, original.BuyTotalPrice, decoded.BuyTotalPrice)
	require.Equal(t, original.BuyPrice, decoded.BuyPrice)
	require.Equal(t, original.ExchangeRateID, decoded.ExchangeRateID)
	require.Equal(t, original.ExchangeRateValue, decoded.ExchangeRateValue)
	require.Equal(t, original.MarkupID, decoded.MarkupID)
	require.Equal(t, original.Markup, decoded.Markup)
	require.Equal(t, original.AdditionalMarkupID, decoded.AdditionalMarkupID)
	require.Equal(t, original.AdditionalMarkup, decoded.AdditionalMarkup)
	require.Equal(t, "ES", decoded.Lang)
	require.Equal(t, original.Id, decoded.Id)
}

func TestInternalBookingCode_SerializeEncrypted_ReturnsBase64Ciphertext(t *testing.T) {
	t.Parallel()

	ibc := &InternalBookingCode{
		RoomCounter: "01",
		CheckInDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		CheckOutDate: time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
	}

	serialized := ibc.Serialize()
	encrypted := ibc.SerializeEncrypted()

	require.NotEmpty(t, encrypted)
	require.NotEqual(t, serialized, encrypted)
	_, err := base64.StdEncoding.DecodeString(encrypted)
	require.NoError(t, err)
}

func TestMapToStringAndStringToMap_RoundTrip(t *testing.T) {
	t.Parallel()

	in := map[string]string{
		"rIdx": "1",
		"foo":  "bar:baz",
	}

	out := StringToMap(MapToString(in))
	require.Equal(t, in, out)
}

func TestTextToStructAndStructToText(t *testing.T) {
	t.Parallel()

	type sample struct {
		Name   string
		Count  int
		Price  float64
		Active bool
	}

	original := sample{Name: "item", Count: 3, Price: 9.99, Active: true}
	text := StructToText(original)
	require.NotEmpty(t, text)

	var decoded sample
	err := TextToStruct(text, &decoded)
	require.NoError(t, err)
	require.Equal(t, original, decoded)
}

func TestTextToStruct_ReturnsErrorForNonPointer(t *testing.T) {
	t.Parallel()

	type sample struct{ Name string }
	var s sample

	err := TextToStruct("Name:test", s)
	require.Error(t, err)
}
