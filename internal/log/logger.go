package log

import "go.uber.org/zap"

var Logger *zap.Logger

func InitLogger() {
	initLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	Logger = initLogger
}
