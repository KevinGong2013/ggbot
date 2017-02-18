package notice

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	wx "github.com/KevinGong2013/ggbot/wechat"
)

var logger = log.WithFields(log.Fields{
	"module": "notice",
})

// Notice module
type Notice struct {
	wx *wx.WeChat
	to string
}

type event struct {
	Kind string `json:"object_kind"`
	Name string `json:"user_name"`
}

// NewNotice ..
func NewNotice(noticeTo string) *Notice {
	n := &Notice{nil, noticeTo}
	http.HandleFunc(`/notice`, n.handle)
	return n
}

// WechatDidLogin ...
func (notice *Notice) WechatDidLogin(wechat *wx.WeChat) {
	notice.wx = wechat
}

// WechatDidLogout ...
func (notice *Notice) WechatDidLogout(wechat *wx.WeChat) {
	notice.wx = wechat
}

// Handler ..
func (notice *Notice) handle(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	var e event
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&e)
	if err != nil {
		logger.Error(`can't decoder push event`)
		return
	}

	msg := fmt.Sprintf(`Event: [%v], User: [%v]`, e.Kind, e.Name)

	if notice.wx != nil {
		notice.wx.SendTextMsg(msg, notice.to)
	} else {
		logger.Warnf(`cancel notice [%s] because wechat offline`, msg)
	}
}
