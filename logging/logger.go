package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	conf := zap.NewDevelopmentConfig()
	conf.OutputPaths = []string{"stdout"}
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	conf.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	var err error
	if logger, err = conf.Build(); err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}
