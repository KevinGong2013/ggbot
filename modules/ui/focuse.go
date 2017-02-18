package ui

// Focuse ...
type Focuse interface {
	Focused()
	Unfocused()
	IsFocused() bool
}
