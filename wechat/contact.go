package wechat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"github.com/KevinGong2013/ggbot/utils"
)

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
	Seq         float64
}

type batchGetContactResponse struct {
	Response
	Count       int
	ContactList []map[string]interface{}
}

var maxCountOnceLoadGroupMember = 50

// To is contact's ID can be used in msg struct
func (contact *Contact) To() string {
	return contact.UserName
}

func (wechat *WeChat) getContacts(seq float64) ([]map[string]interface{}, float64, error) {

	url := fmt.Sprintf(`%s/webwxgetcontact?%s&%s&r=%s&seq=%v`, wechat.BaseURL, wechat.PassTicketKV(), wechat.SkeyKV(), utils.Now(), seq)
	resp := new(getContactResponse)

	err := wechat.Excute(url, nil, resp)

	if err != nil {
		return nil, 0, err
	}

	return resp.MemberList, resp.Seq, nil
}

// SyncContact with Wechat server.
func (wechat *WeChat) SyncContact() error {

	// 从头拉取通讯录
	seq := float64(-1)

	var cts []map[string]interface{}

	for seq != 0 {
		if seq == -1 {
			seq = 0
		}
		memberList, s, err := wechat.getContacts(seq)
		if err != nil {
			return err
		}
		seq = s
		cts = append(cts, memberList...)
	}

	var groupUserNames []string

	var tempIdxMap = make(map[string]int)

	for idx, v := range cts {

		vf, _ := v[`VerifyFlag`].(float64)
		un, _ := v[`UserName`].(string)

		if vf/8 != 0 {
			v[`Type`] = Offical
		} else if strings.HasPrefix(un, `@@`) {
			v[`Type`] = Group
			groupUserNames = append(groupUserNames, un)
		} else {
			v[`Type`] = Friend
		}
		tempIdxMap[un] = idx
	}

	groups, _ := wechat.fetchGroups(groupUserNames)

	for _, group := range groups {

		groupUserName := group[`UserName`].(string)
		contacts := group[`MemberList`].([]interface{})

		for _, c := range contacts {
			ct := c.(map[string]interface{})
			un := ct[`UserName`].(string)
			if idx, found := tempIdxMap[un]; found {
				cts[idx][`Type`] = FriendAndMember
			} else {
				ct[`HeadImgUrl`] = fmt.Sprintf(`/cgi-bin/mmwebwx-bin/webwxgeticon?seq=0&username=%s&chatroomid=%s&skey=`, un, groupUserName)
				ct[`Type`] = Member
				cts = append(cts, ct)
			}
		}

		group[`Type`] = Group
		idx := tempIdxMap[groupUserName]
		cts[idx] = group
	}

	wechat.syncContacts(cts)

	return nil
}

// GetContactHeadImg ...
func (wechat *WeChat) GetContactHeadImg(c *Contact) ([]byte, error) {

	urlOBJ, err := url.Parse(wechat.BaseURL)

	if err != nil {
		return nil, err
	}

	host := urlOBJ.Host

	url := fmt.Sprintf(`https://%s%s`, host, c.HeadImgURL)

	resp, err := wechat.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
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

	if _, err := wechat.cache.ggidsByNickName(groupID); err == nil {
		logger.Debug(`already has group infomation`)
		return
	}

	wechat.FourceUpdateGroup(groupID)
}

// FourceUpdateGroup upate group infomation
func (wechat *WeChat) FourceUpdateGroup(groupID string) {

	groups, err := wechat.fetchGroups([]string{groupID})
	if err != nil || len(groups) != 1 {
		logger.Error(`sync group failed`)
		return
	}

	group := groups[0]
	group[`Type`] = Group

	var cts []map[string]interface{}

	cts = append(cts, groups[0])

	memberList, err := wechat.fetchGroupsMembers(groups)
	if err != nil {
		logger.Error(`sync group failed`)
		return
	}

	for _, v := range memberList {
		if _, found := wechat.cache.userGG[v[`UserName`].(string)]; found {
			v[`Type`] = FriendAndMember
		} else {
			v[`Type`] = Group
		}
	}

	wechat.appendContacts(append(cts, memberList...))
}

// ContactByUserName ...
func (wechat *WeChat) ContactByUserName(un string) (*Contact, error) {

	ggid, found := wechat.cache.userGG[un]
	if !found {
		return nil, errors.New(`not found`)
	}

	return wechat.cache.contactByGGID(ggid)
}

// UserNameByNickName ..
func (wechat *WeChat) UserNameByNickName(nn string) ([]string, error) {

	cs, err := wechat.ContactByNickName(nn)
	if err != nil {
		return nil, err
	}

	var uns []string
	for _, c := range cs {
		uns = append(uns, c.UserName)
	}

	return uns, nil
}

// ContactByNickName search contact with nick name
func (wechat *WeChat) ContactByNickName(nn string) ([]*Contact, error) {
	ggids, found := wechat.cache.nickGG[nn]
	if !found {
		return nil, errors.New(`not found`)
	}
	var cs []*Contact
	for _, ggid := range ggids {
		c, err := wechat.cache.contactByGGID(ggid)
		if err == nil {
			cs = append(cs, c)
		}
	}
	if len(cs) > 0 {
		return cs, nil
	}
	return nil, errors.New(`not found`)
}

// ContactByGGID ...
func (wechat *WeChat) ContactByGGID(id string) (*Contact, error) {
	if c, found := wechat.cache.ggmap[id]; found {
		return c, nil
	}
	return nil, errors.New(`not found`)
}

// AllContacts ...
func (wechat *WeChat) AllContacts() []*Contact {
	var vs []*Contact
	for _, c := range wechat.cache.ggmap {
		vs = append(vs, c)
	}
	return vs
}

// TODO
func (wechat *WeChat) modifyRemarkName(un string) (string, error) {

	data, _ := json.Marshal(map[string]interface{}{
		`BaseRequest`: wechat.BaseRequest,
		`UserName`:    un,
		`CmdId`:       2,
		`NickName`:    `Test`,
	})

	url := fmt.Sprintf(`%s/webwxoplog?lang=zh_CN&%v`, wechat.BaseURL, wechat.PassTicketKV())
	resp := new(Response)

	wechat.Excute(url, bytes.NewReader(data), resp)

	if !resp.IsSuccess() {
		logger.Error(resp.Error())
	}

	return `Test`, nil
}

func (wechat *WeChat) contactDidChange(cts *CountedContent, changeType int) {
	if changeType == 0 { // 修改
		for _, v := range cts.Content {
			vf, _ := v[`VerifyFlag`].(float64)
			un, _ := v[`UserName`].(string)

			if vf/8 != 0 {
				v[`Type`] = Offical
			} else if strings.HasPrefix(un, `@@`) {
				v[`Type`] = Group
			} else {
				v[`Type`] = Friend
			}
		}
		wechat.appendContacts(cts.Content)
	}
}
