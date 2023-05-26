package conf

import (
	"net/url"
	"time"

	"github.com/caarlos0/env/v8"
	"github.com/charmbracelet/log"
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
	UserBalUpdateTimeout time.Duration `env:"USER_BAL_UPDATE_TIMEOUT" envDefault:"50s"`
	ParseBlockTimeout    time.Duration `env:"PARSE_BLOCK_TIMEOUT" envDefault:"2m"`
	FetchBlockTimeout    time.Duration `env:"FETCH_BLOCK_TIMEOUT" envDefault:"5m"`
	TestTimeout          time.Duration `env:"TEST_RPC_CONNECTION_TIMEOUT" envDefault:"10s"`
}

var Config config

func LoadConfig() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("%+v", err)
	}
	Config = cfg

	log.Infof("%+v", Config)
}
