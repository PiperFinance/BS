package main

import (
	"fmt"

	"github.com/PiperFinance/BS/src/conf"
	_ "github.com/joho/godotenv/autoload"
)

// init App startup configurations
func init() {
	fmt.Println("BOOT : Loading Configs ... ")
	conf.LoadConfig()
	fmt.Println("BOOT : Loading Debug Tools ... ")
	conf.LoadDebugItems()
	fmt.Println("BOOT : Loading Logger ... ")
	conf.LoadLogger()
	fmt.Println("BOOT : Loading Local Cache ... ")
	conf.LoadLocalCache()
	fmt.Println("BOOT : Loading Mongo ...")
	conf.LoadMongo()
	fmt.Println("BOOT : Loading Redis ...")
	conf.LoadRedis()
	fmt.Println("BOOT : Loading Tokens ...")
	conf.LoadTokens()
	fmt.Println("BOOT : Loading Mainnets ...")
	conf.LoadMainNets()
	fmt.Println("BOOT : Loading Networks ...")
	conf.LoadNetwork()
	fmt.Println("BOOT : Initializing project workspace ...")
	conf.LoadProjectInit()
	fmt.Println("BOOT : Loading Q ...")
	conf.LoadQueue()
}

// ONLY FOR TESTING PURPOSES ...

func main() {
	if conf.Config.IsLocal {
		(&StartConf{}).StartLocalConf()
	} else {
		(&StartConf{}).StartAll()
	}
	select {}
}
