package ui

import (
	log "github.com/Sirupsen/logrus"
	"github.com/KevinGong2013/ggbot/wechat"
	"github.com/gizak/termui"
)

// UserInterface ...
type UserInterface struct {
	mediaDir string
	wx       *wechat.WeChat
}

var logger = log.WithFields(log.Fields{
	"module": "ui",
})

var didInLoop = false
var didFocuseWidge Focuse

// 1
var titlePar *termui.Par

// 2
var versionTable *termui.Table

// 3
var liveMsgList *List
var mediaList *List

// 4
var msgSummaryChart *termui.BarChart

// 5
var featureList *List

// 6
var logList *List

// 7
var tipPar *termui.Par

var indicatorPar *termui.Par

// NewUI ...
func NewUI(dir string) *UserInterface {

	u := &UserInterface{dir, nil}

	// 代理log的输出源
	log.SetOutput(u)
	log.SetFormatter(logformatter{})

	return u
}

func buildUI() {

	// 1
	titlePar = termui.NewPar(`[Welcome to GGBOT ~](fg-green)
							  [取消](fg-red)选择请按[Esc](fg-red)
							  [退出](fg-red)请按[q](fg-red)`)
	titlePar.Height = 5

	indicatorPar = termui.NewPar(`
		◉◉◉
		`)
	indicatorPar.Height = 5
	indicatorPar.BorderLabel = `登录态`

	// 2
	rows := [][]string{
		[]string{"Date", "Version"},
		[]string{"2017-01-20", "v0.0.9-beta.1"},
		[]string{"2017-03-20", "v0.0.9-beta.2"},
	}

	versionTable := termui.NewTable()
	versionTable.Rows = rows
	versionTable.FgColor = termui.ColorWhite
	versionTable.BgColor = termui.ColorDefault
	versionTable.Separator = false
	versionTable.Height = 5
	versionTable.Analysis()
	versionTable.BgColors[2] = termui.ColorGreen

	//  3
	liveMsgList = NewList()
	liveMsgList.BorderLabel = `实时消息`
	liveMsgList.Height = 10

	mediaList = NewList()
	mediaList.BorderLabel = `媒体文件`
	mediaList.Height = 10

	// 4
	msgSummaryChart := termui.NewBarChart()
	msgSummaryChart.BorderLabelFg = termui.ColorRed
	msgSummaryChart.Data = []int{30, 2, 5, 23, 9, 5, 30, 30}
	msgSummaryChart.BarWidth = 4
	msgSummaryChart.Height = 10
	msgSummaryChart.DataLabels = []string{"文本", "语音", "图片", "红包", "连接", "分享", "其他", "心跳"}
	msgSummaryChart.TextColor = termui.ColorWhite
	msgSummaryChart.BarColor = termui.ColorGreen
	msgSummaryChart.NumColor = termui.ColorWhite

	// 5
	featureList = NewList()
	featureList.BorderLabel = `功能区`
	featureList.Height = 10
	featureList.Items = []string{
		`[0] TODO`,
	}

	// 6
	logList = NewList()
	logList.BorderLabel = `日志`
	logList.Height = 10
	logList.Items = logs

	// 7
	tipPar = termui.NewPar(``)
	tipPar.Height = 3
	tipPar.Border = false

	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(6, 0, titlePar),
			termui.NewCol(2, 0, indicatorPar),
			termui.NewCol(4, 0, versionTable),
		),
		termui.NewRow(
			termui.NewCol(12, 0, tipPar),
		),
		termui.NewRow(
			termui.NewCol(6, 0, featureList),
			termui.NewCol(6, 0, mediaList),
		),
		termui.NewRow(
			termui.NewCol(8, 0, liveMsgList),
			termui.NewCol(4, 0, msgSummaryChart),
		),
		termui.NewRow(
			termui.NewCol(12, 0, logList),
		),
	)
}

func registerEvent() {
	// 1
	termui.Handle(`/sys/kbd/q`, func(e termui.Event) {
		if didFocuseWidge != nil {
			tipPar.Text = `请先按[Esc](fg-red)取消选择在退出`
			termui.Render(tipPar)
			return
		}
		stopLoop()
	})

	termui.Handle(`/sys/kbd`, func(arg2 termui.Event) {

		key := arg2.Data.(termui.EvtKbd).KeyStr

		if key == `<escape>` {
			setFocuseWidge(nil)
			return
		}
		if key == `<f1>` {
			setFocuseWidge(liveMsgList)
		}
		if key == `<f2>` {
			setFocuseWidge(mediaList)
		}
		if key == `<f3>` {
			setFocuseWidge(featureList)
		}
		if key == `<f4>` {
			setFocuseWidge(logList)
		}

		if didFocuseWidge == nil {
			tipPar.Text = `请先用 F1~F4 选择功能再操作`
			termui.Render(tipPar)
		}
	})

	termui.Handle("/timer/1s", func(e termui.Event) {

		t := e.Data.(termui.EvtTimer)

		if t.Count%3 == 0 {
			indicatorPar.Text = (`
				[◉](fg-red)[◉](fg-green)[◉](fg-yellow)
				`)
		} else if t.Count%2 == 0 {
			indicatorPar.Text = (`
				[◉](fg-green)[◉](fg-yellow)[◉](fg-red)
				`)
		} else {
			indicatorPar.Text = (`
				[◉](fg-yellow)[◉](fg-red)[◉](fg-green)
				`)
		}
		termui.Render(indicatorPar)
	})

	termui.Handle(`/sys/wnd/resize`, func(arg2 termui.Event) {
		termui.Body.Width = termui.TermWidth()
		termui.Body.Align()
		termui.Clear()
		termui.Render(termui.Body)
	})

	//
}

// Loop prepase cui
func (ui *UserInterface) Loop() {

	if didInLoop {
		return
	}

	err := termui.Init()
	defer termui.Close()
	if err != nil {
		logger.Error(err)
		return
	}

	if tipPar == nil {
		buildUI()
		go ui.beginWatcher()
	}

	termui.DefaultEvtStream.ResetHandlers()
	registerEvent()

	termui.Render(termui.Body)
	termui.Body.Align()
	termui.Loop()

}

// StopLoop ...
func stopLoop() {
	termui.StopLoop()
	didInLoop = false
}

func setFocuseWidge(f Focuse) {
	if didFocuseWidge != nil {
		didFocuseWidge.Unfocused()
		didFocuseWidge = nil
	}
	if f != nil {
		f.Focused()
		didFocuseWidge = f
	}
}
