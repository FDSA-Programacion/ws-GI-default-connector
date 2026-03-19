package serializer

type Serializer interface {
	ToJSON(v any) ([]byte, error)
	FromJSON(data []byte, v any) error

	ToXML(v any) ([]byte, error)
	FromXML(data []byte, v any) error
}
