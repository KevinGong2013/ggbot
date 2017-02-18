package media

import (
	"fmt"

	wx "github.com/KevinGong2013/ggbot/wechat"
)

// WechatDidLogin ...
func (d *Downloader) WechatDidLogin(wechat *wx.WeChat) {
	d.wx = wechat
}

// WechatDidLogout ...
func (d *Downloader) WechatDidLogout(wechat *wx.WeChat) {
}

// MapMsgs ...
func (d *Downloader) MapMsgs(msg *wx.CountedContent) {
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
		if len(path) > 0 {
			m[`needDownload`] = true
			m[`url`] = fmt.Sprintf(`%v/%s?msgid=%v&%v`, d.wx.BaseURL, path, mid, d.wx.SkeyKV())
			return
		}
	}
}

// HandleMsgs ...
func (d *Downloader) HandleMsgs(msg *wx.CountedContent) {
	for _, m := range msg.Content {
		if ok, n := m[`needDownload`].(bool); ok && n {
			url, _ := m[`url`].(string)
			name, _ := m[`MsgId`].(string)
			go d.download(url, name)
		}
	}
}
