package conf

import (
	"github.com/hibiken/asynq"
	"log"
	"time"
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

func init() {
	// Create and configuring Redis connection.
	asyncQRedisClient = asynq.RedisClientOpt{
		Addr: RedisUrl, // Redis server address
	}
	QueueClient = asynq.NewClient(asyncQRedisClient)

	QueueServer = asynq.NewServer(asyncQRedisClient, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6, // processed 60% of the time
			"default":  3, // processed 30% of the time
			"low":      1, // processed 10% of the time
		},
	})
	// Run worker server.
	mux = asynq.NewServeMux()

	// Example of using America/Los_Angeles timezone instead of the default UTC timezone.
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

	// ... Register tasks

}
func RunClient() {
	RunAsClient = true
}

func RunWorker() {
	RunAsServer = true
	if err := QueueServer.Run(mux); err != nil {
		log.Fatal(err)
	}
}

func RunScheduler() {
	RunAsScheduler = true
	if err2 := QueueScheduler.Run(); err2 != nil {
		log.Fatal(err2)
	}
}
