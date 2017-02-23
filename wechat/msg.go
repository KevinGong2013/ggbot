package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/KevinGong2013/ggbot/utils"
	"github.com/KevinGong2013/ggbot/wechat/messages"
	"gopkg.in/h2non/filetype.v1"
)

type uploadMediaResponse struct {
	Response
	MediaID string `json:"MediaId"`
}

type sendMsgResponse struct {
	Response
	MsgID   string
	LocalID string
}

// Msg implement this interface, can added addition send by wechat
type Msg interface {
	Path() string
	To() string
	Content() map[string]interface{}
}

var mediaIndex = int64(1)

// SendMsg is desined to send Message to group or contact
func (wechat *WeChat) SendMsg(message Msg) error {

	if wechat.BaseRequest == nil {
		return fmt.Errorf(`wechat BaseRequest is empty`)
	}

	msg := baseMsg(message.To())

	for k, v := range message.Content() {
		msg[k] = v
	}
	msg[`FromUserName`] = wechat.MySelf.UserName

	buffer := new(bytes.Buffer)
	enc := json.NewEncoder(buffer)
	enc.SetEscapeHTML(false)

	err := enc.Encode(map[string]interface{}{
		`BaseRequest`: wechat.BaseRequest,
		`Msg`:         msg,
		`Scene`:       0,
	})

	if err != nil {
		return err
	}

	logger.Debugf(`sending [%s]`, msg[`LocalID`])

	resp := new(sendMsgResponse)

	apiURL := fmt.Sprintf(`%s/%s`, wechat.BaseURL, message.Path())

	if strings.Contains(apiURL, `?`) {
		apiURL = apiURL + `&` + wechat.PassTicketKV()
	} else {
		apiURL += `?` + wechat.PassTicketKV()
	}

	err = wechat.Excute(apiURL, buffer, resp)

	if err == nil {
		logger.Debugf(`sended [%s] MsgID=[%s]`, resp.LocalID, resp.MsgID)
	}

	return err
}

// SendTextMsg send text message
func (wechat *WeChat) SendTextMsg(msg, to string) error {
	textMsg := messages.NewTextMsg(msg, to)
	return wechat.SendMsg(textMsg)
}

// SendFile is desined to send contain attachment Message to group or contact.
// path must exit in local file system.
func (wechat *WeChat) SendFile(path, to string) error {
	msg, err := wechat.newMsg(path, to)
	if err != nil {
		return err
	}

	return wechat.SendMsg(msg)
}

// UploadMedia is a convernice method to upload attachment to wx cdn.
func (wechat *WeChat) UploadMedia(path string) (string, error) {

	info, err := os.Stat(path)

	if err != nil {
		return ``, err
	}

	file, err := os.Open(path)
	if err != nil {
		return ``, err
	}
	defer file.Close()

	kind, err := filetype.MatchFile(path)

	if err != nil {
		return ``, err
	}

	mediatype := `doc`
	if strings.HasPrefix(kind.MIME.Value, `image/`) {
		mediatype = `pic`
	}

	fields := map[string]string{
		`id`:                `WU_FILE_` + utils.Str(mediaIndex),
		`name`:              info.Name(),
		`type`:              kind.MIME.Value,
		`lastModifiedDate`:  info.ModTime().UTC().String(),
		`size`:              utils.Str(info.Size()),
		`mediatype`:         mediatype,
		`pass_ticket`:       wechat.BaseRequest.PassTicket,
		`webwx_data_ticket`: wechat.CookieDataTicket(),
	}

	media, err := json.Marshal(&map[string]interface{}{
		`BaseRequest`:   wechat.BaseRequest,
		`ClientMediaId`: utils.Now(),
		`TotalLen`:      utils.Str(info.Size()),
		`StartPos`:      0,
		`DataLen`:       utils.Str(info.Size()),
		`MediaType`:     4,
	})

	if err != nil {
		return ``, err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fw, err := writer.CreateFormFile(`filename`, info.Name())
	if err != nil {
		return ``, err
	}

	_, err = io.Copy(fw, file)
	if err != nil {
		return ``, err
	}

	for k, v := range fields {
		writer.WriteField(k, v)
	}

	writer.WriteField(`uploadmediarequest`, string(media))

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	writer.Close()

	urlOBJ, err := url.Parse(wechat.BaseURL)

	if err != nil {
		return ``, err
	}

	host := urlOBJ.Host

	urls := [2]string{
		fmt.Sprintf(`https://file.%s/cgi-bin/mmwebwx-bin/webwxuploadmedia?f=json`, host),
		fmt.Sprintf(`https://file2.%s/cgi-bin/mmwebwx-bin/webwxuploadmedia?f=json`, host),
	}

	for _, url := range urls {

		var req *http.Request
		req, err = http.NewRequest(`POST`, url, body)
		if err != nil {
			return ``, err
		}

		req.Header.Set(`Content-Type`, writer.FormDataContentType())

		resp := new(uploadMediaResponse)

		err = wechat.ExcuteRequest(req, resp)
		if err != nil {
			return ``, err
		}

		return resp.MediaID, nil
	}

	return ``, err
}

// DownloadMedia use to download a voice or immage msg
func (wechat *WeChat) DownloadMedia(url string, localPath string) (string, error) {

	req, err := http.NewRequest(`GET`, url, nil)
	if err != nil {
		return ``, err
	}

	req.Header.Set(`Range`, `bytes=0-`) // 只有小视频才需要加这个headers

	resp, err := wechat.Client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return ``, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ``, err
	}

	t, err := filetype.Get(data)
	if err != nil {
		return ``, err
	}

	path := filepath.Join(localPath + `.` + t.Extension)
	err = utils.CreateFile(path, data, false)
	if err != nil {
		return ``, err
	}

	return path, nil
}

// NewMsg create new message instance
func (wechat *WeChat) newMsg(filepath, to string) (Msg, error) {

	media, err := wechat.UploadMedia(filepath)

	if err != nil {
		return nil, err
	}

	kind, err := filetype.MatchFile(filepath)
	if err != nil {
		return nil, err
	}

	isImage := strings.HasPrefix(kind.MIME.Value, `image`)

	var msg Msg

	if isImage {
		if strings.HasSuffix(kind.MIME.Value, `gif`) {
			msg = messages.NewEmoticonMsgMsg(media, to)
		} else {
			msg = messages.NewFileMsg(`webwxsendmsgimg?fun=async&f=json`, media, to, 3, nil)
		}
	} else {
		info, _ := os.Stat(filepath)
		msg = messages.NewFileMsg(`webwxsendappmsg?fun=async&f=json`, media, to, 6, info)
	}

	return msg, err
}

func clientMsgID() string {
	return strconv.FormatInt(time.Now().Unix()*1000, 10) + strconv.Itoa(rand.Intn(10000))
}

func baseMsg(to string) map[string]interface{} {

	randomID := clientMsgID()

	msg := map[string]interface{}{
		`ToUserName`:  to,
		`LocalID`:     randomID,
		`ClientMsgId`: randomID,
	}

	return msg
}

// CookieDataTicket ...
func (wechat *WeChat) CookieDataTicket() string {

	url, err := url.Parse(wechat.BaseURL)

	if err != nil {
		return ``
	}

	ticket := ``

	cookies := wechat.Client.Jar.Cookies(url)

	for _, cookie := range cookies {
		if cookie.Name == `webwx_data_ticket` {
			ticket = cookie.Value
			break
		}
	}

	return ticket
}
