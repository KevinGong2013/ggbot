package ui

import (
	"fmt"
	"sync"

	"github.com/gizak/termui"
)

const (
	kbdUp   = `<up>`
	kbdDown = `<down>`

	kbdRight = `<right>`
	kbdEnter = `<enter>`
)

// List ...
type List struct {
	termui.List
	focused     bool
	selectedIdx int
	mutex       sync.Mutex
}

// NewList ...
func NewList() *List {
	return &List{
		List:        *termui.NewList(),
		focused:     false,
		selectedIdx: -1}
}

func (l *List) hlightSelectedItem() {

	item := l.Items[l.selectedIdx]
	var str = ``
	for _, cell := range termui.DefaultTxBuilder.Build(item, 0, 0) {
		str += string(cell.Ch)
	}
	str = fmt.Sprintf(`[%s](fg-white,bg-green)`, str)
	l.Items[l.selectedIdx] = str
}

func (l *List) unhilightSelectedItem() {

	if l.selectedIdx != -1 {
		item := l.Items[l.selectedIdx]
		var str = ``
		for _, cell := range termui.DefaultTxBuilder.Build(item, 0, 0) {
			str += string(cell.Ch)
		}
		l.Items[l.selectedIdx] = str
	}
}

// Append item to list at idx.
func (l *List) Append(item string, at int) {

	l.unhilightSelectedItem()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	if at < len(l.Items) {
		l.Items[at] = item
	} else {
		l.Items = append(l.Items, item)
	}
	if len(l.Items) > l.Block.Height-2 {
		l.Items = l.Items[len(l.Items)-l.Block.Height+2:]
	}

	l.selectedIdx = -1

	termui.Render(l)
}

// AppendAtLast ....
func (l *List) AppendAtLast(item string) {
	l.Append(item, len(l.Items))
}

// Focused ...
func (l *List) Focused() {
	l.focused = true
	l.BorderFg = termui.ColorGreen
	termui.Render(l)
}

// Unfocused ...
func (l *List) Unfocused() {
	l.focused = false
	l.BorderFg = termui.ColorWhite
	termui.Render(l)
}

// IsFocused ...
func (l *List) IsFocused() bool {
	return l.focused
}
