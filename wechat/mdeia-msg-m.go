package wechat

import "fmt"

type mediaMsgMap struct {
	wx *WeChat
}

func newMediaMsgMap(wx *WeChat) *mediaMsgMap {
	return &mediaMsgMap{wx}
}

// MapMsgs ...
func (mm *mediaMsgMap) MapMsgs(msg *CountedContent) {
	for _, m := range msg.Content {
		msgType, _ := m[`MsgType`].(float64)
		mid, _ := m[`MsgId`]
		var path = ``
		switch msgType {
		case 3:
			path = `webwxgetmsgimg`
		case 47:
			pid, _ := m[`HasProductId`].(float64)
			if pid == 0 {
				path = `webwxgetmsgimg`
			}
		case 34:
			path = `webwxgetvoice`
		case 43:
			path = `webwxgetvideo`
		}
		m[`isMediaMsg`] = false
		if len(path) > 0 {
			m[`isMediaMsg`] = true
			m[`MediaMsgDownloadUrl`] = fmt.Sprintf(`%v/%s?msgid=%v&%v`, mm.wx.BaseURL, path, mid, mm.wx.SkeyKV())
			return
		}
	}
}

// HandleMsgs ...
func (mm *mediaMsgMap) HandleMsgs(msg *CountedContent) {}
