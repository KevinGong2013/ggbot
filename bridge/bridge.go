package bridge

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/KevinGong2013/ggbot/bridge/arg"
	r "github.com/KevinGong2013/ggbot/bridge/result"
	log "github.com/Sirupsen/logrus"
)

// Connector is design to connect iOS wechat app
type Connector interface {
	RefreshToken(token string)
	Send(a *arg.Arg) error
}

var logger = log.WithFields(log.Fields{
	"module": "bridge",
})

// Wrapper ...
type Wrapper struct {
	connector Connector

	mutex   sync.Mutex
	seq     uint64
	pending map[uint64]chan *r.Result
}

// NewWrapper ...
func NewWrapper(c Connector) *Wrapper {

	w := &Wrapper{
		connector: c,
		pending:   make(map[uint64]chan *r.Result),
	}

	//
	http.HandleFunc("/bridge", w.handle)

	go http.ListenAndServe(`:3280`, nil)

	return w
}

// Go ...
func (w *Wrapper) Go(a *arg.Arg, done chan *r.Result) {

	w.mutex.Lock()

	seq := w.seq
	w.seq++
	w.pending[seq] = done

	a.Seq = seq // 很重要

	w.mutex.Unlock()

	err := w.connector.Send(a)

	if err != nil {
		done <- r.NewResultWithError(err.Error())
		w.mutex.Lock()
		delete(w.pending, seq)
		w.mutex.Unlock()
	}
}

// Call ..
func (w *Wrapper) Call(a *arg.Arg) *r.Result {
	done := make(chan *r.Result, 1)
	w.Go(a, done)

	timer := time.NewTimer(time.Minute * 1).C

	select {
	case r := <-done:
		if r.IsFailure() {
			logger.Errorf(`send %v to wechat app failed error: %v`, a, r.Err)
		} else {
			logger.Infof(`send %v to wechat app successed`, a)
		}
		return r
	case <-timer:
		logger.Errorf(`send %v to wechat app time out`, a)
		return r.NewResultWithError(fmt.Sprintf(`time out %v`, a))
	}
}

func (w *Wrapper) handle(writer http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	logger.Debugf(`---cmd:%v------ did receive wechat app response`, req.Header[`Cmd`])

	seqs := req.Header[`Seq`]
	cmds := req.Header[`Cmd`]

	if len(seqs) == 0 || len(cmds) == 0 {
		logger.Errorf(`invalidate req %v`, req)
		return
	}

	seq, _ := strconv.ParseUint(seqs[0], 10, 32)
	cmd, _ := strconv.Atoi(cmds[0])

	var result map[string]interface{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&result)
	if err != nil {
		logger.Errorf(`decode error: %v`, err)
		return
	}

	logger.Debugf(`---seq:[%v]------ body: %v`, seq, result)

	if seq == 0 && cmd == arg.Token {
		t, _ := result[`token`].(string)
		w.connector.RefreshToken(t)
	} else if seq == 0 && cmd == arg.RedPacket {
		status, _ := result[`status`].(string)
		if status == `opened` {
			logger.Info(`wechat app did open a redpacket`)
		}
	} else {
		w.mutex.Lock()
		done := w.pending[seq]
		delete(w.pending, seq)
		w.mutex.Unlock()

		done <- r.NewResultWithMap(result)
	}
}
