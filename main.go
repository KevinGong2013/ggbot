package main

import (
	"io/ioutil"
	"os"

	"github.com/KevinGong2013/ggbot/service"
	"github.com/KevinGong2013/ggbot/uuidprocessor"
	"github.com/KevinGong2013/wechat"
	"github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var logger = logrus.WithFields(logrus.Fields{
	"module": "ggbot",
})

// Config ..
type Config struct {
	ShowQRCodeOnTerminal bool
	FuzzyDiff            bool
	UniqueGroupMember    bool
	Features             struct {
		Assistant struct {
			Enable    bool
			OwnerGGID string
		}
		Guard struct {
			Enable bool
		}
		Tuling struct {
			Enable bool
			Key    string
		}
		Xiaoice struct {
			Enable bool
		}
		WebhookService struct {
			Enable            bool
			MsgWebhook        string
			LoginStateWebhook string
			UUIDWebhook       string
		}
	}
}

var config = Config{}

func main() {

	tf := logrus.TextFormatter{}
	tf.FullTimestamp = true
	tf.TimestampFormat = `2006-01-02 15:04:05`
	logrus.SetFormatter(&tf)
	logrus.SetLevel(logrus.DebugLevel)

	path := `conf.yaml`
	_, err := os.Stat(path)
	if !(err == nil || os.IsExist(err)) {
		config.ShowQRCodeOnTerminal = false
		config.Features.Tuling.Key = ``
		config.Features.Tuling.Enable = false
		config.FuzzyDiff = true
		data, _ := yaml.Marshal(config)
		createFile(path, data)
	}
	data, _ := ioutil.ReadFile(path)
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		logger.Error(`配置文件不正确`)
	}

	options := wechat.DefaultConfigure()

	// 是否开启联系人模糊匹配
	options.FuzzyDiff = config.FuzzyDiff
	options.UniqueGroupMember = config.UniqueGroupMember

	if config.ShowQRCodeOnTerminal {
		options.Processor = uuidprocessor.New()
	}

	var webhookService *service.Wrapper
	if config.Features.WebhookService.Enable {
		webhookService = service.NewWrapper(config.Features.WebhookService.UUIDWebhook)
	}

	if webhookService != nil && len(config.Features.WebhookService.UUIDWebhook) > 0 {
		options.Processor = webhookService
	}

	bot, err := wechat.AwakenNewBot(options)
	if err != nil {
		panic(err)
	}

	t := newTuling(config.Features.Tuling.Key, bot)
	x := newXiaoice(bot)
	a := newAssisant(bot, config.Features.Assistant.OwnerGGID)
	g := newGuard(bot)

	bot.Handle(`/msg`, func(evt wechat.Event) {
		logger.Debug(`begin handle [/msg]`)
		data := evt.Data.(wechat.EventMsgData)
		if config.Features.Tuling.Enable {
			go t.autoReplay(data)
		}
		if config.Features.Xiaoice.Enable {
			go x.autoReplay(data)
		}
		if config.Features.Guard.Enable {
			go g.autoAcceptAddFirendRequest(data)
		}
		if config.Features.Assistant.Enable {
			go a.handle(data)
		}
		if webhookService != nil && len(config.Features.WebhookService.MsgWebhook) > 0 {
			go webhookService.Forward(config.Features.WebhookService.MsgWebhook, data)
		}
	})

	bot.Handle(`/login`, func(arg2 wechat.Event) {
		isSuccess := arg2.Data.(int) == 1
		if isSuccess && x != nil {
			if cs, err := bot.ContactsByNickName(`小冰`); err == nil {
				for _, c := range cs {
					if c.Type == wechat.Offical {
						x.un = c.UserName // 更新小冰的UserName
						break
					}
				}
			}
		}
		if webhookService != nil && len(config.Features.WebhookService.LoginStateWebhook) > 0 {
			go webhookService.Forward(config.Features.WebhookService.LoginStateWebhook, arg2.Data)
		}
	})

	bot.Go()
}

func createFile(name string, data []byte) (err error) {

	defer func() {
		if err != nil {
			logger.Error(err)
		}
	}()

	oflag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC

	file, err := os.OpenFile(name, oflag, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.Write(data)

	return
}
