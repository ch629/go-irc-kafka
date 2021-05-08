package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger zap.SugaredLogger

func init() {
	conf := zap.NewDevelopmentConfig()
	conf.OutputPaths = []string{"stdout"}
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	log, err := conf.Build()

	if err != nil {
		panic(err)
	}

	logger = *log.Sugar()
}

func Logger() zap.SugaredLogger {
	return logger
}
