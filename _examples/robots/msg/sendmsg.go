package main

import (
	"github.com/KevinGong2013/ggbot/modules/gguuid"
	"github.com/KevinGong2013/ggbot/wechat"
)

func main() {

	bot, err := wechat.WakeUp(gguuid.New())
	if err != nil {
		panic(err)
	}

	bot.RegisterModule(new(tester))

	select {}
}

type tester struct{}

func (t *tester) WechatDidLogin(wechat *wechat.WeChat) {
	to := `filehelper`
	wechat.SendTextMsg(`~~~~~~~~~~~~~~~~~~`, to)
	wechat.SendFile(`testResource/test.mov`, to)
	wechat.SendFile(`testResource/test.png`, to)
	wechat.SendFile(`testResource/test.gif`, to)
	wechat.SendFile(`testResource/test.txt`, to)
	wechat.SendFile(`testResource/test.mp3`, to)
	wechat.SendTextMsg(`--------------------------`, to)
}

func (t *tester) WechatDidLogout(wechat *wechat.WeChat) {}
