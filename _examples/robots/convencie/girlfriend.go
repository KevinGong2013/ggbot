package main

import (
	"fmt"
	"time"

	"github.com/KevinGong2013/ggbot/modules/convenience"
	"github.com/KevinGong2013/ggbot/modules/gguuid"
	"github.com/KevinGong2013/ggbot/wechat"
)

func main() {

	gfNickName := `老婆`

	bot, err := wechat.WakeUp(gguuid.New())
	if err != nil {
		panic(err)
	}

	cm := convenience.NewMsgStream()
	bot.RegisterModule(cm)

	// 收到单聊消息以后回复
	cm.Handle(fmt.Sprintf(`/msg/solo/%s`, gfNickName), func(e convenience.Event) {
		uns, _ := bot.UserNameByNickName(gfNickName)
		for _, un := range uns {
			bot.SendTextMsg(`我收到啦`, un)
		}
	})

	// 收到群聊消息以后回复
	cm.Handle(fmt.Sprintf(`/msg/group/%s`, gfNickName), func(e convenience.Event) {
		uns, _ := bot.UserNameByNickName(gfNickName)
		for _, un := range uns {
			bot.SendTextMsg(`你说的真好`, un)
		}
	})

	// 比如每5分钟执行一次的定时任务
	cm.AddTimer(5 * time.Minute)
	cm.Handle(`/timer/5m`, func(e convenience.Event) {
		data := e.Data.(convenience.TimerEventData)
		fmt.Print(data)
		uns, _ := bot.UserNameByNickName(gfNickName)
		for _, un := range uns {
			bot.SendTextMsg(`你在忙什么呀`, un)
		}
	})

	// 比如每天早上10点准时签到
	cm.AddTiming(`9:00`)
	cm.Handle(`/timing/9:00`, func(e convenience.Event) {
		uns, _ := bot.UserNameByNickName(gfNickName)
		bot.SendTextMsg(`9点啦，快起床上班了`, uns[1]) // 每女朋友会panic哦
	})

	cm.Listen()
}
