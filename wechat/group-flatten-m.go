package wechat

import "strings"

// Flatten ...
type flatten struct {
	wx *WeChat
}

// NewFlatten ...
func newFlatten(wechat *WeChat) *flatten {
	return &flatten{wechat}
}

// MapMsgs ...
func (f *flatten) MapMsgs(msg *CountedContent) {
	for _, m := range msg.Content {

		fromID, _ := m[`FromUserName`].(string)

		m[`IsSendByMySelf`] = fromID == f.wx.MySelf.UserName // TODO
		isGroup := false

		toID := m[`ToUserName`].(string)

		if strings.HasPrefix(fromID, `@@`) || strings.HasPrefix(toID, `@@`) { // TODO
			isGroup = true
		}

		m[`IsGroupMsg`] = isGroup
		if !isGroup {
			continue
		}

		if strings.HasPrefix(fromID, `@@`) {
			f.wx.UpateGroupIfNeeded(fromID)
		}

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
func (f *flatten) HandleMsgs(msg *CountedContent) {}
