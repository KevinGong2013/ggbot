package wechat

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/allegro/bigcache"
	"github.com/KevinGong2013/ggbot/utils"
)

var logger = log.WithFields(log.Fields{
	"module": "wechat",
})

// Debug is a flag, if turn it on, will record all wx `send` api result to json file.
var Debug = true

// Version is wechat's main version
var Version = `0.9.0`

const httpOK = `200`

var debugPath = `.ggbot/debug/`

// BaseRequest is a base for all wx api request.
type BaseRequest struct {
	XMLName xml.Name `xml:"error" json:"-"`

	Ret        int    `xml:"ret" json:"-"`
	Message    string `xml:"message" json:"-"`
	Wxsid      string `xml:"wxsid" json:"Sid"`
	Skey       string `xml:"skey"`
	DeviceID   string `xml:"-"`
	Wxuin      int64  `xml:"wxuin" json:"Uin"`
	PassTicket string `xml:"pass_ticket" json:"-"`
}

// Caller is a interface, All response need implement this.
type Caller interface {
	IsSuccess() bool
	Error() error
}

// Response is a wrapper.
type Response struct {
	BaseResponse *BaseResponse
}

// IsSuccess flag this request is success or failed.
func (response *Response) IsSuccess() bool {
	return response.BaseResponse.Ret == 0
}

// response's error msg.
func (response *Response) Error() error {
	return fmt.Errorf("error message:[%s]", response.BaseResponse.ErrMsg)
}

// BaseResponse for all api resp.
type BaseResponse struct {
	Ret    int
	ErrMsg string
}

// WeChat container a default http client and base request.
type WeChat struct {
	Client        *http.Client
	BaseURL       string
	BaseRequest   *BaseRequest
	UUIDProcessor UUIDProcessor
	MySelf        Contact
	IsLogin       bool
	contactCache  *bigcache.BigCache
	nicknameCache *bigcache.BigCache
	syncKey       map[string]interface{}
	syncHost      string
}

// NewWeChat is desined for Create a new Wechat instance.
func newWeChat(up UUIDProcessor) (*WeChat, error) {

	if up == nil {
		return nil, errors.New(`UUID Processor must be not nil`)
	}
	if _, err := os.Stat(debugPath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(debugPath, os.ModePerm)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	client, err := newClient()
	if err != nil {
		return nil, err
	}

	bc, err := bigcache.NewBigCache(bigcache.DefaultConfig(7 * 24 * time.Hour))
	if err != nil {
		return nil, err
	}
	nc, err := bigcache.NewBigCache(bigcache.DefaultConfig(7 * 24 * time.Hour))
	if err != nil {
		return nil, err
	}

	baseReq := new(BaseRequest)
	baseReq.Ret = 1

	wechat := &WeChat{
		Client:        client,
		BaseRequest:   baseReq,
		UUIDProcessor: up,
		IsLogin:       false,
		contactCache:  bc,
		nicknameCache: nc,
	}

	return wechat, nil
}

func newClient() (*http.Client, error) {

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	transport := http.Transport{
		Dial: (&net.Dialer{
			Timeout: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 30 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: &transport,
		Jar:       jar,
		Timeout:   30 * time.Second,
	}

	return client, nil
}

// WakeUp is start point for wx bot.
func WakeUp(up UUIDProcessor) (*WeChat, error) {

	if up == nil {
		up = new(defaultUUIDProcessor)
	}

	wechat, err := newWeChat(up)
	if err != nil {
		return nil, err
	}

	wechat.handleServerEvent()
	wechat.keepAlive()

	// 处理群消息的
	wechat.RegisterModule(newFlatten(wechat))

	return wechat, nil
}

// ExcuteRequest is desined for perform http request
func (wechat *WeChat) ExcuteRequest(req *http.Request, call Caller) error {

	ps := strings.Split(req.URL.Path, `/`)
	lastP := strings.Split(ps[len(ps)-1], `?`)[0][5:]
	filename := debugPath + lastP

	if Debug {
		reqData, _ := httputil.DumpRequest(req, true)
		utils.CreateFile(filename+`_req.kv`, reqData, false)
	}

	resp, err := wechat.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := resp.Body.(io.Reader)

	if Debug {

		data, e := ioutil.ReadAll(reader)
		if e != nil {
			return e
		}

		utils.CreateFile(filename+`_resp.json`, data, true)
		reader = bytes.NewReader(data)
	}

	if err = json.NewDecoder(reader).Decode(call); err != nil {
		return err
	}

	if !call.IsSuccess() {
		return call.Error()
	}

	return nil
}

// Excute a http request by default http client.
func (wechat *WeChat) Excute(path string, body io.Reader, call Caller) error {
	method := "GET"
	if body != nil {
		method = "POST"
	}
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set(`User-Agent`, `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_2) AppleWebKit/602.3.12 (KHTML, like Gecko) Version/10.0.2 Safari/602.3.12`)

	return wechat.ExcuteRequest(req, call)
}

// PassTicketKV return a string like `pass_ticket=sdfewsvdwd=`
func (wechat *WeChat) PassTicketKV() string {
	return fmt.Sprintf(`pass_ticket=%s`, wechat.BaseRequest.PassTicket)
}

// SkeyKV return a string like `skey=ewfwoefjwofjskfwes`
func (wechat *WeChat) SkeyKV() string {
	return fmt.Sprintf(`skey=%s`, wechat.BaseRequest.Skey)
}

// just for debug
type nopCloser struct {
	io.Reader
}

func (np *nopCloser) Close() error { return nil }
