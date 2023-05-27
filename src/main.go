package main

import (
	"github.com/PiperFinance/BS/src/conf"
	_ "github.com/joho/godotenv/autoload"
)

func init() {
	conf.LoadConfig()
	conf.LoadLogger()
	conf.LoadMongo()
	conf.LoadRedis()
	conf.LoadMainNets()
	conf.LoadNetwork()
	conf.LoadQueue()
	conf.LoadDebugItems()
}

// ONLY FOR TESTING PURPOSES ...

func main() {
	(&StartConf{}).StartAll()
	select {}
}
