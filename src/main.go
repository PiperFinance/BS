package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/hibiken/asynq"
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
	// mainConf := StartConf{}
	// mainConf.StartAll()
	(&StartConf{}).StartAll()
	time.Sleep(5 * time.Second)

	// FIXME - Force Create Tasks Here ...
	payload, err := json.Marshal(BlockTask{BlockNumber: conf.Config.StartingBlockNumber})
	if err != nil {
		log.Fatal(err)
	}
	asynq.NewTask(tasks.BlockScanKey, payload)

	// fmt.Println(conf.QueueStatus())

	select {}
}
