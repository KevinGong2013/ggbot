package messages

// TextMsg struct
type TextMsg struct {
	to      string
	content string
}

// Path is text msg's api path
func (msg *TextMsg) Path() string {
	return `webwxsendmsg`
}

// To destation
func (msg *TextMsg) To() string {
	return msg.to
}

// Content text msg's content
func (msg *TextMsg) Content() map[string]interface{} {
	content := make(map[string]interface{}, 0)

	content["Type"] = 1
	content["Content"] = msg.content

	return content
}

// NewTextMsg construct a new TextMsg's instance
func NewTextMsg(text, to string) *TextMsg {
	return &TextMsg{to, text}
}

func (msg *TextMsg) String() string {
	return msg.content
}
