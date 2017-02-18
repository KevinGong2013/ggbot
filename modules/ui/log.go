package ui

var logs = make([]string, 0)

func (ui *UserInterface) Write(p []byte) (n int, err error) {

	log := string(p)
	logs = append(logs, log)
	if logList != nil {
		logList.AppendAtLast(log)
	}

	return len(p), err
}
