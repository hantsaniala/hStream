package hStream

import (
	"log"

	"github.com/hibiken/asynq"
)

// Start Task queue server.
func StartTaskClient() {
	hTaskClient = asynq.NewClient(asynq.RedisClientOpt{Addr: GetEnv("REDIS_SERVER")})
	// defer hTaskClient.Close()
}

func StartTaskServer() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: GetEnv("REDIS_SERVER")},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// See the godoc for other configuration options
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeVideoEncode, HandleVideoEncodeTask)
	// mux.Handle(tasks.TypeImageResize, tasks.NewImageProcessor())
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}

func EnqueueEncodeVideoTask(id string, keyinfoPath string) {
	StartTaskClient()

	task, err := NewVideoEncodeTask(id, keyinfoPath)
	if err != nil {
		log.Fatalf("Could not create task: %v", err)
	}

	info, err := hTaskClient.Enqueue(task)
	if err != nil {
		log.Fatalf("Could not enqueue task: %v", err)
	}
	log.Printf("Enqueued task: id=%s queue=%s videoId=%s", info.ID[:8], info.Queue, id[:8])

}
