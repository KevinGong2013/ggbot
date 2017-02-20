package storage

import (
	"time"

	"github.com/KevinGong2013/ggbot/utils"
	wx "github.com/KevinGong2013/ggbot/wechat"
)

// WechatDidLogin ...
func (st *Storage) WechatDidLogin(wechat *wx.WeChat) {
	st.db.Write(`login`, `WechatDidLogin`, utils.Now())
}

// WechatDidLogout ...
func (st *Storage) WechatDidLogout(wechat *wx.WeChat) {
	st.db.Write(`login`, `WechatDidLogout`, utils.Now())
}

// MapContact ...
func (st *Storage) MapContact(contact *wx.Contact) {
	return
}

// HandleContact ...
func (st *Storage) HandleContact(contact *wx.Contact) {
	st.db.Write(`contact`, contact.NickName, contact)
}

// MapMsgs ...
func (st *Storage) MapMsgs(msg *wx.CountedContent) {
	return
}

// HandleMsgs ...
func (st *Storage) HandleMsgs(msg *wx.CountedContent) {
	for _, m := range msg.Content {
		st.db.Write(`msg`, time.Now().String(), m)
	}
}
