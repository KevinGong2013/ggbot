package xiaoice

import (
	"sync"

	"github.com/KevinGong2013/ggbot/utils"
	wx "github.com/KevinGong2013/ggbot/wechat"
)

import log "github.com/Sirupsen/logrus"

var logger = log.WithFields(log.Fields{
	"module": "xiaoice",
})

// Brain ...
type Brain struct {
	sync.Mutex
	wx             *wx.WeChat
	xiaoice        *wx.Contact
	waittingReplay []string
}

// NewBrain ...
func NewBrain() *Brain {
	return &Brain{waittingReplay: []string{}}
}

// WechatDidLogin ...
func (b *Brain) WechatDidLogin(wechat *wx.WeChat) {
	b.wx = wechat
	b.xiaoice, _ = wechat.ContactByNickName(`小冰`)
}

// WechatDidLogout ...
func (b *Brain) WechatDidLogout(wechat *wx.WeChat) {
}

// MapMsgs ...
func (b *Brain) MapMsgs(msg *wx.CountedContent) {
	for _, m := range msg.Content {
		isSendByMySelf, _ := m[`IsSendByMySelf`].(bool)
		if isSendByMySelf {
			continue
		}
		from, _ := m[`FromUserName`].(string)
		contact, err := b.wx.ContactByUserName(from)
		if err != nil {
			m[`needXiaoiceResponse`] = false
			logger.Error(err)
			continue
		}

		switch contact.Type {
		case wx.ContactTypeFriend:
			m[`needXiaoiceResponse`] = true
			m[`xiaoice_info`] = m[`Content`]
			m[`xiaoice_to`] = m[`FromUserName`]
		case wx.ContactTypeOfficial:
			if b.xiaoice.NickName == contact.NickName {
				len := len(b.waittingReplay)
				if len > 0 {
					b.Lock()
					m[`isXiaoiceReplay`] = true
					m[`ReplayUserName`] = b.waittingReplay[len-1]
					m[`localFileId`] = m[`MsgId`]
					b.waittingReplay = b.waittingReplay[:len-1]
					b.Unlock()
				} else {
					logger.Warnf(`xiaoice replay %s`, m)
				}
			} else {
				logger.Warn(`offical msg %s`, contact.NickName)
			}
			m[`needXiaoiceResponse`] = false
		case wx.ContactTypeGroup:
			m[`needXiaoiceResponse`] = true
			m[`xiaoice_info`] = m[`Content`]
			m[`xiaoice_to`] = m[`FromUserName`]
		}
	}
}

// HandleMsgs ...
func (b *Brain) HandleMsgs(msg *wx.CountedContent) {
	for _, m := range msg.Content {
		needResponse, _ := m[`needXiaoiceResponse`].(bool)
		isReplay, _ := m[`isXiaoiceReplay`].(bool)
		if needResponse {
			c, _ := m[`xiaoice_info`].(string)
			to, _ := m[`xiaoice_to`].(string)

			if b.xiaoice != nil {
				var err error
				if ok, isM := m[`isMediaMsg`].(bool); ok && isM {
					path, e := b.wx.DownloadMedia(m[`MediaMsgDownloadUrl`].(string), m[`MsgId`].(string))
					defer utils.DeleteFile(path)
					if e == nil {
						err = b.wx.SendFile(path, b.xiaoice.To())
					} else {
						err = e
					}
				} else {
					err = b.wx.SendTextMsg(c, b.xiaoice.To())
				}
				if err == nil {
					b.Lock()
					b.waittingReplay = append(b.waittingReplay, to)
					b.Unlock()
				} else {
					logger.Error(err)
				}
			}
		}
		if isReplay {
			to, _ := m[`ReplayUserName`].(string)
			c, _ := m[`Content`].(string)
			msgType, _ := m[`MsgType`].(float64)

			if msgType == 1 {
				b.wx.SendTextMsg(c, to)
			} else if m[`isMediaMsg`].(bool) {
				path, err := b.wx.DownloadMedia(m[`MediaMsgDownloadUrl`].(string), m[`MsgId`].(string))
				defer utils.DeleteFile(path)
				if err != nil {
					logger.Error(err)
				}
				err = b.wx.SendFile(path, to)
				if err != nil {
					logger.Error(err)
				}
			}
		}
	}
}
