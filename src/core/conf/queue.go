package conf

import (
	"context"
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

type queueStatus struct {
	Client    bool
	Worker    bool
	Scheduler bool
}
type QueueSchedules struct {
	Cron    string
	Key     string
	Payload []byte
}

type MuxHandler struct {
	Key     string
	Handler func(context.Context, *asynq.Task) error
}

func init() {
	// Create and configuring Redis connection.
	asyncQRedisClient = asynq.RedisClientOpt{
		Addr: RedisUrl, // Redis server address
	}
	QueueClient = asynq.NewClient(asyncQRedisClient)

	// Run worker server.
	QueueServer = asynq.NewServer(asyncQRedisClient, asynq.Config{
		Concurrency: 1,
		Queues: map[string]int{
			"critical": 6, // processed 60% of the time
			"default":  3, // processed 30% of the time
			"low":      1, // processed 10% of the time
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
func QueueStatus() queueStatus {
	return queueStatus{RunAsClient, RunAsServer, RunAsScheduler}
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
		_, err := QueueScheduler.Register(qs.Cron, asynq.NewTask(qs.Key, qs.Payload))
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
		RedisConnOpt: asyncQRedisClient})

	http.Handle(h.RootPath()+"/", h)

	// Go to http://localhost:8080/monitoring to see asynqmon homepage.
	log.Fatal(http.ListenAndServe(URL, nil))
}
