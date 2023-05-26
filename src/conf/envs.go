package conf

import (
	"fmt"
	"net/url"
	"time"

	"github.com/caarlos0/env/v8"
	"go.uber.org/zap/zapcore"
)

type config struct {
	MongoUrl             url.URL       `env:"MONGO_URL"`
	MongoDBName          string        `env:"MONGO_DBNAME"`
	RedisUrl             url.URL       `env:"REDIS_URL"`
	RedisDB              int           `env:"REDIS_DB"`
	MaxConcurrency       int           `env:"MAX_CONCURRENT_WORKER"`
	MaxTaskTimeOut       time.Duration `env:"MaxTaskTimeOut" envDefault:"10m"`
	AsynqMonUrl          string        `env:"ASYNQ_MON_URL" envDefault:":7654"`
	ApiUrl               string        `env:"API_URL" envDefault:":1300"`
	StartingBlockNumber  uint64        `env:"STARTING_BLOCK_NUMBER" envDefault:"3"`
	SupportedChains      []int64       `env:"SUPPORTED_CHAINS" envSeparator:","`
	ParseBlockTimeout    time.Duration `env:"PARSE_BLOCK_TIMEOUT" envDefault:"2m"`
	FetchBlockTimeout    time.Duration `env:"FETCH_BLOCK_TIMEOUT" envDefault:"5m"`
	UserBalUpdateTimeout time.Duration `env:"USER_BAL_UPDATE_TIMEOUT" envDefault:"5m"`
	TestTimeout          time.Duration `env:"TEST_RPC_CONNECTION_TIMEOUT" envDefault:"10s"`
	ScanTaskTimeout      time.Duration `env:"SCAN_TASK_TIMEOUT" envDefault:"25s"`
	LogLevel             string        `env:"LOG_LEVEL" envDefault:"warn"`
	DEV                  bool          `env:"DEV_DEBUG" envDefault:"true"`
	ZapLogLevel          zapcore.Level
}

var Config config

func LoadConfig() {
	// cfg := config{}
	if err := env.Parse(&Config); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
	if lvl, err := zapcore.ParseLevel(Config.LogLevel); err != nil {
		panic(fmt.Sprintf("%+v", err))
	} else {
		Config.ZapLogLevel = lvl
	}

	// log.Infof("%+v", Config)
}
