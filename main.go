package main

import (
	"io/ioutil"
	"os"

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

	bot, err := wechat.AwakenNewBot(nil)
	if err != nil {
		panic(err)
	}

	conf, err := readConf()
	if err != nil {
		conf, err = createDefaultConf()
		if err != nil {
			panic(err)
		}
	}
	logger.Debugf(`%v`, conf)
	features, _ := conf[`features`].(map[string]interface{})
	logger.Debugf(`%v`, features)
	var t *tuling
	var x *xiaoice
	var g *guard
	var a *assisant

	tl, _ := features[`tuling`].(map[string]interface{})
	if tl[`enable`].(bool) {
		// 添加图灵自动回复
		t = newTuling(`b6b93435df0e4b71aff460231b89d8eb`, bot)
	}

	xi, _ := features[`xiaoice`].(map[string]interface{})
	if xi[`enable`].(bool) {
		// 添加小冰自动回复
		x = newXiaoice(bot)
	}

	aa, _ := features[`assistant`].(map[string]interface{})
	if aa[`enable`].(bool) {
		// 加群欢迎语和简单的签到
		a = newAssisant(bot)
	}

	gg, _ := features[`guard`].(map[string]interface{})
	if gg[`enable`].(bool) {
		// 添加图灵自动回复
		g = newGuard(bot)
	}

	bot.Handle(`/msg`, func(evt wechat.Event) {
		data := evt.Data.(wechat.EventMsgData)
		t.autoReplay(data)
		x.autoReplay(data)
		g.autoAcceptAddFirendRequest(data)
		a.handle(data)
	})

	bot.Handle(`/login`, func(arg2 wechat.Event) {
		isSuccess := arg2.Data.(int) == 1
		if isSuccess {
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
		`features`: map[string]interface{}{
			`assistant`: map[string]interface{}{
				`enable`: true,
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
