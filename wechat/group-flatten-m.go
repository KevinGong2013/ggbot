package wechat

import "strings"

// Flatten ...
type Flatten struct {
	wx *WeChat
}

// NewFlatten ...
func newFlatten(wechat *WeChat) *Flatten {
	return &Flatten{wechat}
}

// MapMsgs ...
func (f *Flatten) MapMsgs(msg *CountedContent) {
	for _, m := range msg.Content {

		fromID, _ := m[`FromUserName`].(string)

		m[`IsSendByMySelf`] = fromID == f.wx.MySelf.UserName // TODO
		m[`IsGroupMsg`] = true

		if !strings.HasPrefix(fromID, `@@`) { // TODO
			m[`IsGroupMsg`] = false
			continue
		}

		f.wx.UpateGroupIfNeeded(fromID)

		logger.Debugf(`will map group chat msg `)

		content, _ := m[`Content`].(string)

		atme := `@`
		if len(f.wx.MySelf.DisplayName) > 0 {
			atme += f.wx.MySelf.DisplayName
		} else {
			atme += f.wx.MySelf.NickName
		}
		m[`AtMe`] = strings.Contains(content, atme)

		infos := strings.Split(content, `:<br/>`)
		if len(infos) != 2 {
			continue
		}

		contact, err := f.wx.ContactByUserName(infos[0])
		if err != nil {
			logger.Error(err)
			f.wx.FourceUpdateGroup(fromID)
			continue
		}

		m[`ActualUserName`] = contact.UserName
		m[`ActualNickName`] = contact.NickName
		m[`Content`] = infos[1]
	}
}

// HandleMsgs ...
func (f *Flatten) HandleMsgs(msg *CountedContent) {}
