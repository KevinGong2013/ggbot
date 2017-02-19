package convenience

import (
	"time"

	wx "github.com/KevinGong2013/ggbot/wechat"
)

// MsgEventData ...
type MsgEventData struct {
	FromUserName string
	FromNickName string
	Msg          map[string]interface{}
}

// WechatDidLogin ...
func (ms *MsgStream) WechatDidLogin(wechat *wx.WeChat) {
	ms.wx = wechat
}

// WechatDidLogout ...
func (ms *MsgStream) WechatDidLogout(wechat *wx.WeChat) {
}

// MapMsgs ...
func (ms *MsgStream) MapMsgs(msg *wx.CountedContent) {}

// HandleMsgs ...
func (ms *MsgStream) HandleMsgs(msg *wx.CountedContent) {
	for _, m := range msg.Content {
		isSendByMySelf, _ := m[`IsSendByMySelf`].(bool)
		if isSendByMySelf {
			continue
		}
		isGroupMsg, _ := m[`IsGroupMsg`].(bool)
		ms.deliverMsg(m, isGroupMsg)
	}
}

func (ms *MsgStream) deliverMsg(msg map[string]interface{}, isGroupMsg bool) {

	var username string
	var path string

	if isGroupMsg {
		username, _ = msg[`ActualUserName`].(string)
		path = `/msg/group/`
	} else {
		username, _ = msg[`FromUserName`].(string)
		path = `/msg/solo/`
	}

	contact, err := ms.wx.ContactByUserName(username)
	if err != nil {
		logger.Error(err)
		return
	}

	d := MsgEventData{
		FromUserName: contact.UserName,
		FromNickName: contact.NickName,
		Msg:          msg,
	}

	e := Event{
		Path: path + contact.NickName,
		Time: time.Now().Unix(),
		Data: d,
	}
	ms.msgEvent <- e
}
