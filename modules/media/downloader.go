package media

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	filetype "gopkg.in/h2non/filetype.v1"

	log "github.com/Sirupsen/logrus"
	"github.com/KevinGong2013/ggbot/utils"
	"github.com/KevinGong2013/ggbot/wechat"
)

var logger = log.WithFields(log.Fields{
	"module": "media-downloader",
})

// Downloader ...
type Downloader struct {
	dir string
	wx  *wechat.WeChat
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

func (d *Downloader) download(url string, name string) {
	if d.wx != nil {
		req, err := http.NewRequest(`GET`, url, nil)
		if err != nil {
			logger.Error(err)
			return
		}

		req.Header.Set(`Range`, `bytes=0-`) // 只有小视频才需要加这个headers

		resp, err := d.wx.Client.Do(req)
		defer resp.Body.Close()

		if err != nil {
			logger.Error(err)
		} else {
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error(err)
				return
			}

			t, err := filetype.Get(data)
			if err != nil {
				logger.Error(err)
				return
			}

			err = utils.CreateFile(filepath.Join(d.dir, name+`.`+t.Extension), data, false)
			if err == nil {
				logger.Debugf(`download finished %s`, name)
			} else {
				logger.Error(err)
			}
		}
	}
}
