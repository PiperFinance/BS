package main

import (
	"fmt"
	"math/rand"

	"github.com/PiperFinance/BS/src/api"
	"github.com/PiperFinance/BS/src/api/views"
	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/PiperFinance/BS/src/core/tasks/handlers"
	"github.com/PiperFinance/BS/src/utils"
	"github.com/hibiken/asynq"
)

type StartConf struct{}

func (r *StartConf) xChainSchedule() []conf.QueueSchedules {
	// NOTE - Enqueuing Jobs via scheduler... Use only supported Chains !
	sq := make([]conf.QueueSchedules, 0)
	for chainId := range conf.SupportedNetworks {
		sq = append(sq, conf.QueueSchedules{Cron: fmt.Sprintf("@every %ds", (15 + rand.Intn(15))), Payload: utils.MustBlockTaskGen(chainId), Q: asynq.Queue(conf.ScanQ), Timeout: conf.Config.ScanTaskTimeout, Key: tasks.BlockScanKey})
	}
	// Check online users
	sq = append(sq, conf.QueueSchedules{Cron: fmt.Sprintf("@every %ds", (2 + rand.Intn(5))), Q: asynq.Queue(conf.ScanQ), Timeout: conf.Config.UpdateOnlineUsersTaskTimeout, Key: tasks.UpdateOnlineUsersKey})
	return sq
}

func (r *StartConf) xHandlers() []conf.MuxHandler {
	return []conf.MuxHandler{
		{Key: tasks.BlockScanKey, Handler: handlers.BlockScanTaskHandler, Q: asynq.Queue(conf.ScanQ)},                    // 1
		{Key: tasks.FetchBlockEventsKey, Handler: handlers.BlockEventsTaskHandler, Q: asynq.Queue(conf.FetchQ)},          // 2
		{Key: tasks.ParseBlockEventsKey, Handler: handlers.ParseBlockEventsTaskHandler, Q: asynq.Queue(conf.ParseQ)},     // 3
		{Key: tasks.UpdateUserBalanceKey, Handler: handlers.UpdateUserBalTaskHandler, Q: asynq.Queue(conf.ProcessQ)},     // 4
		{Key: tasks.UpdateUserApproveKey, Handler: handlers.UpdateUserApproveTaskHandler, Q: asynq.Queue(conf.ProcessQ)}, // 4
		{Key: tasks.UpdateOnlineUsersKey, Handler: handlers.OnlineUsersHandler, Q: asynq.Queue(conf.UsersQ)},             //~TBD
		{Key: tasks.VacuumLogsKey, Handler: handlers.VacuumLogHandler, Q: asynq.Queue(conf.UnImportantQ)},                //~TBD
	}
}

func (r *StartConf) xUrls() []api.Route {
	return []api.Route{
		{Path: "/lsb", Method: api.Get, Handler: views.LastScannedBlock},
		{Path: "/lsb/100", Method: api.Get, Handler: views.LastScannedBlocks},
		{Path: "/bal", Method: api.Get, Handler: views.GetBal},
		{Path: "/mbal", Method: api.Get, Handler: views.GetUser},
		{Path: "/bal/users", Method: api.Get, Handler: views.GetUsers},

		{Path: "/stats/", Method: api.Get, Handler: views.Status},
		{Path: "/stats/call", Method: api.Get, Handler: views.CallStatus},
		{Path: "/stats/block", Method: api.Get, Handler: views.NewBlockStatus},
		{Path: "/stats/block/simple", Method: api.Get, Handler: views.NewBlockStatusSimple},
		{Path: "/stats/block/stats", Method: api.Get, Handler: views.BlockStats},
	}
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
	conf.Logger.Info("Starting Worker")
	r.StartWorker() // Consumer
	conf.Logger.Info("Starting Client")
	r.StartClient() // Producer
	conf.Logger.Info("Starting Scheduler")
	r.StartScheduler() // Scheduled Producer
	conf.Logger.Info("Starting AsynQMon")
	r.StartMon() // asynqMon
	conf.Logger.Info("Starting Api")
	r.StartApi()
}
