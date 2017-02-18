package redpacket

import (
	log "github.com/Sirupsen/logrus"
	"github.com/KevinGong2013/ggbot/modules/bridge"
	"github.com/KevinGong2013/ggbot/wechat"
)

var logger = log.WithFields(log.Fields{
	"module": "redpacket",
})

// Redpacket module
type Redpacket struct {
	bw *bridge.Wrapper
}

// NewRedpacket ...
func NewRedpacket(bw *bridge.Wrapper) *Redpacket {
	return &Redpacket{bw}
}

// MapMsgs implement Module
func (rp *Redpacket) MapMsgs(msg *wechat.CountedContent) {
	for _, m := range msg.Content {
		content, _ := m[`Content`].(string)
		msgType, _ := m[`MsgType`].(float64)
		if msgType == 10000 && (content == `收到红包，请在手机上查看` || content == `Red packet received, view on phone`) {
			m[`MsgType`] = 90000
		}
	}
}

// HandleMsgs implement MsgModule
func (rp *Redpacket) HandleMsgs(msg *wechat.CountedContent) {
	for _, m := range msg.Content {
		msgType, _ := m[`MsgType`].(float64)
		if msgType == 90000 {
			rp.bw.OpenRedPacket()
		}
	}
}
