package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/KevinGong2013/wechat"
)

type assistant struct {
	bot      *wechat.WeChat
	username string
}

func newAssistant(bot *wechat.WeChat, username string) *assistant {
	return &assistant{bot, username}
}

func (a *assistant) delMember(groupUserName, memberUserName string) error {
	ps := map[string]interface{}{
		`DelMemberList`: memberUserName,
		`ChatRoomName`:  groupUserName,
		`BaseRequest`:   a.bot.BaseRequest,
	}
	data, _ := json.Marshal(ps)

	url := fmt.Sprintf(`%s/webwxupdatechatroom?fun=delmember`, a.bot.BaseURL)

	var resp wechat.Response

	err := a.bot.Execute(url, bytes.NewReader(data), &resp)

	if err != nil {
		return err
	}

	if resp.IsSuccess() {
		return nil
	}

	return fmt.Errorf(`delete %s on %s failed`, memberUserName, groupUserName)
}

func (a *assistant) handle(msg wechat.EventMsgData) {
	if msg.IsGroupMsg {
		if msg.MsgType == 10000 && strings.Contains(msg.Content, `加入群聊`) {
			nn, err := search(msg.Content, `"`, `"通过`)
			if err != nil {
				logger.Errorf(`send group welcome failed %s`, msg.Content)
			}
			a.bot.SendTextMsg(`欢迎【`+nn+`】加入群聊`, msg.FromUserName)
		} else if strings.Contains(msg.Content, `签到`) {
			c := a.bot.ContactByUserName(msg.SenderUserName)
			a.bot.SendTextMsg(fmt.Sprintf(`%s 完成了签到`, c.NickName), msg.FromUserName)
		}

		// 群主踢人
		if msg.SenderUserName == a.username && strings.HasPrefix(msg.Content, `滚蛋`) {
			gun := msg.FromUserName
			if msg.IsSendedByMySelf {
				gun = msg.ToUserName
			}
			nn := strings.Replace(msg.Content, `滚蛋`, ``, 1)
			if members, err := a.bot.MembersOfGroup(gun); err == nil {
				for _, c := range members {
					logger.Debug(c.NickName)
					if c.NickName == nn {
						a.bot.SendTextMsg(nn+` 送你免费飞机票`, gun)
						time.Sleep(3 * time.Second)
						err := a.delMember(gun, c.UserName)
						if err != nil {
							a.bot.SendTextMsg(`暂时不T你把`, gun)
						}
					}
				}
			}
		}
	}
}

func search(source, prefix, suffix string) (string, error) {

	index := strings.Index(source, prefix)
	if index == -1 {
		err := fmt.Errorf("can't find [%s] in [%s]", prefix, source)
		return ``, err
	}
	index += len(prefix)

	end := strings.Index(source[index:], suffix)
	if end == -1 {
		err := fmt.Errorf("can't find [%s] in [%s]", suffix, source)
		return ``, err
	}

	result := source[index : index+end]

	return result, nil
}
