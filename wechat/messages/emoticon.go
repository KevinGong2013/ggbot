package messages

// EmoticonMsg is wechat emoji msg
type EmoticonMsg struct {
	to      string
	mediaID string
}

// Path is text msg's api path
func (msg *EmoticonMsg) Path() string {
	return `webwxsendemoticon?fun=sys`
}

// To destation
func (msg *EmoticonMsg) To() string {
	return msg.to
}

// Content text msg's content
func (msg *EmoticonMsg) Content() map[string]interface{} {
	content := make(map[string]interface{}, 0)

	content[`Type`] = 47
	content[`MediaId`] = msg.mediaID
	content[`EmojiFlag`] = 2

	return content
}

// NewEmoticonMsgMsg create a new instance
func NewEmoticonMsgMsg(mid, to string) *EmoticonMsg {
	return &EmoticonMsg{to, mid}
}

func (msg *EmoticonMsg) String() string {
	return `GIF EMOTICON`
}
