package convenience

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	wx "github.com/KevinGong2013/ggbot/wechat"
)

var logger = log.WithFields(log.Fields{
	"module": "convenience",
})

// MsgStream ...
type MsgStream struct {
	sync.RWMutex
	Handlers map[string]func(msg map[string]interface{})
	wx       *wx.WeChat
}

//DefaultMsgStream ...
var DefaultMsgStream = NewMsgStream()

// NewMsgStream ...
func NewMsgStream() *MsgStream {
	return &MsgStream{
		Handlers: make(map[string]func(msg map[string]interface{})),
	}
}

// Handle path use default ms
func Handle(path string, handler func(msg map[string]interface{})) {
	DefaultMsgStream.Handle(path, handler)
}

// WechatDidLogin ...
func (ms *MsgStream) WechatDidLogin(wechat *wx.WeChat) {
	ms.wx = wechat
}

// WechatDidLogout ...
func (ms *MsgStream) WechatDidLogout(wechat *wx.WeChat) {
}

// Handle ...
// /msg/solo
// /msg/solo/GGBot
// /msg/group
// /msg/group/GGBot测试群
func (ms *MsgStream) Handle(path string, handler func(msg map[string]interface{})) {
	ms.Handlers[cleanPath(path)] = handler
}

// ResetHandlers can Remove all existing defined Handlers from the map
func (ms *MsgStream) ResetHandlers() {
	for path := range ms.Handlers {
		delete(ms.Handlers, path)
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

	ms.RLock()
	defer ms.RUnlock()

	if pattern := ms.match(path + contact.NickName); pattern != `` {
		ms.Handlers[pattern](msg)
	}
}

func (ms *MsgStream) match(path string) string {
	return findMatch(ms.Handlers, path)
}

func findMatch(mux map[string]func(msg map[string]interface{}), path string) string {
	n := -1
	pattern := ""
	for m := range mux {
		if !isPathMatch(m, path) {
			continue
		}
		if len(m) > n {
			pattern = m
			n = len(m)
		}
	}
	return pattern

}

func isPathMatch(pattern, path string) bool {
	if len(pattern) == 0 {
		return false
	}
	n := len(pattern)
	return len(path) >= n && path[0:n] == pattern
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return p
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
