package media

import (
	"os"

	wx "github.com/KevinGong2013/ggbot/wechat"
	log "github.com/Sirupsen/logrus"
)

var logger = log.WithFields(log.Fields{
	"module": "media-downloader",
})

// Downloader ...
type Downloader struct {
	dir string
	wx  *wx.WeChat
}

// NewDownloader ...
func NewDownloader(path string) (*Downloader, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &Downloader{path, nil}, nil
}

// WechatDidLogin ...
func (d *Downloader) WechatDidLogin(wechat *wx.WeChat) {
	d.wx = wechat
}

// WechatDidLogout ...
func (d *Downloader) WechatDidLogout(wechat *wx.WeChat) {
}

// MapMsgs ...
func (d *Downloader) MapMsgs(msg *wx.CountedContent) {
}

// HandleMsgs ...
func (d *Downloader) HandleMsgs(msg *wx.CountedContent) {
	for _, m := range msg.Content {
		if m[`isMediaMsg`].(bool) {
			url, _ := m[`MediaMsgDownloadUrl`].(string)
			name, _ := m[`MsgId`].(string)
			go func() {
				p, err := d.wx.DownloadMedia(url, d.dir+`/`+name)
				if err != nil {
					logger.Error(err)
				} else {
					logger.Infof(`did download file to %s`, p)
				}
			}()
		}
	}
}
