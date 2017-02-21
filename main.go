package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/KevinGong2013/ggbot/wechat"

	"github.com/KevinGong2013/ggbot/modules/echo"
	"github.com/KevinGong2013/ggbot/modules/service"
)

var logger = log.WithFields(log.Fields{
	"module": "main",
})

var showCUI = flag.Bool(`cui`, false, `是否要启用图形界面 默认不启用`)
var mediaPath = flag.String(`mp`, `.ggbot/media`, `多媒体文件存放根目录`)
var dbPath = flag.String(`dp`, `.ggbot/db`, `联系人和消息存放目录`)

func main() {

	flag.Parse()

	log.SetLevel(log.DebugLevel)

	wxbot, err := wechat.WakeUp(nil)
	if err != nil {
		logger.Error(err)
		return
	}

	// d, err := media.NewDownloader(*mediaPath)
	// if err == nil {
	// 	wxbot.RegisterModule(d)
	// } else {
	// 	logger.Error(err)
	// }

	// st, err := storage.NewStorage(*dbPath)
	// if err == nil {
	// 	wxbot.RegisterModule(st)
	// } else {
	// 	logger.Error(err)
	// }

	// wxbot.RegisterModule(tuling.NewBrain(`b6b93435df0e4b71aff460231b89d8eb`))
	// wxbot.RegisterModule(xiaoice.NewBrain())
	wxbot.RegisterModule(echo.New())
	wxbot.RegisterModule(service.NewWrapper(
		`http://127.0.0.1:3288/msg`,
		`http://127.0.0.1:3288/contact`,
		`http://127.0.0.1:3288/login_state`))

	// 	ui := ui.NewUI(*mediaPath)
	// 	wxbot.RegisterModule(ui)
	// 	ui.Loop()
	waitForExit()
}

func waitForExit() os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
	return <-c
}
