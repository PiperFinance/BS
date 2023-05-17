package main

import (
	"os"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/charmbracelet/log"
)

type StartConf struct{}

func (r *StartConf) xChainSchedule() []conf.QueueSchedules {
	// NOTE - Enqueuing Jobs via scheduler...
	return []conf.QueueSchedules{
		{Cron: conf.ETHBlocks, Key: tasks.BlockScanKey, Payload: nil},
	}
}

func (r *StartConf) xHandlers() []conf.MuxHandler {
	return []conf.MuxHandler{
		{Key: tasks.BlockScanKey, Handler: tasks.BlockScanTaskHandler},               // 1
		{Key: tasks.FetchBlockEventsKey, Handler: tasks.BlockEventsTaskHandler},      // 2
		{Key: tasks.ParseBlockEventsKey, Handler: tasks.ParseBlockEventsTaskHandler}, // 3
		{Key: tasks.UpdateUserBalanceKey, Handler: tasks.UpdateUserBalTaskHandler},   // 4
		{Key: tasks.VacuumLogsKey, Handler: tasks.VacuumLogHandler},                  //~TBD
	}
}

func (r *StartConf) xMonPort() string {
	AsynqMonUrl, ok := os.LookupEnv("ASYNQ_MON_URL")
	if !ok {
		log.Warn("ASYNQ_MON_URL not Found! Setting Default Of :8765")
		AsynqMonUrl = ":7654"
	}
	return AsynqMonUrl
}

func (r *StartConf) StartClient() {
	go conf.RunClient()
}

func (r *StartConf) StartWorker() {
	go conf.RunWorker(r.xHandlers())
}

func (r *StartConf) StartScheduler() {
	go conf.RunScheduler(r.xChainSchedule())
}

func (r *StartConf) StartMon() {
	go conf.RunMonitor(r.xMonPort())
}

func (r *StartConf) StartAll() {
	log.Info("Starting Worker")
	r.StartWorker() // Consumer
	log.Info("Starting Client")
	r.StartClient() // Producer
	log.Info("Starting Scheduler")
	r.StartScheduler() // Scheduled Producer
	log.Info("Starting AsynQMon")
	r.StartMon() // asynqMon
}
