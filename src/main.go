package main

import (
	"github.com/PiperFinance/BS/src/core/conf"
	_ "github.com/joho/godotenv/autoload"
)

type BlockTask struct {
	BlockNumber uint64
}

func init() {
	conf.LoadConfig("./")
	conf.LoadNetwork()
	conf.LoadQueue()
	conf.LoadMongo()
	conf.LoadRedis()

	conf.LoadDebugItems()
}

// ONLY FOR TESTING PURPOSES ...
func main() {
	(&StartConf{}).StartAll()
	select {}
}
