package convenience

import (
	"sync"

	"github.com/KevinGong2013/ggbot/wechat"
	log "github.com/Sirupsen/logrus"
)

var logger = log.WithFields(log.Fields{
	"module": "convenience",
})

// Event ...
type Event struct {
	Path string
	From string
	Data interface{}
	Time int64
}

// MsgStream ...
type MsgStream struct {
	sync.RWMutex
	wg          sync.WaitGroup
	srcMap      map[string]chan Event
	stream      chan Event
	Handlers    map[string]func(e Event)
	sigStopLoop chan Event
	msgEvent    chan Event
	wx          *wechat.WeChat
}

//DefaultMsgStream ...
var DefaultMsgStream = NewMsgStream()

// NewMsgStream ...
func NewMsgStream() *MsgStream {
	ms := &MsgStream{
		srcMap:      make(map[string]chan Event),
		stream:      make(chan Event),
		Handlers:    make(map[string]func(e Event)),
		sigStopLoop: make(chan Event),
		msgEvent:    make(chan Event),
	}

	ms.init()

	return ms
}

func (ms *MsgStream) init() {
	ms.Merge(`linsten`, ms.sigStopLoop)
	ms.Merge(`msg`, ms.msgEvent)
	go func() {
		ms.wg.Wait()
		close(ms.stream)
	}()
}

// Listen ...
func Listen() {
	DefaultMsgStream.Listen()
}

// Unlisten ...
func Unlisten() {
	DefaultMsgStream.Unlisten()
}

// Listen to start handle event
func (ms *MsgStream) Listen() {
	for {
		e := <-ms.stream
		switch e.Path {
		case `/sig/unlisten`:
			return
		}
		go func(te Event) {
			ms.RLock()
			defer ms.RUnlock()

			if pattern := ms.match(te.Path); pattern != `` {
				ms.Handlers[pattern](te)
			}
		}(e)
	}
}

// Unlisten stop handle event
func (ms *MsgStream) Unlisten() {
	go func() {
		e := Event{
			Path: `/sig/unlisten`,
		}
		ms.sigStopLoop <- e
	}()
}

// Handle path use default ms
func Handle(path string, handler func(e Event)) {
	DefaultMsgStream.Handle(path, handler)
}

// Handle ...
// /msg/solo
// /msg/solo/GGBot
// /msg/group
// /msg/group/GGBot测试群
func (ms *MsgStream) Handle(path string, handler func(e Event)) {
	ms.Handlers[cleanPath(path)] = handler
}

// ResetHandlers can Remove all existing defined Handlers from the map
func (ms *MsgStream) ResetHandlers() {
	for path := range ms.Handlers {
		delete(ms.Handlers, path)
	}
}

func (ms *MsgStream) match(path string) string {
	return findMatch(ms.Handlers, path)
}

func findMatch(mux map[string]func(e Event), path string) string {
	n := -1
	pattern := ""
	for m := range mux {
		if !isPathMatch(m, path) {
			continue
		}
		if len(m) > n {
			pattern = m
			n = len(m)
		}
	}
	return pattern

}

func isPathMatch(pattern, path string) bool {
	if len(pattern) == 0 {
		return false
	}
	n := len(pattern)
	return len(path) >= n && path[0:n] == pattern
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return p
}

// Merge other event to msgstream
func (ms *MsgStream) Merge(name string, ec chan Event) {
	ms.Lock()
	defer ms.Unlock()

	ms.wg.Add(1)
	ms.srcMap[name] = ec

	go func(a chan Event) {
		for n := range a {
			n.From = name
			ms.stream <- n
		}
		ms.wg.Done()
	}(ec)
}
