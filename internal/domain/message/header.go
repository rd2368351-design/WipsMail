package message

type Header struct {
	Name  string
	Value string
}

func NewHeader(name, value string) Header {
	return Header{Name: name, Value: value}
}