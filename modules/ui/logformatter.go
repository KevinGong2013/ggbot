package ui

import (
	"fmt"

	"github.com/Sirupsen/logrus"
)

type logformatter struct{}

func (lf logformatter) Format(entry *logrus.Entry) ([]byte, error) {

	var msgHead string

	switch entry.Level {
	case logrus.DebugLevel:
		msgHead = `[DEBU](fg-cyan)`
	case logrus.InfoLevel:
		msgHead = `[INFO](fg-green)`
	case logrus.WarnLevel:
		msgHead = `[WARN](fg-yellow)`
	case logrus.ErrorLevel:
		msgHead = `[ERRO](fg-white,bg-red)`
	default:
		msgHead = `[UNKN](fg-white,bg-blue)`
	}

	return []byte(fmt.Sprintf(`%v [%v](fg-magenta) %v`, msgHead, entry.Data[`module`], entry.Message)), nil
}
