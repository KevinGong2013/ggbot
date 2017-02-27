package assistant

import (
	"strings"

	"github.com/KevinGong2013/ggbot/utils"
	wx "github.com/KevinGong2013/ggbot/wechat"
	log "github.com/Sirupsen/logrus"
)

var logger = log.WithFields(log.Fields{
	"module": "assistant",
})

// GroupManager ...
type GroupManager struct {
	groupNickName string
	welcome       string
	wx            *wx.WeChat
}

// NewAssistant ...
func NewAssistant(groupNickName string, welcome string) *GroupManager {
	return &GroupManager{groupNickName, welcome, nil}
}

// WechatDidLogin ...
func (gm *GroupManager) WechatDidLogin(wechat *wx.WeChat) {
	gm.wx = wechat

	group, err := gm.wx.ContactByNickName(gm.groupNickName)

	logger.Debug(gm.groupNickName)

	if err != nil {
		logger.Error(err)
	} else {
		for _, c := range group.MemberList {
			logger.Debug(c.NickName)
		}
	}
}

// WechatDidLogout ...
func (gm *GroupManager) WechatDidLogout(wechat *wx.WeChat) {
}

// MapContact ...
func (gm *GroupManager) MapContact(contact map[string]interface{}) {
	return
}

// HandleContact ...
func (gm *GroupManager) HandleContact(contact map[string]interface{}) {
	logger.Info(contact)
}

// MapMsgs ...
func (gm *GroupManager) MapMsgs(msg *wx.CountedContent) {
	for _, m := range msg.Content {
		m[`needWelecome`] = false
		if m[`IsGroupMsg`].(bool) { // 判断是不是入群提醒
			mt := m[`MsgType`].(float64)
			content := m[`Content`].(string)
			if strings.Contains(content, `加入群聊`) && mt == 10000 {
				// 查找nickname
				nn, err := utils.Search(content, `"`, `"通过`)
				if err == nil {
					m[`needWelecome`] = true
					m[`needWelecomeNickName`] = nn
				}
			} else if strings.Contains(content, `签到`) {

			}
		}
	}
}

// HandleMsgs ...
func (gm *GroupManager) HandleMsgs(msg *wx.CountedContent) {
	for _, m := range msg.Content {
		if m[`needWelecome`].(bool) {
			nn := m[`needWelecomeNickName`].(string)
			err := gm.wx.SendTextMsg(`欢迎【`+nn+`】加入群聊`, m[`FromUserName`].(string))
			if err != nil {
				logger.Error(err)
			}
		}
	}
}
