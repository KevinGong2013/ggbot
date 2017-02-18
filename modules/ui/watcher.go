package ui

import "github.com/fsnotify/fsnotify"

func (ui *UserInterface) beginWatcher() {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				switch event.Op {
				case fsnotify.Create:
					if mediaList != nil { // TODO
						mediaList.AppendAtLast(event.Name)
					}
				}
			case err = <-watcher.Errors:
				logger.Error(err)
			}
		}
	}()

	err = watcher.Add(ui.mediaDir)
	if err != nil {
		logger.Error(err)
	}
	<-done
}
