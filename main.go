package main

import (
	"errors"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	yaml "gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"

	"github.com/KevinGong2013/ggbot/modules/assistant"
	"github.com/KevinGong2013/ggbot/modules/echo"
	"github.com/KevinGong2013/ggbot/modules/gguuid"
	"github.com/KevinGong2013/ggbot/modules/media"
	"github.com/KevinGong2013/ggbot/modules/service"
	"github.com/KevinGong2013/ggbot/modules/tuling"
	"github.com/KevinGong2013/ggbot/modules/ui"
	"github.com/KevinGong2013/ggbot/modules/xiaoice"
	"github.com/KevinGong2013/ggbot/utils"
	"github.com/KevinGong2013/ggbot/wechat"
)

var logger = log.WithFields(log.Fields{
	"module": "main",
})

// Conf ...
type Conf struct {
	LogLevel int `yaml:"log-level"`
	Debug    bool
	Modules  map[string]map[string]interface{}
}

var confPath = `conf.yaml`

var version = `0.9.1`
var date = `2017-02-22`

func main() {

	logger.Infof(`version: %s, date: %s`, version, date)

	tf := log.TextFormatter{}
	tf.FullTimestamp = true
	tf.TimestampFormat = `2006-01-02 15:04:05`
	log.SetFormatter(&tf)

	var conf = &Conf{}

	err := readConf(conf)

	if err != nil {
		logger.Error(err)
		logger.Info(`读取配置文件出错，GGBot 将自动生成默认的配置文件`)

		conf, err = createDefaultConf()
		if err != nil {
			panic(err)
		}
	}

	wechat.Debug = conf.Debug
	switch conf.LogLevel {
	case 0:
		log.SetLevel(log.DebugLevel)
	case 1:
		log.SetLevel(log.InfoLevel)
	case 2:
		log.SetLevel(log.WarnLevel)
	case 3:
		log.SetLevel(log.ErrorLevel)
	}

	var up wechat.UUIDProcessor
	if conf.Modules[`gguuid`] != nil {
		up = gguuid.New()
	}

	wxbot, err := wechat.WakeUp(up)
	if err != nil {
		logger.Error(err)
		return
	}

	err = registerModules(conf, wxbot)
	if err != nil {
		panic(err)
	}

	waitForExit()
}

func readConf(conf *Conf) error {

	file, err := os.Open(confPath)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(buf, &conf)
}

func createDefaultConf() (*Conf, error) {

	conf := &Conf{
		LogLevel: 0,
		Debug:    true,
		Modules: map[string]map[string]interface{}{
			`assistant`: {
				`groupName`: `GGBot测试群`,
				`welcome`:   `大家鼓掌欢迎 👏👏👏`,
			},
			`echo`:   {},
			`gguuid`: {},
			`media`: {
				`path`: `.ggbot/media`,
			},
			`service`: {
				`msg-webhook`:         `http://127.0.0.1:3288/msg`,
				`contact-webhook`:     `http://127.0.0.1:3288/contact`,
				`login-state-webhook`: `http://127.0.0.1:3288/login_state`,
				`uuid-webhook`:        `http://127.0.0.1:3288/uuid`,
			},
			`storage`: {
				`path`: `.ggbot/db`,
			},
			`tuling`: {
				`api-key`: `b6b93435df0e4b71aff460231b89d8eb`,
			},
			// `ui`:      {},
			`xiaoice`: {},
		},
	}

	buff, err := yaml.Marshal(conf)
	if err != nil {
		return nil, err
	}

	return conf, utils.CreateFile(confPath, buff, false)
}

func registerModules(conf *Conf, bot *wechat.WeChat) error {

	// 1. 不能同时注册gguuid和ui
	if conf.Modules[`ui`] != nil && conf.Modules[`gguuid`] != nil {
		return errors.New(`[ui]模块和[gguid]模块不能共存，请二选一]`)
	}

	for k, v := range conf.Modules {
		switch k {
		case `assistant`:
			gn := v[`groupName`].(string)
			welcome := v[`welcome`].(string)
			bot.RegisterModule(assistant.NewAssistant(gn, welcome))
		case `echo`:
			bot.RegisterModule(echo.New())
		case `media`:
			path := v[`path`].(string)
			d, err := media.NewDownloader(path)
			if err == nil {
				bot.RegisterModule(d)
			} else {
				logger.Warnf(`regist media module failed err: %v`, err)
			}
		case `service`:
			msgWebhook := v[`msg-webhook`].(string)
			contactWebhook := v[`contact-webhook`].(string)
			loginWebhook := v[`login-state-webhook`].(string)
			uuidWebhook := v[`uuid-webhook`].(string)
			bot.RegisterModule(service.NewWrapper(msgWebhook, contactWebhook, loginWebhook, uuidWebhook))
		case `tuling`:
			apiKey := v[`api-key`].(string)
			if len(apiKey) == 0 {
				logger.Warn(`regsit tuling module failed api-key is needed`)
			} else {
				bot.RegisterModule(tuling.NewBrain(apiKey))
			}
		case `ui`:
			path := v[`path`].(string)
			u := ui.NewUI(path)
			bot.RegisterModule(u)
			go u.Loop()
		case `xiaoice`:
			bot.RegisterModule(xiaoice.NewBrain())
		}
	}

	return nil
}

func waitForExit() os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
	return <-c
}
