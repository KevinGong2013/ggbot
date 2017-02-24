package guard

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/KevinGong2013/ggbot/utils"
	wx "github.com/KevinGong2013/ggbot/wechat"
	log "github.com/Sirupsen/logrus"
)

var logger = log.WithFields(log.Fields{
	"module": "guard",
})

// Guard ...
type Guard struct {
	autoAddFriend bool
	wx            *wx.WeChat
}

// NewGuard ...
func NewGuard(autoAddFriend bool) *Guard {
	return &Guard{autoAddFriend, nil}
}

// AddFriend ...
func (g *Guard) AddFriend(name, content string) error {
	return g.verifyUser(name, content, 2)
}

// AcceptAddFriend ...
func (g *Guard) AcceptAddFriend(name, content string) error {
	return g.verifyUser(name, content, 3)
}

func (g *Guard) verifyUser(name, content string, status int) error {

	if g.wx == nil {
		return errors.New(`Please Login`)
	}

	url := fmt.Sprintf(`%s/webwxverifyuser?r=%s&%s`, g.wx.BaseURL, utils.Now(), g.wx.PassTicketKV())

	data := map[string]interface{}{
		`BaseRequest`:        g.wx.BaseRequest,
		`Opcode`:             status,
		`VerifyUserListSize`: 1,
		`VerifyUserList`: map[string]string{
			`Value`:            name,
			`VerifyUserTicket`: ``,
		},
		`VerifyContent`:  content,
		`SceneListCount`: 1,
		`SceneList`:      33,
		`skey`:           g.wx.BaseRequest.Skey,
	}

	bs, _ := json.Marshal(data)

	var resp wx.Response

	err := g.wx.Excute(url, bytes.NewReader(bs), &resp)
	if err != nil {
		return err
	}
	if resp.IsSuccess() {
		return nil
	}
	return resp.Error()
}

// WechatDidLogin ...
func (g *Guard) WechatDidLogin(wechat *wx.WeChat) {
	g.wx = wechat
}

// WechatDidLogout ...
func (g *Guard) WechatDidLogout(wechat *wx.WeChat) {
}

// MapMsgs ...
func (g *Guard) MapMsgs(msg *wx.CountedContent) {
	for _, m := range msg.Content {
		m[`isAddFriendMsg`] = false
		m[`needWelecome`] = false
		mt := m[`MsgType`].(float64)
		if mt == 37 && g.autoAddFriend {
			m[`isAddFriendMsg`] = true
			rInfo := m[`RecommendInfo`].(map[string]interface{})
			m[`AddFriendUserName`] = rInfo[`UserName`]
		}
	}
}

// HandleMsgs ...
func (g *Guard) HandleMsgs(msg *wx.CountedContent) {
	for _, m := range msg.Content {
		if m[`isAddFriendMsg`].(bool) {
			err := g.AddFriend(m[`AddFriendUserName`].(string),
				m[`Ticket`].(string))
			if err != nil {
				logger.Error(err)
			}
			err = g.wx.SendTextMsg(`新添加了一个好友`, `filehelper`)
			if err != nil {
				logger.Error(err)
			}
		}
	}
}
