package log

import (
	"www.baidu.com/golang-lib/log"
)

func Debug(i interface{}) {
	log.Logger.Debug(i)
}

func Info(i interface{}) {
	log.Logger.Info(i)
}

func Warn(i interface{}) {
	log.Logger.Warn(i)
}

func Error(i interface{}) {
	log.Logger.Error(i)
}
