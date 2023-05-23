package conf

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type config struct {
	MongoUrl    string `mapstructure:"MONGO_URL"`
	MongoDBName string `mapstructure:"MONGO_DBNAME"`

	RedisPort string `mapstructure:"REDIS_PORT"`
	RedisHost string `mapstructure:"REDIS_HOST"`
	RedisDB   int64  `mapstructure:"REDIS_DB"`

	MaxConcurrency int64         `mapstructure:"MAX_CONCURRENT_WORKER"`
	MaxTaskTimeOut time.Duration `mapstructure:"MaxTaskTimeOut" default:"10m"`
	AsynqMonUrl    string        `mapstructure:"ASYNQ_MON_URL" default:":7654"`
	ApiUrl         string        `mapstructure:"API_URL" default:":1300"`

	StartingBlockNumber uint64 `mapstructure:"STARTING_BLOCK_NUMBER" default:"3"`
	// RPCUrl              string `mapstructure:"RPC_URL"`
	RPCUrls string `mapstructure:"RPC_URL_MULTI"`
}

var Config config

func LoadConfig(path string) error {
	conf := config{}

	// viper.AddConfigPath(path)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.Unmarshal(&conf)
	if err != nil {
		log.Fatal(err)
	}

	Config = conf
	Config.MongoUrl = os.Getenv("MONGO_URL")
	Config.MongoDBName = os.Getenv("MONGO_DBNAME")
	Config.RedisHost = os.Getenv("REDIS_HOST")
	Config.RedisPort = os.Getenv("REDIS_PORT")
	Config.RedisDB, _ = strconv.ParseInt(os.Getenv("REDIS_DB"), 10, 64)
	Config.StartingBlockNumber, _ = strconv.ParseUint(os.Getenv("STARTING_BLOCK_NUMBER"), 10, 64)
	Config.AsynqMonUrl = os.Getenv("ASYNQ_MON_URL")
	Config.ApiUrl = os.Getenv("API_URL")
	Config.RPCUrls = os.Getenv("RPC_URL_MULTI")

	fmt.Printf("redis[ r://%s:%s/%d] \n ", Config.RedisHost, conf.RedisPort, conf.RedisDB)
	fmt.Printf("mongo[ m://%s/%s ] \n", Config.MongoUrl, conf.MongoDBName)
	fmt.Printf("endpoints[ mon://%s  api://%s ] \n ", Config.AsynqMonUrl, conf.ApiUrl)
	return err
}
