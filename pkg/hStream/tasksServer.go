package hStream

import (
	"log"

	"github.com/hibiken/asynq"
)

var res = [][]int{
	{640, 360}, // 360p
	// {854, 480},   // 480p
	// {1280, 720},  // 720p
	// {1020, 1080}, // 1080p
}

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

func encodeVideo(id string) {
	StartTaskClient()

	for _, r := range res {
		task, err := NewVideoEncodeTask(id, r[0], r[1])
		if err != nil {
			log.Fatalf("Could not create task: %v", err)
		}

		info, err := hTaskClient.Enqueue(task)
		if err != nil {
			log.Fatalf("Could not enqueue task: %v", err)
		}
		log.Printf("Enqueued task: id=%s queue=%s videoId=%s", info.ID, info.Queue, id[:8])
	}

}
