package messages

import (
	"fmt"
	"os"
)

// FileMsg struct
type FileMsg struct {
	to       string
	mediaID  string
	path     string
	filetype int
	info     os.FileInfo
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

	content[`Type`] = msg.filetype

	if msg.filetype == 6 {
		content[`Content`] = fmt.Sprintf(`<appmsg appid='wxeb7ec651dd0aefa9' sdkver=''><title>%s</title><des></des><action></action><type>6</type><content></content><url></url><lowurl></lowurl><appattach><totallen>10</totallen><attachid>%s</attachid><fileext>txt</fileext></appattach><extinfo></extinfo></appmsg>`, msg.info.Name(), msg.mediaID)
	} else {
		content[`MediaId`] = msg.mediaID
	}

	return content
}

// NewFileMsg construct a new FileMsg's instance
func NewFileMsg(path, mediaID, to string, filetype int, info os.FileInfo) *FileMsg {
	return &FileMsg{to, mediaID, path, filetype, info}
}

func (msg *FileMsg) String() string {
	if msg.filetype == 3 {
		return `IMAGE`
	} else if msg.filetype == 4 {
		return `GIF EMOTICON`
	} else {
		return `FILE`
	}
}
