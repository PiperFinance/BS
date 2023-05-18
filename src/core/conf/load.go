package conf

import (
	"log"

	"github.com/spf13/viper"
)

type config struct {
	MongoUrl    string `mapstructure:"MONGO_URL"`
	MongoDBName string `mapstructure:"MONGO_DBNAME"`

	RedisPort string `mapstructure:"REDIS_PORT"`
	RedisHost string `mapstructure:"REDIS_HOST"`
	RedisDB   int    `mapstructure:"REDIS_DB"`

	MaxConcurrency int `mapstructure:"MAX_CONCURRENT_WORKER"`

	StartingBlockNumber uint64 `mapstructure:"STARTING_BLOCK_NUMBER" default:"3"`
	RPCUrl              string `mapstructure:"RPC_URL"`
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
	return err
}
