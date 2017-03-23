package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/KevinGong2013/wechat"
)

// Guard ...
type guard struct {
	bot *wechat.WeChat
}

func newGuard(bot *wechat.WeChat) *guard {
	return &guard{bot}
}

// AddFriend ...
func (g *guard) addFriend(username, content string) error {
	return g.verifyUser(username, content, 2)
}

// AcceptAddFriend ...
func (g *guard) acceptAddFriend(username, content string) error {
	return g.verifyUser(username, content, 3)
}

func (g *guard) verifyUser(username, content string, status int) error {

	url := fmt.Sprintf(`%s/webwxverifyuser?r=%s&%s`, g.bot.BaseURL, strconv.FormatInt(time.Now().Unix(), 10), g.bot.PassTicketKV())

	data := map[string]interface{}{
		`BaseRequest`:        g.bot.BaseRequest,
		`Opcode`:             status,
		`VerifyUserListSize`: 1,
		`VerifyUserList`: map[string]string{
			`Value`:            username,
			`VerifyUserTicket`: ``,
		},
		`VerifyContent`:  content,
		`SceneListCount`: 1,
		`SceneList`:      33,
		`skey`:           g.bot.BaseRequest.Skey,
	}

	bs, _ := json.Marshal(data)

	var resp wechat.Response

	err := g.bot.Excute(url, bytes.NewReader(bs), &resp)
	if err != nil {
		return err
	}
	if resp.IsSuccess() {
		return nil
	}
	return resp.Error()
}

func (g *guard) autoAcceptAddFirendRequest(msg wechat.EventMsgData) {
	if msg.MsgType == 37 {
		rInfo := msg.OriginalMsg[`RecommendInfo`].(map[string]interface{})
		err := g.addFriend(rInfo[`UserName`].(string),
			msg.OriginalMsg[`Ticket`].(string))
		if err != nil {
			logger.Error(err)
		}
		err = g.bot.SendTextMsg(`新添加了一个好友`, `filehelper`)
		if err != nil {
			logger.Error(err)
		}
	}
}
