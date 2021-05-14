package utils

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

var _instance *log.Logger
var once sync.Once

//GetLogger ... Gets the instance for the logger.
func GetLogger() *log.Logger {
	once.Do(func() {
		_instance = log.New()
	})
	return _instance
}
