package bookingcode

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	PIPE                       = "|"
	HYPHEN                     = "-"
	NEGATION                   = "¬"
	ESC_PIPE                   = "\\|"
	PIPE_STR                   = "|"
	AT_DOLAR                   = "@$"
	DATE_SHORT_FORMAT          = "20060102"
	DATE_SCRIPT_FORMAT         = "2006-01-02"
	COMMA                      = ","
	EXTRAPARAMS_SEPARATOR      = ":_#_:"
	EXTRAPARAM_VALUE_SEPARATOR = "<:=:>"
	PROVIDER_ROOM_INDEX        = "rIdx"
)

type InternalBookingCode struct {
	BookingCode          string
	EncryptedBookingCode string
	RoomCounter          string
	EchoToken            string
	Version              string
	ProviderCode         string
	GiHotelCode          string
	PrvHotelCode         string
	CheckInDate          time.Time
	CheckOutDate         time.Time
	GIRoomID             string
	GIRoomCode           string
	GIRoomName           string
	PrvRoomCode          string
	PrvRoomName          string
	IsDynamicRoom        int
	NumberOfRooms        int
	Adults               int
	Children             int
	ChildrenAges         string
	Infant               int
	CustomerCountry      string
	Market               string
	BoardId              string
	BuyTotalPrice        float32
	BuyPrice             float32
	RetailAmount         string
	PreBookCode          string
	Lang                 string
	Id                   string
	Distribution         string
	AdultAges            string
	NumberOfService      int
	RateDistribution     string
	RateChildAges        string
	RateCode             string
	IsSpecificRate       int
	ExtraParams          string
	BuyCurrency          string
	SaleCurrency         string
	ExchangeRateID       int
	ExchangeRateValue    float32
	Markup               float32
	MarkupID             int
	AdditionalMarkup     float32
	AdditionalMarkupID   int
	NrRate               int
	InvBlockCode         string
}

func (ibc *InternalBookingCode) Serialize() string {
	b := bytes.Buffer{}
	b.WriteString(ibc.RoomCounter)
	b.WriteString(replacePipe(ibc.EchoToken))
	b.WriteString(replacePipe(ibc.Version))
	b.WriteString(replacePipe(ibc.ProviderCode))
	b.WriteString(replacePipe(ibc.GiHotelCode))
	b.WriteString(replacePipe(ibc.PrvHotelCode))
	b.WriteString(ibc.CheckInDate.Format(DATE_SHORT_FORMAT))
	b.WriteString(PIPE)
	b.WriteString(ibc.CheckOutDate.Format(DATE_SHORT_FORMAT))
	b.WriteString(PIPE)
	b.WriteString(replacePipe(ibc.GIRoomID))
	b.WriteString(replacePipe(ibc.GIRoomCode))
	b.WriteString(replacePipe(ibc.GIRoomName))
	b.WriteString(replacePipe(ibc.PrvRoomCode))
	b.WriteString(replacePipe(ibc.PrvRoomName))
	b.WriteString(strconv.Itoa(ibc.IsDynamicRoom))
	b.WriteString(PIPE)
	b.WriteString(strconv.Itoa(ibc.NumberOfService))
	b.WriteString(PIPE)
	b.WriteString(strconv.Itoa(ibc.NumberOfRooms))
	b.WriteString(PIPE)
	b.WriteString(replacePipe(ibc.Distribution))
	b.WriteString(strconv.Itoa(ibc.Adults))
	b.WriteString(PIPE)
	b.WriteString(strconv.Itoa(ibc.Children))
	b.WriteString(PIPE)
	b.WriteString(replacePipe(ibc.AdultAges))
	b.WriteString(replacePipe(ibc.ChildrenAges))
	b.WriteString(strconv.Itoa(ibc.Infant))
	b.WriteString(PIPE)
	b.WriteString(replacePipe(ibc.CustomerCountry))
	b.WriteString(replacePipe(ibc.Market))
	b.WriteString(replacePipe(ibc.BoardId))
	b.WriteString(strconv.FormatFloat(float64(ibc.BuyTotalPrice), 'f', 2, 64))
	b.WriteString(PIPE)
	b.WriteString(strconv.FormatFloat(float64(ibc.BuyPrice), 'f', 2, 64))
	b.WriteString(PIPE)
	b.WriteString(replacePipe(ibc.BuyCurrency))
	b.WriteString(replacePipe(ibc.SaleCurrency))
	b.WriteString(strconv.Itoa(ibc.ExchangeRateID))
	b.WriteString(PIPE)
	b.WriteString(strconv.FormatFloat(float64(ibc.ExchangeRateValue), 'f', 2, 64))
	b.WriteString(PIPE)
	b.WriteString(strconv.Itoa(ibc.MarkupID))
	b.WriteString(PIPE)
	b.WriteString(strconv.FormatFloat(float64(ibc.Markup), 'f', 2, 64))
	b.WriteString(PIPE)
	b.WriteString(strconv.Itoa(ibc.AdditionalMarkupID))
	b.WriteString(PIPE)
	b.WriteString(strconv.FormatFloat(float64(ibc.AdditionalMarkup), 'f', 2, 64))
	b.WriteString(PIPE)
	b.WriteString(replacePipe(ibc.RetailAmount))
	b.WriteString(replacePipe(ibc.RateDistribution))
	b.WriteString(replacePipe(ibc.RateChildAges))
	b.WriteString(replacePipe(ibc.RateCode))
	b.WriteString(strconv.Itoa(ibc.IsSpecificRate))
	b.WriteString(PIPE)
	b.WriteString(strconv.Itoa(ibc.NrRate))
	b.WriteString(PIPE)
	b.WriteString(replacePipe(ibc.InvBlockCode))
	b.WriteString(replacePipe(ibc.ExtraParams))
	b.WriteString(replacePipe(ibc.PreBookCode))
	if ibc.Lang != "" {
		b.WriteString(replacePipe(strings.ToUpper(ibc.Lang)))
	}
	b.WriteString(strings.Replace(ibc.Id, PIPE, AT_DOLAR, -1))

	return b.String()
}

func (ibc *InternalBookingCode) SerializeEncrypted() string {
	serialized := ibc.Serialize()
	encrypted, err := encryptBookingCode(serialized)
	if err != nil {
		return serialized
	}
	return encrypted
}

func strToFloat32(input string) float32 {
	val, _ := strconv.ParseFloat(replaceAtDollar(input), 64)
	return float32(val)
}

func (ibc *InternalBookingCode) Deserialize(bookingCode string) {
	parts := strings.Split(bookingCode, "|")
	ibc.RoomCounter = replaceAtDollar(parts[0][:2])
	ibc.EchoToken = replaceAtDollar(parts[0][2:])
	ibc.Version = replaceAtDollar(parts[1])
	ibc.ProviderCode = replaceAtDollar(parts[2])
	ibc.GiHotelCode = replaceAtDollar(parts[3])
	ibc.PrvHotelCode = replaceAtDollar(parts[4])
	ibc.CheckInDate, _ = time.Parse(DATE_SHORT_FORMAT, replaceAtDollar(parts[5]))
	ibc.CheckOutDate, _ = time.Parse(DATE_SHORT_FORMAT, replaceAtDollar(parts[6]))
	ibc.GIRoomID = replaceAtDollar(parts[7])
	ibc.GIRoomCode = replaceAtDollar(parts[8])
	ibc.GIRoomName = replaceAtDollar(parts[9])
	ibc.PrvRoomCode = replaceAtDollar(parts[10])
	ibc.PrvRoomName = replaceAtDollar(parts[11])
	ibc.IsDynamicRoom, _ = strconv.Atoi(replaceAtDollar(parts[12]))
	ibc.NumberOfService, _ = strconv.Atoi(replaceAtDollar(parts[13]))
	ibc.NumberOfRooms, _ = strconv.Atoi(replaceAtDollar(parts[14]))
	ibc.Distribution = replaceAtDollar(parts[15])
	ibc.Adults, _ = strconv.Atoi(replaceAtDollar(parts[16]))
	ibc.Children, _ = strconv.Atoi(replaceAtDollar(parts[17]))
	ibc.AdultAges = replaceAtDollar(parts[18])
	ibc.ChildrenAges = replaceAtDollar(parts[19])
	ibc.Infant, _ = strconv.Atoi(replaceAtDollar(parts[20]))
	ibc.CustomerCountry = replaceAtDollar(parts[21])
	ibc.Market = replaceAtDollar(parts[22])
	ibc.BoardId = replaceAtDollar(parts[23])
	ibc.BuyTotalPrice = strToFloat32(replaceAtDollar(parts[24]))
	ibc.BuyPrice = strToFloat32(replaceAtDollar(parts[25]))
	ibc.BuyCurrency = replaceAtDollar(parts[26])
	ibc.SaleCurrency = replaceAtDollar(parts[27])
	ibc.ExchangeRateID, _ = strconv.Atoi(replaceAtDollar(parts[28]))
	ibc.ExchangeRateValue = strToFloat32(replaceAtDollar(parts[29]))
	ibc.MarkupID, _ = strconv.Atoi(replaceAtDollar(parts[30]))
	ibc.Markup = strToFloat32(replaceAtDollar(parts[31]))
	ibc.AdditionalMarkupID, _ = strconv.Atoi(replaceAtDollar(parts[32]))
	ibc.AdditionalMarkup = strToFloat32(replaceAtDollar(parts[33]))
	ibc.RetailAmount = replaceAtDollar(parts[34])
	ibc.RateDistribution = replaceAtDollar(parts[35])
	ibc.RateChildAges = replaceAtDollar(parts[36])
	ibc.RateCode = replaceAtDollar(parts[37])
	ibc.IsSpecificRate, _ = strconv.Atoi(replaceAtDollar(parts[38]))
	ibc.NrRate, _ = strconv.Atoi(replaceAtDollar(parts[39]))
	ibc.InvBlockCode = replaceAtDollar(parts[40])
	ibc.ExtraParams = replaceAtDollar(parts[41])
	ibc.PreBookCode = replaceAtDollar(parts[42])
	if len(parts) == 45 {
		ibc.Lang = strings.ToUpper(replaceAtDollar(parts[43]))
		ibc.Id = replaceAtDollar(parts[44])
		return
	}
	ibc.Lang = ""
	ibc.Id = replaceAtDollar(parts[43])
}

func MapToString(v map[string]string) string {
	parts := []string{}
	for key, value := range v {
		parts = append(parts, fmt.Sprintf("%s:%s", key, value))
	}
	return strings.Join(parts, "@@@")
}

func StringToMap(s string) map[string]string {
	pairs := strings.Split(s, "@@@")
	m := make(map[string]string)

	for _, pair := range pairs {
		kv := strings.SplitN(pair, ":", 2)

		if len(kv) != 2 {
			continue
		}

		m[kv[0]] = kv[1]
	}
	return m
}

// DecodeQuoteFromExtraParams decodifica el valor "quote" de ExtraParams (guardado en base64 en Avail/PreBook).
// Si no es base64 válido, devuelve el valor tal cual (compatibilidad con datos antiguos en JSON plano).
func DecodeQuoteFromExtraParams(encoded string) string {
	if encoded == "" {
		return ""
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return encoded
	}
	return string(decoded)
}

func StructToText(v interface{}) string {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return ""
	}

	var parts []string

	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
		fieldValue := fmt.Sprintf("%v", val.Field(i).Interface())
		parts = append(parts, fmt.Sprintf("%s:%s", fieldName, fieldValue))
	}

	return strings.Join(parts, "@@@")
}

func TextToStruct(s string, out interface{}) error {
	val := reflect.ValueOf(out)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("debes pasar un puntero a struct")
	}
	val = val.Elem()

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("el valor debe ser un struct")
	}

	pairs := strings.Split(s, "@@@")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) != 2 {
			continue
		}
		key := kv[0]
		value := kv[1]

		field := val.FieldByName(key)
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(value)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Type() == reflect.TypeOf(time.Time{}) {
				t, err := time.Parse("2006-01-02", value)
				if err != nil {
					return err
				}
				field.Set(reflect.ValueOf(t))
			} else {
				i, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return err
				}
				field.SetInt(i)
			}

		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			field.SetFloat(f)

		case reflect.Bool:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			field.SetBool(b)
		}
	}

	return nil
}

func replaceAtDollar(str string) string {
	return strings.Replace(str, AT_DOLAR, "|", -1)
}

func replacePipe(str string) string {
	b := bytes.Buffer{}
	b.WriteString(strings.Replace(str, "|", AT_DOLAR, -1))
	b.WriteString(PIPE)
	return b.String()
}

var (
	bookingCodeSalt        = []byte{0xA9, 0x9B, 0xC8, 0x32, 0x56, 0x35, 0xE3, 0x03}
	obtentionIterations    = 19
	bookingCodeSecretValue = "GuEsT.2017.InCoMiNg"
)

func encryptBookingCode(plainText string) (string, error) {
	padNum := byte(8 - len(plainText)%8)
	for i := byte(0); i < padNum; i++ {
		plainText += string(padNum)
	}

	dk, iv := getDerivedKey(bookingCodeSecretValue, bookingCodeSalt, obtentionIterations)

	block, err := des.NewCipher(dk)
	if err != nil {
		return "", err
	}

	encrypter := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(plainText))
	encrypter.CryptBlocks(encrypted, []byte(plainText))

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func getDerivedKey(password string, salt []byte, count int) ([]byte, []byte) {
	key := md5.Sum([]byte(password + string(salt)))
	for i := 0; i < count-1; i++ {
		key = md5.Sum(key[:])
	}
	return key[:8], key[8:]
}

type IBCFile struct {
	NumberOfService int
	NumberOfRooms   int
	Distribution    string
	Adults          int
	Children        int
	AdultAges       string
	ChildrenAges    string
}
