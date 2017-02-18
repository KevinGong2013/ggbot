package echo

import (
	log "github.com/Sirupsen/logrus"
	wx "github.com/KevinGong2013/ggbot/wechat"
)

var logger = log.WithFields(log.Fields{
	"module": "echo  ",
})

// Echo module
type Echo struct {
}

//WechatDidLogin ...
func (e *Echo) WechatDidLogin(wechat *wx.WeChat) {
	logger.Info(`wechat did login`)
}

//WechatDidLogout ...
func (e *Echo) WechatDidLogout(wechat *wx.WeChat) {
	logger.Info(`wechat did logout`)
}

// MapMsgs implement Module
func (e *Echo) MapMsgs(msg *wx.CountedContent) {}

// HandleMsgs implement Module
func (e *Echo) HandleMsgs(msg *wx.CountedContent) {
	logger.Debugf(`did receive %v message(s)`, msg.Count)
	for _, m := range msg.Content {
		logger.Infof(`%v`, m)
	}
}

// MapContact implement Module
func (e *Echo) MapContact(contact map[string]interface{}) {}

// HandleContact implement Module
func (e *Echo) HandleContact(contact map[string]interface{}) {
	logger.Infof(`contact change: %v`, contact)
}
