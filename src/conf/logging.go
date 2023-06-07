package conf

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func LoadLogger() {
	var zp *zap.Logger
	var err error

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	errLog, _, err := zap.Open(fmt.Sprintf("%s%s", Config.LogDir, "/err.log"))
	if err != nil {
		panic(err)
	}
	debugLog, _, err := zap.Open(fmt.Sprintf("%s%s", Config.LogDir, "/debug.log"))
	if err != nil {
		panic(err)
	}
	topicDebugging := debugLog
	topicErrors := errLog

	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	fileEncoder := zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig())
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, topicErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(fileEncoder, topicDebugging, lowPriority),
		zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
	)

	zp = zap.New(core, zap.Development(), zap.AddStacktrace(zap.WarnLevel))
	defer zp.Sync()
	zp.Info("constructed a logger")

	// if Config.DEV {
	// 	zp, err = zap.NewDevelopment(zap.IncreaseLevel(Config.ZapLogLevel), zap.WrapCore())
	// } else {
	// 	zp, err = zap.NewProduction(zap.IncreaseLevel(Config.ZapLogLevel))
	// }
	// if err != nil {
	// 	panic(err)
	// }
	sugar := zp.Sugar()
	Logger = sugar
}
