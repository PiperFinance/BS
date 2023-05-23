package conf

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
)

var (
	RunAsServer       bool
	RunAsClient       bool
	RunAsScheduler    bool
	QueueScheduler    *asynq.Scheduler
	QueueClient       *asynq.Client
	QueueServer       *asynq.Server
	asyncQRedisClient asynq.RedisClientOpt
	mux               *asynq.ServeMux
)

//	type queueStatus struct {
//		Client    bool
//		Worker    bool
//		Scheduler bool
//	}
type QueueSchedules struct {
	Cron    string
	Key     string
	Payload []byte
	Q       asynq.Option
	Timeout time.Duration
}

type MuxHandler struct {
	Key     string
	Handler func(context.Context, *asynq.Task) error
	Q       asynq.Option
}

const (
	ScanQ        = "scan"
	FetchQ       = "fetch"
	ParseQ       = "Parse"
	ProcessQ     = "Process"
	MainQ        = "main"
	DefaultQ     = "default"
	UnImportantQ = "Un-Important"
)

func LoadQueue() {
	// Create and configuring Redis connection.
	asyncQRedisClient = asynq.RedisClientOpt{
		Addr: fmt.Sprintf("%s:%s", Config.RedisHost, Config.RedisPort),
		DB:   Config.RedisDB,
	}
	QueueClient = asynq.NewClient(asyncQRedisClient)

	// Run worker server.
	QueueServer = asynq.NewServer(asyncQRedisClient, asynq.Config{
		Concurrency:  Config.MaxConcurrency,
		ErrorHandler: &QueueErrorHandler{},
		Queues: map[string]int{
			ProcessQ:     7,
			FetchQ:       5,
			ParseQ:       6,
			ScanQ:        4,
			MainQ:        6,
			DefaultQ:     3,
			UnImportantQ: 1,
		},
	})
	mux = asynq.NewServeMux()
	// Block Related

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal(err)
	}
	QueueScheduler = asynq.NewScheduler(
		asyncQRedisClient,
		&asynq.SchedulerOpts{
			Location: loc,
		},
	)
}

func RunClient() {
	RunAsClient = true
}

func RunWorker(muxHandler []MuxHandler) {
	RunAsServer = true
	for _, mh := range muxHandler {
		mux.HandleFunc(mh.Key, mh.Handler)
	}
	if err := QueueServer.Run(mux); err != nil {
		log.Fatal(err)
	}
}

func RunScheduler(queueSchedules []QueueSchedules) {
	RunAsScheduler = true
	for _, qs := range queueSchedules {
		_, err := QueueScheduler.Register(qs.Cron, asynq.NewTask(qs.Key, qs.Payload), qs.Q, asynq.Timeout(qs.Timeout))
		if err != nil {
			log.Fatalf("QueueScheduler: %s", err)
		}
	}
	if err2 := QueueScheduler.Start(); err2 != nil {
		log.Fatal(err2)
	}
}

func RunMonitor(URL string) {
	h := asynqmon.New(asynqmon.Options{
		RootPath:     "/mon",
		RedisConnOpt: asyncQRedisClient,
	})
	http.Handle(h.RootPath()+"/", h)
	log.Fatal(http.ListenAndServe(URL, nil))
}
