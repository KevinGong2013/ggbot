package wechat

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/KevinGong2013/ggbot/utils"
	"github.com/satori/go.uuid"
)

type cache contactCache

type contactCache struct {
	sync.Mutex
	rootPath string
	ggmap    map[string]*Contact
	nickGG   map[string][]string
	userGG   map[string]string
}

// Contact is wx Account struct
type Contact struct {
	GGID            string
	UserName        string
	NickName        string
	HeadImgURL      string `json:"HeadImgUrl"`
	HeadHash        string
	RemarkName      string
	DisplayName     string
	StarFriend      float64
	Sex             float64
	Signature       string
	VerifyFlag      float64
	ContactFlag     float64
	HeadImgFlag     float64
	Province        string
	City            string
	Alias           string
	EncryChatRoomID string `json:"EncryChatRoomId"`
	Type            int
	MemberList      []*Contact
}

const (
	// Offical 公众号 ...
	Offical = 0
	// Friend 好友 ...
	Friend = 1
	// Group 群组 ...
	Group = 2
	// Member 群组成员 ...
	Member = 3
	// FriendAndMember 即是好友也是群成员 ...
	FriendAndMember = 4
)

func newCache(path string) *cache {
	return &cache{
		rootPath: path,
		ggmap:    make(map[string]*Contact),
		nickGG:   make(map[string][]string),
	}
}

func removeDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

func (c *cache) updateContact(contact *Contact) {
	c.ggmap[contact.GGID] = contact
	ggids := append(c.nickGG[contact.NickName], contact.GGID)
	removeDuplicates(&ggids)
	c.nickGG[contact.NickName] = ggids
	c.userGG[contact.UserName] = contact.GGID
}

func (wechat *WeChat) syncContacts(cts []map[string]interface{}) {
	wechat.cache.Lock()
	defer wechat.cache.Unlock()

	count := len(cts)

	logger.Debugf(`一共需要处理 [%d] 个联系人`, count)
	if count > 300 {
		logger.Warn(`您的联系人较多，可能需要等待1分钟左右`) // TODO 用多线程比较
	}
	// 有以下几种情况需要处理
	// 1. 内存和文件系统中都不存在联系人 ==> 直接新的cts数据初始化内存然后写文件
	// 2. 内存中没有联系人， 文件系统中有联系人 ==> 使用新数据和文件系统对比填充内存，然后写文件
	// 3. 内存和文件系统中都有联系人 ==> 以内存中的数据为主，更新数据，然后写文件
	//
	c := wechat.cache
	logger.Debug(`准备开始处理联系人信息`)
	c.userGG = make(map[string]string)

	if len(c.ggmap) == 0 {
		ggmap, nickGG := c.load()
		if ggmap != nil && nickGG != nil {
			c.ggmap = ggmap
			c.nickGG = nickGG
		}
	}

	if len(c.ggmap) == 0 { // 第一次启动最简单，直接刷进去
		logger.Debug(`联系人没有本地缓存，为每一个用户生成唯一ID`)
		for _, v := range cts {
			var nc *Contact
			bs, _ := json.Marshal(v)
			json.NewDecoder(bytes.NewReader(bs)).Decode(&nc)

			nc.GGID = uuid.NewV4().String()
			nc.HeadHash = contactHeadImgHash(wechat, nc) // 这里可能会比较耗时，但是是必须的

			c.updateContact(nc)
		}
	} else {
		logger.Debug(`发现联系人本地缓存，执行diff逻辑`)

		tempNickGG := c.nickGG

		c.nickGG = make(map[string][]string)
		var badguys []map[string]interface{}

		for _, v := range cts {
			nc, _ := newContact(v)

			// 唯一性判断
			ggids := tempNickGG[nc.NickName]

			if len(ggids) == 0 { // 由于改名，找不到这个人待处理
				logger.Warnf(`新添加或者离线时修改过昵称的联系人 [%s]`, nc.NickName)
				badguys = append(badguys, v)
			} else if len(ggids) == 1 { // 找到了1个id，对比其他信息
				oc := c.ggmap[ggids[0]]
				nc.GGID = oc.GGID
				nc.HeadHash = contactHeadImgHash(wechat, nc)
				if nc.HeadHash != oc.HeadHash {
					logger.Warnf(`我们认为[%s]修改了他的头像，但是也有可能是有2个人同时修改了昵称,
						请仔细检查,如若有误,请手动更改cache文件中的mapping 关系 GGID: %s`, nc.NickName, nc.GGID)
				}
				c.updateContact(nc)
				delete(tempNickGG, nc.NickName)
			} else { // 找到多个gid 有人改名字了
				gid := -1
				for idx, id := range ggids {
					oc := c.ggmap[id]
					// 这里认为找到唯一id的名字了
					if oc.HeadHash == contactHeadImgHash(wechat, nc) {
						logger.Infof(`已经处理同名联系人: %s`, nc.NickName)
						nc.GGID = oc.GGID
						nc.HeadHash = oc.HeadHash
						c.updateContact(nc)
						gid = idx
						break
					}
				}
				if gid != -1 {
					tempNickGG[nc.NickName] = append(ggids[:gid], ggids[gid+1:]...)
				} else { // 实在区分不出来的用户
					badguys = append(badguys, v)
				}
			}
		}

		// 这里的每个字典都很小，所以这么操作没有问题
		for _, b := range badguys {

			nc, _ := newContact(b)

			needRemoveNick := ``
			needRemoveidx := -1

			for nick, ids := range tempNickGG {
				for idx, id := range ids {
					oc := c.ggmap[id]
					if oc.HeadHash == contactHeadImgHash(wechat, nc) {
						needRemoveNick = nick
						needRemoveidx = idx
						break
					}
				}
				if len(needRemoveNick) > 0 {
					break
				}
			}

			if len(needRemoveNick) > 0 {

				i := needRemoveidx
				ggids := tempNickGG[needRemoveNick]
				oc := c.ggmap[ggids[i]]

				logger.Warnf(`我们认为[%s]将昵称改为[%s] GGID:%s`, oc.NickName, nc.NickName, oc.GGID)

				nc.GGID = oc.GGID
				nc.HeadHash = oc.HeadHash

				tempNickGG[oc.NickName] = append(ggids[:i], ggids[i+1:]...)
			} else {
				logger.Warnf(`无法确认用户id 作为新用户处理 nickName: [%s]`, nc.NickName)

				nc.GGID = uuid.NewV4().String()
				nc.HeadHash = contactHeadImgHash(wechat, nc)
			}

			c.updateContact(nc)
		}

		lostUser := make(map[string][]string)
		if len(tempNickGG) != 0 {
			for nick, ggids := range tempNickGG {
				if len(ggids) != 0 {
					lostUser[nick] = ggids
				}
			}
		}
		if len(lostUser) != 0 {
			logger.Warn(`丢失了以下用户 so sorry ~`)
			for nick, ggids := range lostUser {
				logger.Warnf(`NickName:%s GGIDs: %s`, nick, ggids)
			}
		}
	}

	for _, contact := range c.ggmap {
		if contact.Type == Group {
			for _, m := range contact.MemberList {
				gid, _ := c.userGG[m.UserName] // 为所有群里的成员添加GGID
				m.GGID = gid
			}
		}
	}

	//持久化到文件
	c.writeToFile()
}

func (wechat *WeChat) appendContacts(cts []map[string]interface{}) {

	wechat.cache.Lock()
	defer wechat.cache.Unlock()

	c := wechat.cache
	for _, v := range cts {
		nc, _ := newContact(v)
		// 看下系统中有没有
		if ggid, found := c.userGG[nc.UserName]; found {
			// 系统中已经存在了
			oc := c.ggmap[ggid]
			nc.GGID = oc.GGID
			nc.HeadHash = oc.HeadHash
			c.updateContact(nc)
		} else {
			// 新建
			nc.GGID = uuid.NewV4().String()
			nc.HeadHash = contactHeadImgHash(wechat, nc)
			c.updateContact(nc)
		}
	}

	for _, contact := range c.ggmap {
		if contact.Type == Group {
			for _, m := range contact.MemberList {
				gid, _ := c.userGG[m.UserName] // 为所有群里的成员添加GGID
				m.GGID = gid
			}
		}
	}

	wechat.cache.writeToFile()
}

func (c *cache) contactByGGID(ggid string) (*Contact, error) {
	if contact, found := c.ggmap[ggid]; found {
		return contact, nil
	}
	return nil, errors.New(`Not Found`)
}

func (c *cache) ggidsByNickName(nn string) ([]string, error) {
	if ggids, found := c.nickGG[nn]; found {
		return ggids, nil
	}
	return nil, errors.New(`Not Found`)
}

func (c *cache) contactByUserName(nn string) (*Contact, error) {
	if ggid, found := c.userGG[nn]; found {
		return c.contactByGGID(ggid)
	}
	return nil, errors.New(`Not Found`)
}

func (c *cache) writeToFile() error {

	buf1, _ := json.Marshal(c.ggmap)
	buf2, _ := json.Marshal(c.nickGG)

	path := c.rootPath + `/contact-cache-`
	err := utils.CreateFile(path+`1.json`, buf1, false)
	if err != nil {
		return err
	}
	return utils.CreateFile(path+`2.json`, buf2, false)
}

func (c *cache) load() (m1 map[string]*Contact, m2 map[string][]string) {

	path := c.rootPath + `/contact-cache-`

	_ = unmarshalLocalFile(path+`1.json`, &m1)
	_ = unmarshalLocalFile(path+`2.json`, &m2)

	return m1, m2
}

func unmarshalLocalFile(path string, obj interface{}) error {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, obj)
}

func contactHeadImgHash(wechat *WeChat, contact *Contact) string {

	data, err := wechat.GetContactHeadImg(contact)
	if err != nil {
		logger.Errorf(`can't get [%s]'s head image`, contact.NickName)
		return ``
	}

	hs := sha1.New()
	hs.Write(data)

	return fmt.Sprintf(`%x`, hs.Sum(nil))
}

func newContact(m map[string]interface{}) (*Contact, error) {

	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var c *Contact
	err = json.Unmarshal(data, &c)
	return c, err
}
