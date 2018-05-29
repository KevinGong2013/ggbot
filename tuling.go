package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/KevinGong2013/wechat"
)

type tuling struct {
	key string
	bot *wechat.WeChat
}

func newTuling(key string, bot *wechat.WeChat) *tuling {
	return &tuling{key, bot}
}

func (t *tuling) autoReplay(data wechat.EventMsgData) {
	if data.IsSendedByMySelf {
		return
	}
	replay, err := t.response(data.Content, data.FromUserName)
	if err != nil {
		logger.Error(err)
		t.bot.SendTextMsg(`你接着说 ... `, data.FromUserName)
	} else {
		t.bot.SendTextMsg(replay, data.FromUserName)
	}
}

func (t *tuling) response(msg, to string) (string, error) {

	values := url.Values{}

	values.Add(`key`, t.key)
	values.Add(`info`, msg)
	values.Add(`userid`, to)

	resp, err := http.PostForm(`http://www.tuling123.com/openapi/api`, values)
	if err != nil {
		return ``, err
	}

	reader := resp.Body
	defer resp.Body.Close()

	result := make(map[string]interface{})

	err = json.NewDecoder(reader).Decode(&result)
	if err != nil {
		return ``, err
	}

	code := result[`code`].(float64)

	if code == 100000 {
		text := result[`text`].(string)
		return text, nil
	} else if code == 200000 {
		text := result[`text`].(string)
		url := result[`url`].(string)
		return text + `
		` + url, nil
	}

	logger.Errorf(`info: [%s], userid: [%s]`, msg, to)
	logger.Error(result)
	return ``, errors.New(`tuling api unkonw error`)
}
