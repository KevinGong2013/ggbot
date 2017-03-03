package main

import (
	"io/ioutil"
	"os"

	"github.com/KevinGong2013/ggbot/uuidprocessor"
	"github.com/KevinGong2013/wechat"
	"github.com/Sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var logger = logrus.WithFields(logrus.Fields{
	"module": "ggbot",
})

var confPath = `conf.yaml`

func main() {

	tf := logrus.TextFormatter{}
	tf.FullTimestamp = true
	tf.TimestampFormat = `2006-01-02 15:04:05`
	logrus.SetFormatter(&tf)

	conf, err := readConf()
	if err != nil {
		conf, err = createDefaultConf()
		if err != nil {
			panic(err)
		}
	}

	options := wechat.DefaultConfigure()

	if conf[`showQRCodeOnTerminal`].(bool) {
		options.Processor = uuidprocessor.New()
	}

	bot, err := wechat.AwakenNewBot(options)
	if err != nil {
		panic(err)
	}

	features, _ := conf[`features`].(map[interface{}]interface{})

	var t *tuling
	var x *xiaoice
	var g *guard
	var a *assisant

	tl, _ := features[`tuling`].(map[interface{}]interface{})
	if tl[`enable`].(bool) {
		if ak, ok := tl[`api-key`].(string); ok {
			// 添加图灵自动回复
			t = newTuling(ak, bot)
		}
	}

	xi, _ := features[`xiaoice`].(map[interface{}]interface{})
	if xi[`enable`].(bool) {
		// 添加小冰自动回复
		x = newXiaoice(bot)
	}

	aa, _ := features[`assistant`].(map[interface{}]interface{})
	if aa[`enable`].(bool) {
		if owner, ok := aa[`ownerGGID`].(string); ok {
			// 加群欢迎语和简单的签到
			a = newAssisant(bot, owner)
		}
	}

	gg, _ := features[`guard`].(map[interface{}]interface{})
	if gg[`enable`].(bool) {
		// 添加图灵自动回复
		g = newGuard(bot)
	}

	bot.Handle(`/msg`, func(evt wechat.Event) {
		logger.Debug(`begin handle [/msg]`)
		data := evt.Data.(wechat.EventMsgData)
		if t != nil {
			go t.autoReplay(data)
		}
		if x != nil {
			go x.autoReplay(data)
		}
		if g != nil {
			go g.autoAcceptAddFirendRequest(data)
		}
		if a != nil {
			go a.handle(data)
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
	})

	bot.Go()
}

func createDefaultConf() (map[string]interface{}, error) {

	conf := map[string]interface{}{
		`showQRCodeOnTerminal`: false,
		`features`: map[string]interface{}{
			`assistant`: map[string]interface{}{
				`enable`:    true,
				`ownerGGID`: `46feef79-ac7d-46df-9e46-302502dfc436`,
			},
			`guard`: map[string]interface{}{
				`enable`: true,
			},
			`tuling`: map[string]interface{}{
				`enable`: false,
				`apikey`: ``,
			},
			`xiaoice`: map[string]interface{}{
				`enable`: true,
			},
		},
	}
	data, err := yaml.Marshal(conf)
	if err != nil {
		return nil, err
	}

	return conf, createFile(confPath, data, false)
}

func readConf() (map[string]interface{}, error) {

	file, err := os.Open(confPath)
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = yaml.Unmarshal(buf, &result)
	return result, err
}

func createFile(name string, data []byte, isAppend bool) (err error) {

	defer func() {
		if err != nil {
			logger.Error(err)
		}
	}()

	oflag := os.O_CREATE | os.O_WRONLY
	if isAppend {
		oflag |= os.O_APPEND
	} else {
		oflag |= os.O_TRUNC
	}

	file, err := os.OpenFile(name, oflag, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.Write(data)

	return
}
