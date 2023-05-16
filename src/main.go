package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/charmbracelet/log"
	"github.com/hibiken/asynq"
	_ "github.com/joho/godotenv/autoload"
)

type BlockTask struct {
	BlockNumber uint64
}

// ONLY FOR TESTING PURPOSES ...
func main() {
	// mainConf := StartConf{}
	// mainConf.StartAll()
	(&StartConf{}).StartAll()
	time.Sleep(5 * time.Second)

	// FIXME - Force Create Tasks Here ...
	payload, err := json.Marshal(BlockTask{BlockNumber: 16978252})
	if err != nil {
		log.Fatal(err)
	}
	asynq.NewTask(tasks.BlockScanKey, payload)

	fmt.Println(conf.QueueStatus())

	select {}
}
