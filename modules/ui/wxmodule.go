package ui

import (
	"fmt"

	wx "github.com/KevinGong2013/ggbot/wechat"
)

// WechatDidLogin ...
func (ui *UserInterface) WechatDidLogin(wechat *wx.WeChat) {
	ui.wx = wechat
}

// WechatDidLogout ...
func (ui *UserInterface) WechatDidLogout(wechat *wx.WeChat) {
}

// MapMsgs ...
func (ui *UserInterface) MapMsgs(msg *wx.CountedContent) {

	for _, m := range msg.Content {
		msgType, _ := m[`MsgType`].(float64)
		if msgType == 51 {
			m[`Content`] = `别乱动手机`
		}
	}
}

// HandleMsgs ...
func (ui *UserInterface) HandleMsgs(msg *wx.CountedContent) {
	for _, m := range msg.Content {
		name := `unkow`
		from, _ := m[`FromUserName`].(string)
		contact, err := ui.wx.ContactByUserName(from)
		if err == nil {
			name = contact.NickName
		}
		content, _ := m[`Content`].(string)

		if liveMsgList != nil {
			item := fmt.Sprintf(`[%v](fg-green) %v`, name, content)
			liveMsgList.AppendAtLast(item)
		}
	}
}
