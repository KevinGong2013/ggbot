package assistant

import (
	"fmt"
	"strings"
	"time"

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
	group         *wx.Contact
}

// NewAssistant ...
func NewAssistant(groupNickName string, welcome string) *GroupManager {
	return &GroupManager{groupNickName, welcome, nil, nil}
}

// WechatDidLogin ...
func (gm *GroupManager) WechatDidLogin(wechat *wx.WeChat) {
	gm.wx = wechat

	groups, err := gm.wx.ContactByNickName(gm.groupNickName)
	if err != nil || len(groups) != 1 {
		panic(fmt.Sprintf(`找不到指定群，请仔细检查群名称 [%s]`, gm.groupNickName))
	}
	gm.group = groups[0]
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
	ct := contact[`ChangeType`].(int)
	if ct == 1 && gm.wx != nil && gm.group != nil {
		nn, _ := contact[`NickName`].(string)
		gm.wx.SendTextMsg(fmt.Sprintf(`%s 退出了群聊`, nn), gm.group.To())
	}
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
		if m[`IsGroupMsg`].(bool) {
			c := m[`Content`].(string)
			if strings.Contains(c, `签到`) {
				// 写数据库
				var name string
				if m[`IsSendByMySelf`].(bool) {
					name = gm.wx.MySelf.NickName
				} else {
					un := m[`ActualUserName`].(string)
					member, _ := gm.wx.ContactByUserName(un)
					name = member.NickName
				}
				name = utils.ReplaceEmoji(name)
				gm.wx.SendTextMsg(fmt.Sprintf(`%s完成签到 %s`, name, time.Now()), gm.group.To())
				// TODO 写数据库，进行业务逻辑判断 等等
			}
		}
	}
}
