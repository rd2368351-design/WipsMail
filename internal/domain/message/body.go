package message

type Body struct {
	PlainText string
	HTML      string
	Raw       string
}

func NewBody(plainText, html string) Body {
	return Body{
		PlainText: plainText,
		HTML:      html,
	}
}

func (b Body) IsEmpty() bool {
	return b.PlainText == "" && b.HTML == ""
}