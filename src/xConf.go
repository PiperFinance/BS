package main

import (
	"os"
	"time"

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
		{Cron: conf.ETHBlocks, Key: tasks.BlockScanKey, Payload: nil, Q: asynq.Queue(conf.ScanQ), Timeout: 25 * time.Second},
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
		{Path: "/lsb/100", Method: api.Get, Handler: views.LastScannedBlocks},
		{Path: "/bal", Method: api.Get, Handler: views.GetBal},
		{Path: "/bal/users", Method: api.Get, Handler: views.GetUsers},
		{Path: "/stats/call", Method: api.Get, Handler: views.CallStatus},
	}
}

func (r *StartConf) xAPIPort() string {
	ApiUrl, ok := os.LookupEnv("API_URL")
	if !ok {
		log.Warn("ASYNQ_MON_URL not Found! Setting Default Of :1300")
		ApiUrl = ":1300"
	}
	return ApiUrl
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
	go conf.RunMonitor(conf.Config.AsynqMonUrl)
}

func (r *StartConf) StartApi() {
	go api.RunApi(conf.Config.ApiUrl, r.xUrls())
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
