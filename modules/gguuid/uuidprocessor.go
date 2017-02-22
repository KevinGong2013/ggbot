package gguuid

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	qrcode "github.com/skip2/go-qrcode"
)

const (
	fg = "\033[48;5;2m  \033[0m"
	bg = "\033[48;5;7m  \033[0m"
)

var logger = log.WithFields(log.Fields{
	"module": "uuidProcessor",
})

// UUIDProcessor ...
type UUIDProcessor struct{}

// New a uuid processor
func New() *UUIDProcessor {
	return &UUIDProcessor{}
}

// ProcessUUID impolements UUIDProcessor interface
func (up *UUIDProcessor) ProcessUUID(uuid string) error {

	content := fmt.Sprintf(`https://login.weixin.qq.com/l/%s`, uuid)

	code, err := qrcode.New(content, qrcode.Low)
	if err != nil {
		return err
	}

	showQRCode(code)

	return nil
}

// UUIDDidConfirm impolements UUIDProcessor interface
func (up *UUIDProcessor) UUIDDidConfirm(err error) {
	if err != nil {
		logger.Errorf(`above QRCODE has been invalidated`)
	} else {
		logger.Info(`uuid did confirmed`)
	}
}

func showQRCode(code *qrcode.QRCode) {

	for ir, row := range code.Bitmap() {
		lr := len(row)
		if ir == 0 || ir == 1 || ir == 2 ||
			ir == lr-1 || ir == lr-2 || ir == lr-3 {
			continue
		}
		for ic, col := range row {
			lc := len(code.Bitmap())
			if ic == 0 || ic == 1 || ic == 2 ||
				ic == lc-1 || ic == lc-2 || ic == lc-3 {
				continue
			}
			if col {
				fmt.Print(fg)
			} else {
				fmt.Print(bg)
			}
		}
		fmt.Println()
	}
}
