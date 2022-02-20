package logger

import (
	"github.com/blendle/zapdriver"
	"go.uber.org/zap"
)

func NewLogger() *zap.SugaredLogger {
	logger, err := zapdriver.NewProduction()
	if err != nil {
		panic(err)
	}

	sugar := logger.Sugar()

	return sugar
}