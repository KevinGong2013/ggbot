package main

import (
	"fmt"
	"strings"

	"github.com/KevinGong2013/wechat"
)

type assisant struct {
	bot *wechat.WeChat
}

func newAssisant(bot *wechat.WeChat) *assisant {
	return &assisant{bot}
}

func (a *assisant) handle(msg wechat.EventMsgData) {
	if msg.IsGroupMsg {
		if msg.MsgType == 10000 && strings.Contains(msg.Content, `加入群聊`) {
			nn, err := search(msg.Content, `"`, `"通过`)
			if err != nil {
				logger.Errorf(`send group welcome failed %s`, msg.Content)
			}
			a.bot.SendTextMsg(`欢迎【`+nn+`】加入群聊`, msg.FromUserName)
		} else if strings.Contains(msg.Content, `签到`) {
			if c, err := a.bot.ContactByUserName(msg.SenderUserName); err == nil {
				a.bot.SendTextMsg(fmt.Sprintf(`%s 完成了签到`, c.NickName), msg.FromUserName)
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
