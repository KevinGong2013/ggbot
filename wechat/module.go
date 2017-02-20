package wechat

// LoginStateModule ..
type LoginStateModule interface {
	WechatDidLogin(wechat *WeChat)
	WechatDidLogout(wechat *WeChat)
}

// MsgModule ...
type MsgModule interface {
	MapMsgs(msg *CountedContent)
	HandleMsgs(msg *CountedContent)
}

// ContactModule ...
type ContactModule interface {
	MapContact(contact map[string]interface{})
	HandleContact(contact map[string]interface{})
}

var wxInstance *WeChat

var registedMsgModules []MsgModule
var registedContactModules []ContactModule
var registedLoginStateModules []LoginStateModule
var holdUselessModules []interface{}

var addMsg = make(chan *CountedContent)
var modContact = make(chan *CountedContent)
var delContact = make(chan *CountedContent)
var modChatRoomMember = make(chan *CountedContent)

func init() {
	go func() {
		for {
			state := <-loginState
			wxInstance.IsLogin = state == 1
			for _, m := range registedLoginStateModules {
				if wxInstance.IsLogin {
					m.WechatDidLogin(wxInstance)
				} else {
					m.WechatDidLogout(wxInstance)
				}
			}
		}
	}()
}

// RegisterModule desgin for handle server event
func (wechat *WeChat) RegisterModule(m interface{}) {

	wxInstance = wechat

	useless := true

	if lm, ok := m.(LoginStateModule); ok {
		registedLoginStateModules = append(registedLoginStateModules, lm)
		useless = false
	}
	if mm, ok := m.(MsgModule); ok {
		registedMsgModules = append(registedMsgModules, mm)
		useless = false
	}
	if cm, ok := m.(ContactModule); ok {
		registedContactModules = append(registedContactModules, cm)
		useless = false
	}

	if useless {
		holdUselessModules = append(holdUselessModules, m)
	}
}

func (wechat *WeChat) handleServerEvent() {
	go func() {
		for {
			select {
			case msg := <-addMsg:
				go wechat.handleMsgs(msg)
			case modyfyContact := <-modContact:
				go wechat.handleContacts(modyfyContact, 0)
			case delContact := <-delContact:
				go wechat.handleContacts(delContact, 1)
			case modChatRoomMember := <-modChatRoomMember:
				go wechat.handleContacts(modChatRoomMember, 2)
			}
		}
	}()
}

func (wechat *WeChat) handleMsgs(msg *CountedContent) {
	for _, m := range registedMsgModules {
		m.MapMsgs(msg)
	}

	for _, m := range registedMsgModules {
		go m.HandleMsgs(msg)
	}
}

func (wechat *WeChat) handleContacts(cts *CountedContent, changeType int) {

	wechat.contactDidChange(cts, changeType)

	for _, v := range cts.Content {
		v[`ChangeType`] = changeType
		for _, m := range registedContactModules {
			m.MapContact(v)
		}
	}

	for _, m := range registedContactModules {
		for _, c := range cts.Content {
			go m.HandleContact(c)
		}
	}
}
