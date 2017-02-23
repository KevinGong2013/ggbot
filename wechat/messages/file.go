package messages

import "fmt"

// FileMsg struct
type FileMsg struct {
	to      string
	mediaID string
	path    string
	ftype   int
	fname   string
	ext     string
}

// Path is text msg's api path
func (msg *FileMsg) Path() string {
	return msg.path
}

// To destation
func (msg *FileMsg) To() string {
	return msg.to
}

// Content text msg's content
func (msg *FileMsg) Content() map[string]interface{} {
	content := make(map[string]interface{}, 0)

	content[`Type`] = msg.ftype

	if msg.ftype == 6 {
		content[`Content`] = fmt.Sprintf(`<appmsg appid='wxeb7ec651dd0aefa9' sdkver=''><title>%s</title><des></des><action></action><type>6</type><content></content><url></url><lowurl></lowurl><appattach><totallen>10</totallen><attachid>%s</attachid><fileext>%s</fileext></appattach><extinfo></extinfo></appmsg>`, msg.fname, msg.mediaID, msg.ext)
	} else {
		content[`MediaId`] = msg.mediaID
	}

	return content
}

// NewFileMsg construct a new FileMsg's instance
func NewFileMsg(mediaID, to, name, ext string) *FileMsg {
	return &FileMsg{to, mediaID, `webwxsendappmsg?fun=async&f=json`, 6, name, ext}
}

// NewImageMsg ..
func NewImageMsg(mediaID, to string) *FileMsg {
	return &FileMsg{to, mediaID, `webwxsendmsgimg?fun=async&f=json`, 3, ``, ``}
}

// NewVideoMsg ..
func NewVideoMsg(mediaID, to string) *FileMsg {
	return &FileMsg{to, mediaID, `webwxsendvideomsg?fun=async&f=json`, 43, ``, ``}
}

func (msg *FileMsg) String() string {
	if msg.ftype == 3 {
		return `IMAGE`
	} else if msg.ftype == 4 {
		return `GIF EMOTICON`
	} else {
		return `FILE`
	}
}
