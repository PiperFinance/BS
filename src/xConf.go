package main

import (
	"os"

	"github.com/PiperFinance/BS/src/api"
	"github.com/PiperFinance/BS/src/api/views"
	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/PiperFinance/BS/src/core/tasks/handlers"
	"github.com/charmbracelet/log"
	"github.com/hibiken/asynq"
)

type StartConf struct{}

func (r *StartConf) xChainSchedule() []conf.QueueSchedules {
	// NOTE - Enqueuing Jobs via scheduler...
	return []conf.QueueSchedules{
		{Cron: conf.ETHBlocks, Key: tasks.BlockScanKey, Payload: nil, Q: asynq.Queue(conf.ScanQ)},
	}
}

func (r *StartConf) xHandlers() []conf.MuxHandler {
	return []conf.MuxHandler{
		{Key: tasks.BlockScanKey, Handler: handlers.BlockScanTaskHandler, Q: asynq.Queue(conf.ScanQ)},                    // 1
		{Key: tasks.FetchBlockEventsKey, Handler: handlers.BlockEventsTaskHandler, Q: asynq.Queue(conf.FetchQ)},          // 2
		{Key: tasks.ParseBlockEventsKey, Handler: handlers.ParseBlockEventsTaskHandler, Q: asynq.Queue(conf.ParseQ)},     // 3
		{Key: tasks.UpdateUserBalanceKey, Handler: handlers.UpdateUserBalTaskHandler, Q: asynq.Queue(conf.ProcessQ)},     // 4
		{Key: tasks.UpdateUserApproveKey, Handler: handlers.UpdateUserApproveTaskHandler, Q: asynq.Queue(conf.ProcessQ)}, // 4
		{Key: tasks.VacuumLogsKey, Handler: handlers.VacuumLogHandler, Q: asynq.Queue(conf.UnImportantQ)},                //~TBD
	}
}

func (r *StartConf) xUrls() []api.Route {
	return []api.Route{
		{Path: "/lsb", Method: api.Get, Handler: views.LastScannedBlock},
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

func (r *StartConf) StartApi() {
	go api.RunApi(":1300", r.xUrls())
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
	log.Info("Starting Api")
	r.StartApi()
}
