package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/KevinGong2013/ggbot/wechat"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"

	wx "github.com/KevinGong2013/ggbot/wechat"
	log "github.com/Sirupsen/logrus"
)

var innerlogger = log.WithFields(log.Fields{
	"module": "service",
})

// Wrapper ...
type Wrapper struct {
	f  *Forward
	r  *chi.Mux
	wx *wechat.WeChat
}

// NewWrapper ...
func NewWrapper(msgWebhookURL, contactWebhookURL, loginStateWebhookURL, uuidWebhookURL string) *Wrapper {

	f := &Forward{msgWebhookURL, contactWebhookURL, loginStateWebhookURL, uuidWebhookURL}
	r := chi.NewRouter()
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	w := &Wrapper{f, r, nil}
	w.initService()

	go http.ListenAndServe(`:3280`, r)

	return w
}

func (w *Wrapper) initService() {

	w.r.Get(`/contacts/:nickName`, func(writer http.ResponseWriter, req *http.Request) {
		if w.wx == nil {
			http.Error(writer, `wechat did logout`, 500)
		} else {
			nn := chi.URLParam(req, `nickName`)
			contact, err := w.wx.ContactByNickName(nn)
			if err != nil {
				http.Error(writer, `not found contact`, 404)
			} else {
				bs, _ := json.Marshal(contact)
				writer.Write(bs)
			}
		}
	})

	w.r.Post(`/msg`, func(writer http.ResponseWriter, req *http.Request) {
		if w.wx == nil {
			http.Error(writer, `wechat did logout`, 500)
		} else {
			var body map[string]interface{}
			defer req.Body.Close()
			bs, _ := ioutil.ReadAll(req.Body)
			err := json.Unmarshal(bs, &body)
			if err != nil {
				http.Error(writer, err.Error(), 500)
			} else {
				content, _ := body[`content`].(string)
				to, _ := body[`to`].(string)
				w.wx.SendTextMsg(content, to)
			}
		}
	})
}

// WechatDidLogin ...
func (w *Wrapper) WechatDidLogin(wechat *wx.WeChat) {
	w.wx = wechat
	if len(w.f.loginStateWebhookURL) > 0 {
		go w.f.forward(w.f.loginStateWebhookURL, map[string]bool{
			`isLogin`: true,
		})
	}
}

// WechatDidLogout ...
func (w *Wrapper) WechatDidLogout(wechat *wx.WeChat) {
	if len(w.f.loginStateWebhookURL) > 0 {
		go w.f.forward(w.f.loginStateWebhookURL, map[string]bool{
			`isLogin`: false,
		})
	}
}

// MapContact ...
func (w *Wrapper) MapContact(contact map[string]interface{}) {
	return
}

// HandleContact ...
func (w *Wrapper) HandleContact(contact map[string]interface{}) {
	if len(w.f.contactWebhookURL) > 0 {
		go w.f.forward(w.f.contactWebhookURL, contact)
	}
}

// MapMsgs ...
func (w *Wrapper) MapMsgs(msg *wx.CountedContent) {
	return
}

// HandleMsgs ...
func (w *Wrapper) HandleMsgs(msg *wx.CountedContent) {
	if len(w.f.msgWebhookURL) > 0 {
		go w.f.forward(w.f.msgWebhookURL, msg)
	}
}
