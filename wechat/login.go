package wechat

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/KevinGong2013/ggbot/utils"
	"github.com/skratchdot/open-golang/open"
)

var cookieCachePath = `.ggbot/cookie-cache`

// UUIDProcessor scan this uuid
type UUIDProcessor interface {
	ProcessUUID(uuid string) error
	UUIDDidConfirm(err error)
}

// implements UUIDProcessor
type defaultUUIDProcessor struct {
	path string
}

func (dp *defaultUUIDProcessor) ProcessUUID(uuid string) error {
	// 2.``
	path, err := utils.FetchORCodeImage(uuid)

	if err != nil {
		return err
	}
	logger.Debugf(`qrcode image path: %s`, path)

	// 3.
	go func() {
		dp.path = path
		open.Start(path)
	}()
	logger.Info(`please scan ORCode by wechat mobile application`)

	return nil
}

func (dp *defaultUUIDProcessor) UUIDDidConfirm(err error) {
	if len(dp.path) > 0 {
		utils.DeleteFile(dp.path)
	}
}

var retryTimes = time.Duration(1)

var loginState = make(chan int) // -1 登录失败 1登录成功

type initRequest struct {
	BaseRequest *BaseRequest
}

type initResp struct {
	Response
	User    Contact
	Skey    string
	SyncKey map[string]interface{}
}

func (wechat *WeChat) reLogin() error {

	client, err := newClient()
	if err != nil {
		return err
	}

	wechat.Client = client

	err = wechat.beginLoginFlow()
	if err != nil {
		return err
	}

	return nil
}

// run is used to login to wechat server. Need end user scan orcode.
func (wechat *WeChat) beginLoginFlow() error {

	logger.Info(`wait a moment, prepare login parameters ... ...`)

	redirectURL, err := wechat.quickLogin()

	if err != nil {
		utils.DeleteFile(cookieCachePath)
		redirectURL = ``
		logger.Warn(err)

		// 1.
		uuid, err := wechat.fetchUUID()

		if err != nil {
			return err
		}

		// 2.
		err = wechat.UUIDProcessor.ProcessUUID(uuid)

		if err != nil {
			return err
		}

		// 3.
		code, tip := ``, 1

		for code != httpOK {
			redirectURL, code, tip, err = wechat.waitConfirmUUID(uuid, tip)
			if err != nil {
				wechat.UUIDProcessor.UUIDDidConfirm(err)
				return err
			}
		}

		wechat.UUIDProcessor.UUIDDidConfirm(nil)

		req, _ := http.NewRequest(`GET`, redirectURL, nil)

		// 4.
		if err = wechat.login(req); err != nil {
			return err
		}
	}

	//5.
	index := strings.LastIndex(redirectURL, "/")
	if index == -1 {
		index = len(redirectURL)
	}
	wechat.BaseURL = redirectURL[:index]

	// 6.
	return wechat.init()
}

func (wechat *WeChat) quickLogin() (string, error) {

	file, err := os.Open(cookieCachePath)
	if err != nil {
		return ``, err
	}
	bs, err := ioutil.ReadAll(file)
	if err != nil {
		return ``, err
	}

	var info map[string]interface{}
	err = json.Unmarshal(bs, &info)
	if err != nil {
		return ``, err
	}
	url := info[`url`].(string)
	cookies := info[`cookies`].([]interface{})
	req, _ := http.NewRequest(`GET`, url, nil)
	for _, c := range cookies {
		b, _ := json.Marshal(c)
		var cookie *http.Cookie
		e := json.Unmarshal(b, cookie)
		if e == nil {
			req.AddCookie(cookie)
		}
	}

	logger.Debug(`quick login flow`)

	return url, wechat.login(req)
}

func (wechat *WeChat) fetchUUID() (string, error) {

	jsloginURL := "https://login.weixin.qq.com/jslogin"

	params := url.Values{}
	params.Set("appid", "wx782c26e4c19acffb")
	params.Set("fun", "new")
	params.Set("lang", "zh_CN")
	params.Set("_", strconv.FormatInt(time.Now().Unix(), 10))

	resp, err := wechat.Client.PostForm(jsloginURL, params)
	if err != nil {
		return ``, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ``, err
	}
	ds := string(data)

	code, err := utils.Search(ds, `window.QRLogin.code = `, `;`)

	if err != nil {
		return ``, err
	}

	if code != httpOK {
		err = fmt.Errorf("error code is unexpect:[%s], api result:[%s]", code, ds)
		return ``, err
	}

	uuid, err := utils.Search(ds, `window.QRLogin.uuid = "`, `";`)
	if err != nil {
		return ``, err
	}

	return uuid, nil
}

func (wechat *WeChat) waitConfirmUUID(uuid string, tip int) (redirectURI, code string, rt int, err error) {

	loginURL, rt := fmt.Sprintf("https://login.weixin.qq.com/cgi-bin/mmwebwx-bin/login?tip=%d&uuid=%s&_=%s", tip, uuid, strconv.FormatInt(time.Now().Unix(), 10)), tip
	resp, err := wechat.Client.Get(loginURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	ds := string(data)

	code, err = utils.Search(ds, `window.code=`, `;`)
	if err != nil {
		return
	}

	rt = 0
	switch code {
	case "201":
		logger.Debug(`scan successed, waitting wechat app send confirm request.`)
	case httpOK:
		redirectURI, err = utils.Search(ds, `window.redirect_uri="`, `";`)
		if err != nil {
			return
		}
		redirectURI += "&fun=new"
	default:
		err = fmt.Errorf("time out, will retry %v", err)
	}
	return
}

func (wechat *WeChat) login(req *http.Request) error {

	resp, err := wechat.Client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := resp.Body.(io.Reader)

	// full fill base request
	if err = xml.NewDecoder(reader).Decode(wechat.BaseRequest); err != nil {
		logger.Debug(err)
		return err
	}

	if wechat.BaseRequest.Ret != 0 { // 0 is success
		err = fmt.Errorf("login failed message:[%s]", wechat.BaseRequest.Message)
		return err
	}

	// added device id
	wechat.BaseRequest.DeviceID = `e999471493880231`

	info := map[string]interface{}{
		`cookies`: resp.Cookies(),
		`url`:     req.URL.String(),
	}

	b, err := json.Marshal(info)
	if err != nil {
		logger.Warnf(`save cookie error: %v`, err)
	} else {
		utils.CreateFile(cookieCachePath, b, false)
		logger.Info(`did upate cookie cache`)
	}

	return nil
}

func (wechat *WeChat) init() error {

	data, err := json.Marshal(initRequest{
		BaseRequest: wechat.BaseRequest,
	})
	if err != nil {
		return err
	}

	resp := new(initResp)
	apiURI := fmt.Sprintf("%s/webwxinit?%s&%s&r=%s", wechat.BaseURL, wechat.PassTicketKV(), wechat.SkeyKV(), utils.Now())

	if err = wechat.Excute(apiURI, bytes.NewReader(data), resp); err != nil {
		return err
	}

	wechat.BaseRequest.Skey = resp.Skey

	wechat.MySelf = resp.User
	wechat.syncKey = resp.SyncKey

	return nil
}

func (wechat *WeChat) keepAlive() {
	go func() {

		err := wechat.reLogin()

		if err != nil {
			logger.Errorf(`login failed: %v`, err)
			triggerAfter := time.After(time.Minute * retryTimes)
			logger.Warnf(`will retry login after %d minute(s)`, retryTimes)
			<-triggerAfter
			retryTimes++
			wechat.keepAlive()
			return
		}

		logger.Info(`CONGRATULATION login successed`)

		err = wechat.SyncContact()
		if err != nil {
			logger.Errorf(`sync contact error: %v`, err)
		}

		loginState <- 1

		err = wechat.listen(addMsg, modContact, delContact, modChatRoomMember)

		if err != nil {
			logger.Errorf(`can't listen server event %v`, err)
		}

		wechat.keepAlive() // if listen occured error will excute this cmd.
		loginState <- -1
		return
	}()
}
