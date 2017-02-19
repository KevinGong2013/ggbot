package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/KevinGong2013/ggbot/utils"
)

// Contact is wx Account struct
type Contact struct {
	UserName          string
	NickName          string
	HeadImgURL        string `json:"HeadImgUrl"`
	RemarkName        string
	PYInitial         string
	PYQuanPin         string
	RemarkPYInitial   string
	RemarkPYQuanPin   string
	HideInputBarFlag  float64
	StarFriend        float64
	Sex               float64
	Signature         string
	AppAccountFlag    float64
	VerifyFlag        float64
	ContactFlag       float64
	WebWxPluginSwitch float64
	HeadImgFlag       float64
	SnsFlag           float64
	Province          string
	City              string
	Alias             string
	DisplayName       string
	KeyWord           string
	EncryChatRoomID   string `json:"EncryChatRoomId"`
	IsOwner           float64
	Type              int
	ChangeType        int // 0 修改 1 删除
	MemberCount       float64
	MemberList        []*Contact
}

type updateGroupRequest struct {
	BaseRequest
	Count int
	List  []string
}

type updateGroupMemberRequest struct {
	BaseRequest
}

type getContactResponse struct {
	Response
	MemberCount int
	MemberList  []map[string]interface{}
}

type batchGetContactResponse struct {
	Response
	Count       int
	ContactList []map[string]interface{}
}

const (
	// ContactTypeFriend friend
	ContactTypeFriend = 1
	// ContactTypeGroup group
	ContactTypeGroup = 2
	// ContactTypeOfficial official
	ContactTypeOfficial = 3
)

var maxCountOnceLoadGroupMember = 50

// To is contact's ID can be used in msg struct
func (contact *Contact) To() string {
	return contact.UserName
}

// SyncContact with Wechat server.
func (wechat *WeChat) SyncContact() error {

	wechat.resetCache()

	url := fmt.Sprintf(`%s/webwxgetcontact?%s&%s&r=%s`, wechat.BaseURL, wechat.PassTicketKV(), wechat.SkeyKV(), utils.Now())
	resp := new(getContactResponse)

	err := wechat.Excute(url, nil, resp)
	if err != nil {
		return err
	}

	var gs []string

	for _, v := range resp.MemberList {

		vf, _ := v[`VerifyFlag`].(float64)
		un, _ := v[`UserName`].(string)

		if vf/8 != 0 {
			v[`Type`] = ContactTypeOfficial
		} else if strings.HasPrefix(un, `@@`) {
			v[`Type`] = ContactTypeGroup
			gs = append(gs, un)
		} else {
			v[`Type`] = ContactTypeFriend
		}

		wechat.saveContactToCache(v)
	}

	for _, g := range gs {
		wechat.FourceUpdateGroup(g)
	}

	return nil
}

func (wechat *WeChat) fetchGroups(usernames []string) ([]map[string]interface{}, error) {

	var list []map[string]string
	for _, u := range usernames {
		list = append(list, map[string]string{
			`UserName`:   u,
			`ChatRoomId`: ``,
		})
	}

	data, err := json.Marshal(map[string]interface{}{
		`BaseRequest`: wechat.BaseRequest,
		`Count`:       len(list),
		`List`:        list,
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(`%s/webwxbatchgetcontact?type=ex&r=%v`, wechat.BaseURL, time.Now().Unix()*1000)
	resp := new(batchGetContactResponse)

	wechat.Excute(url, bytes.NewReader(data), resp)

	if !resp.IsSuccess() {
		logger.Error(resp.Error())
	}

	return resp.ContactList, nil
}

func (wechat *WeChat) fetchGroupsMembers(groups []map[string]interface{}) ([]map[string]interface{}, error) {

	list := make([]map[string]string, 0)

	for _, group := range groups {

		encryChatRoomID, _ := group[`EncryChatRoomId`].(string)
		members, _ := group[`MemberList`].([]interface{})

		logger.Debugf(`members %v`, members)

		for _, m := range members {
			mmap, _ := m.(map[string]interface{})
			u, _ := mmap[`UserName`].(string)
			list = append(list, map[string]string{
				`UserName`:        u,
				`EncryChatRoomId`: encryChatRoomID,
			})
		}
	}

	return wechat.fetchMembers(list), nil
}

func (wechat *WeChat) fetchMembers(list []map[string]string) []map[string]interface{} {

	if len(list) > maxCountOnceLoadGroupMember {
		return append(wechat.fetchMembers(list[:maxCountOnceLoadGroupMember]), wechat.fetchMembers(list[maxCountOnceLoadGroupMember:len(list)])...)
	}

	data, _ := json.Marshal(map[string]interface{}{
		`BaseRequest`: wechat.BaseRequest,
		`Count`:       len(list),
		`List`:        list,
	})

	url := fmt.Sprintf(`%s/webwxbatchgetcontact?type=ex&r=%v`, wechat.BaseURL, time.Now().Unix()*1000)
	resp := new(batchGetContactResponse)

	wechat.Excute(url, bytes.NewReader(data), resp)

	if !resp.IsSuccess() {
		logger.Error(resp.Error())
	}

	return resp.ContactList
}

// UpateGroupIfNeeded ...
func (wechat *WeChat) UpateGroupIfNeeded(groupID string) {

	bs, _ := wechat.contactCache.Get(groupID)

	if bs == nil {
		wechat.FourceUpdateGroup(groupID)
	}
}

// FourceUpdateGroup upate group infomation
func (wechat *WeChat) FourceUpdateGroup(groupID string) {

	groups, err := wechat.fetchGroups([]string{groupID})
	if err != nil || len(groups) != 1 {
		logger.Error(`sync group failed`)
		return
	}

	// 保存群组
	for _, v := range groups {
		v[`Type`] = ContactTypeGroup
		wechat.saveContactToCache(v)
	}

	memberList, err := wechat.fetchGroupsMembers(groups)
	if err != nil {
		logger.Error(`sync group failed`)
		return
	}

	for _, v := range memberList {
		v[`Type`] = ContactTypeFriend
		wechat.saveContactToCache(v)
	}
}

// ContactByUserName ...
func (wechat *WeChat) ContactByUserName(un string) (*Contact, error) {
	bs, err := wechat.contactCache.Get(un)
	if err != nil {
		return nil, err
	}
	var contact *Contact
	err = json.NewDecoder(bytes.NewReader(bs)).Decode(&contact)
	if err != nil {
		return nil, err
	}
	return contact, nil
}

// UserNameByNickName ..
func (wechat *WeChat) UserNameByNickName(nn string) (string, error) {

	bs, err := wechat.nicknameCache.Get(nn)
	if err != nil {
		return ``, nil
	}
	return string(bs), nil
}

// ContactByNickName search contact with nick name
func (wechat *WeChat) ContactByNickName(nn string) (*Contact, error) {
	un, err := wechat.UserNameByNickName(nn)
	if err != nil {
		return nil, err
	}
	return wechat.ContactByUserName(un)
}

func (wechat *WeChat) saveContactToCache(contact map[string]interface{}) {
	un, _ := contact[`UserName`].(string)
	bs, e := json.Marshal(contact)
	if e != nil {
		logger.Error(e)
	} else {
		wechat.contactCache.Set(un, bs)
		nn, _ := contact[`NickName`].(string)
		wechat.nicknameCache.Set(nn, []byte(un))
	}
}

func (wechat *WeChat) resetCache() {

	wechat.contactCache.Reset()
	wechat.nicknameCache.Reset()
}
