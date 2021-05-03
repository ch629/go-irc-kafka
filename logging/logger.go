package logging

import (
	"go.uber.org/zap"
)

var logger zap.SugaredLogger

func init() {
	log, err := zap.NewDevelopment()

	if err != nil {
		panic(err)
	}

	logger = *log.Sugar()
}

func Logger() zap.SugaredLogger {
	return logger
}
