package storage

import (
	log "github.com/Sirupsen/logrus"
	scribble "github.com/nanobox-io/golang-scribble"
)

// Storage ...
type Storage struct {
	db *scribble.Driver
}

var innerlogger = log.WithFields(log.Fields{
	"module": "storage",
})

// NewStorage ...
func NewStorage(dir string) (*Storage, error) {

	db, err := scribble.New(dir, &scribble.Options{Logger: new(logger)})
	if err != nil {
		return nil, err
	}

	return &Storage{db}, nil
}
