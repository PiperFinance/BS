package conf

import (
	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger

func LoadLogger() {
	var zp *zap.Logger
	var err error
	if Config.DEV {
		zp, err = zap.NewDevelopment(zap.IncreaseLevel(Config.ZapLogLevel))
	} else {
		zp, err = zap.NewProduction(zap.IncreaseLevel(Config.ZapLogLevel))
	}
	if err != nil {
		panic(err)
	}
	sugar := zp.Sugar()
	Logger = sugar
}
