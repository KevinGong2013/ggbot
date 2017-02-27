package main

import (
	"fmt"

	"github.com/KevinGong2013/ggbot/wechat"
	log "github.com/Sirupsen/logrus"
)

func main() {

	log.SetLevel(log.DebugLevel)

	bot, err := wechat.WakeUp(nil)
	if err != nil {
		panic(err)
	}

	bot.RegisterModule(new(tester))

	select {}
}

type tester struct {
}

func (t *tester) WechatDidLogin(wx *wechat.WeChat) {

	// 所有的联系人
	all := wx.AllContacts()

	fmt.Println(`所有联系人`)
	for _, c := range all {
		fmt.Printf(`GGID: %s NickName: %s`, c.GGID, c.NickName)
		fmt.Println()
	}

	// 所有群组
	fmt.Println(`所有群组`)
	for _, c := range wx.AllContacts() {
		if c.Type == wechat.Group {
			fmt.Printf(`GGID: %s, Name: %s`, c.GGID, c.NickName)
			fmt.Println()
		}
	}

	// 所有叫GGBot 的人
	fmt.Println(`查找GGBOT`)
	cs, _ := wx.ContactByNickName(`GGBOT`)
	for _, c := range cs {
		fmt.Printf(`GGID: %s, Name: %s`, c.GGID, c.NickName)
		fmt.Println()
	}

	// 通过ggid找人 2a711e80-46c2-4209-8433-aa02a6412e82
	fmt.Println(`查找： 2748807c-e815-41a4-8946-479c7f92d0f8`)
	c, err := wx.ContactByGGID(`2a711e80-46c2-4209-8433-aa02a6412e82`)
	if err == nil {
		fmt.Println(c.NickName)
		fmt.Println()
	}
}

func (t *tester) WechatDidLogout(wechat *wechat.WeChat) {}
