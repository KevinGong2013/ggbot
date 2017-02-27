package tuling

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	wx "github.com/KevinGong2013/ggbot/wechat"
	log "github.com/Sirupsen/logrus"
)

var logger = log.WithFields(log.Fields{
	"module": "tuling",
})

// Brain ...
type Brain struct {
	key string
	wx  *wx.WeChat
}

// NewBrain ...
func NewBrain(apikey string) *Brain {
	return &Brain{apikey, nil}
}

// WechatDidLogin ...
func (b *Brain) WechatDidLogin(wechat *wx.WeChat) {
	b.wx = wechat
}

// WechatDidLogout ...
func (b *Brain) WechatDidLogout(wechat *wx.WeChat) {
}

// MapMsgs ...
func (b *Brain) MapMsgs(msg *wx.CountedContent) {
	for _, m := range msg.Content {

		msgType, _ := m[`MsgType`].(float64)
		if msgType != 1 { // 目前只回复文字消息
			m[`needTulingResponse`] = false
			continue
		}

		isSendByMySelf, _ := m[`IsSendByMySelf`].(bool)
		if isSendByMySelf {
			continue
		}
		from, _ := m[`FromUserName`].(string)
		contact, err := b.wx.ContactByUserName(from)
		if err != nil {
			logger.Error(err)
			m[`needTulingResponse`] = false
			continue
		}

		switch contact.Type {
		case wx.Friend:
			m[`needTulingResponse`] = true
			m[`info`] = m[`Content`]
			m[`to`] = m[`FromUserName`]
			m[`userid`] = contact.NickName
		case wx.Offical:
			m[`needTulingResponse`] = false
		case wx.Group:
			m[`needTulingResponse`] = m[`AtMe`]
			m[`info`] = m[`Content`]
			m[`to`] = m[`FromUserName`]
			m[`userid`] = m[`ActualNickName`]
		}
	}
	return
}

// HandleMsgs ...
func (b *Brain) HandleMsgs(msg *wx.CountedContent) {
	for _, m := range msg.Content {
		f, _ := m[`needTulingResponse`].(bool)
		if f {
			info, _ := m[`info`].(string)
			to, _ := m[`to`].(string)
			userid, _ := m[`userid`].(string)
			go b.autoReplay(info, to, userid)
		}
	}
}

func (b *Brain) autoReplay(info, to, userid string) {
	if b.wx == nil {
		return
	}
	replay, err := b.response(info, to, userid)
	if err == nil {
		logger.Debugf(`receive: %s from: %s nickname: %s, replay: %s`, info, to, userid, replay)
		b.wx.SendTextMsg(replay, to)
	} else {
		logger.Error(err)
		b.wx.SendTextMsg(`伦家不知道你在说什么啦 嘤嘤嘤   (｡•ˇ‸ˇ•｡)哼`, to)
	}
}

func (b *Brain) response(msg, to, userid string) (string, error) {

	values := url.Values{}

	values.Add(`key`, b.key)
	values.Add(`info`, msg)
	values.Add(`userid`, userid)

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

	logger.Errorf(`info: [%s], userid: [%s]`, msg, userid)
	logger.Error(result)
	return ``, errors.New(`tuling api unkonw error`)
}
