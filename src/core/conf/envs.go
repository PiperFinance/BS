package conf

import (
	"net/url"
	"time"

	"github.com/caarlos0/env/v8"
	"github.com/charmbracelet/log"
)

type config struct {
	MongoUrl       url.URL       `env:"MONGO_URL"`
	MongoDBName    string        `env:"MONGO_DBNAME"`
	RedisUrl       url.URL       `env:"REDIS_URL"`
	RedisDB        int           `env:"REDIS_DB"`
	MaxConcurrency int           `env:"MAX_CONCURRENT_WORKER"`
	MaxTaskTimeOut time.Duration `env:"MaxTaskTimeOut" envDefault:"10m"`
	AsynqMonUrl    string        `env:"ASYNQ_MON_URL" envDefault:":7654"`
	ApiUrl         string        `env:"API_URL" envDefault:":1300"`

	StartingBlockNumber uint64 `env:"STARTING_BLOCK_NUMBER" envDefault:"3"`
	// RPCUrl              string `env:"RPC_URL"`
	RPCUrls string `env:"RPC_URL_MULTI"`
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
