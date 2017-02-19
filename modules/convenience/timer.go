package convenience

import (
	"strconv"
	"strings"
	"time"
)

// TimerEventData ...
type TimerEventData struct {
	Duration time.Duration
	Count    uint64
}

// TimingEventData ...
type TimingEventData struct {
	Count uint64
}

// NewTimerCh ...
func NewTimerCh(du time.Duration) chan Event {
	t := make(chan Event)

	go func(a chan Event) {
		n := uint64(0)
		for {
			n++
			time.Sleep(du)
			e := Event{}
			e.Path = "/timer/" + du.String()
			e.Time = time.Now().Unix()
			e.Data = TimerEventData{
				Duration: du,
				Count:    n,
			}
			t <- e

		}
	}(t)
	return t
}

// AddTimer ...
func AddTimer(du time.Duration) {
	DefaultMsgStream.Merge(`timer`, NewTimerCh(du))
}

// NewTimingCh ...
func NewTimingCh(hm string) chan Event {

	infos := strings.Split(hm, `:`)
	if len(infos) != 2 {
		panic(`hm incorrect`)
	}
	hour, _ := strconv.Atoi(infos[0])
	minute, _ := strconv.Atoi(infos[1])

	t := make(chan Event)

	go func(a chan Event) {
		n := uint64(0)
		for {
			now := time.Now()
			nh, nm, _ := now.Clock()
			next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
			if n > 0 || hour > nh || (hour == nh && minute < nm) {
				next = next.Add(time.Hour * 24)
			}
			logger.Debugf(`next timing %v`, next)
			n++
			time.Sleep(next.Sub(now))
			e := Event{}
			e.Path = `/timing/` + hm
			e.Time = time.Now().Unix()
			e.Data = TimerEventData{
				Count: n,
			}
			t <- e
		}
	}(t)
	return t
}

// AddTiming ...
func AddTiming(hm string) {
	DefaultMsgStream.Merge(`timing`, NewTimingCh(hm))
}
