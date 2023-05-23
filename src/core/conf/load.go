package conf

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

type config struct {
	MongoUrl    string `mapstructure:"MONGO_URL"`
	MongoDBName string `mapstructure:"MONGO_DBNAME"`

	RedisPort string `mapstructure:"REDIS_PORT"`
	RedisHost string `mapstructure:"REDIS_HOST"`
	RedisDB   int    `mapstructure:"REDIS_DB"`

	MaxConcurrency int           `mapstructure:"MAX_CONCURRENT_WORKER"`
	MaxTaskTimeOut time.Duration `mapstructure:"MaxTaskTimeOut" default:"10m"`
	AsynqMonUrl    string        `mapstructure:"ASYNQ_MON_URL" default:":7654"`
	ApiUrl         string        `mapstructure:"API_URL" default:":1300"`

	StartingBlockNumber uint64 `mapstructure:"STARTING_BLOCK_NUMBER" default:"3"`
	RPCUrl              string `mapstructure:"RPC_URL"`
	RPCUrls             string `mapstructure:"RPC_URL_MULTI"`
}

var Config config

func LoadConfig(path string) error {
	conf := config{}

	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	// viper.SetConfigName("app")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = viper.Unmarshal(&conf)
	if err != nil {
		log.Fatal(err)
	}
	Config = conf
	fmt.Printf("redis[ r://%s:%s/%d] \n ", Config.RedisHost, conf.RedisPort, conf.RedisDB)
	fmt.Printf("mongo[ m://%s/%s ] \n", Config.MongoUrl, conf.MongoDBName)
	fmt.Printf("endpoints[ mon://%s  api://%s ] \n ", Config.AsynqMonUrl, conf.ApiUrl)
	return err
}
