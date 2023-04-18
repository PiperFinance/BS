package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/hibiken/asynq"
	_ "github.com/joho/godotenv/autoload"
)

type BlockTask struct {
	BlockNumber uint64
}

// ONLY FOR TESTING PURPOSES ...
func main() {
	handlers := []conf.MuxHandler{
		{Key: tasks.ParseBlockEventsKey, Handler: tasks.ParseBlockEventsTaskHandler},
		{Key: tasks.FetchBlockEventsKey, Handler: tasks.BlockEventsTaskHandler},
		{Key: tasks.BlockScanKey, Handler: tasks.BlockScanTaskHandler},
	}
	//schedules := []conf.QueueSchedules{
	//	{"@every 50s", tasks.BlockScanKey, nil},
	//}
	//schedules := []conf.QueueSchedules{
	//	{"@every 1s", tasks.BlockScanKey, nil},
	//}
	payload, err := json.Marshal(BlockTask{BlockNumber: 16978252})
	if err != nil {
		log.Fatal(err)
	}
	asynq.NewTask("block:parse_events", payload)

	go conf.RunClient()
	go conf.RunWorker(handlers)
	//go conf.RunScheduler(schedules)
	fmt.Println(conf.QueueStatus())

	select {}

}
