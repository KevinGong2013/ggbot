package service

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/KevinGong2013/wechat"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"

	log "github.com/Sirupsen/logrus"
)

var innerlogger = log.WithFields(log.Fields{
	"module": "service",
})

// Wrapper ...
type Wrapper struct {
	uuidWebhookURL string
	r              *chi.Mux
	wx             *wechat.WeChat
}

// NewWrapper ...
func NewWrapper(uuidWebhookURL string) *Wrapper {

	r := chi.NewRouter()
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	w := &Wrapper{uuidWebhookURL, r, nil}
	w.initService()

	go http.ListenAndServe(`:3280`, r)

	return w
}

func (w *Wrapper) initService() {

	// w.r.Get(`/contacts/:nickName`, func(writer http.ResponseWriter, req *http.Request) {
	// 	if w.wx == nil {
	// 		http.Error(writer, `wechat did logout`, 500)
	// 	} else {
	// 		nn := chi.URLParam(req, `nickName`)
	// 		contacts, err := w.wx.ContactsByNickName(nn)
	// 		if err != nil {
	// 			http.Error(writer, `not found contact`, 404)
	// 		} else {
	// 			bs, _ := json.Marshal(contacts)
	// 			writer.Write(bs)
	// 		}
	// 	}
	// })

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

// Forward data to webhook
func (w *Wrapper) Forward(url string, data interface{}) error {

	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = http.Post(url, `application/json`, bytes.NewReader(bs))

	return err
}
