// internal/infrastructure/serializer/go_serializer.go
package serializer

import (
	"encoding/json"
	"encoding/xml"
)

type GoSerializer struct{}

func NewGoSerializer() *GoSerializer {
	return &GoSerializer{}
}

func (s *GoSerializer) ToJSON(v any) ([]byte, error) { // Renamed
	return json.Marshal(v)
}

func (s *GoSerializer) FromJSON(data []byte, v any) error { // Renamed
	return json.Unmarshal(data, v)
}

func (s *GoSerializer) ToXML(v any) ([]byte, error) { // Renamed
	return xml.Marshal(v)
}

func (s *GoSerializer) FromXML(data []byte, v any) error { // Renamed
	return xml.Unmarshal(data, v)
}
