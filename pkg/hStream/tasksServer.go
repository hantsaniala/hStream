package hStream

import (
	"log"

	"github.com/hantsaniala/hStream/pkg/utils"
	"github.com/hibiken/asynq"
)

// Start Task queue server.
func StartTaskClient() {
	hTaskClient = asynq.NewClient(asynq.RedisClientOpt{Addr: utils.GetEnv("REDIS_SERVER")})
	// defer hTaskClient.Close()
}

func StartTaskServer() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: utils.GetEnv("REDIS_SERVER")},
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
	mux.HandleFunc(TypeVideoDownloadPrepare, HandlePrepareVideoDownloadTask)
	mux.HandleFunc(TypeVideoRebuild, HandleVideoRebuildTask)
	// mux.Handle(tasks.TypeImageResize, tasks.NewImageProcessor())
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}

func EnqueueEncodeVideoTask(id string, keyinfoPath string) error {
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
	return nil
}

func EnqueueRebuildVideoTask(id string, keyinfoPath string) error {
	StartTaskClient()

	task, err := NewVideoRebuildTask(id, keyinfoPath)
	if err != nil {
		log.Fatalf("Could not create task: %v", err)
	}

	info, err := hTaskClient.Enqueue(task)
	if err != nil {
		log.Fatalf("Could not enqueue task: %v", err)
	}

	log.Printf("Enqueued task: id=%s queue=%s videoId=%s", info.ID[:8], info.Queue, id[:8])
	return nil
}

func EnqueueDownloadVideoTask(input DownloadRequestInput) error {
	StartTaskClient()

	task, err := NewVideoDownloadPrepareTask(input)
	if err != nil {
		log.Fatalf("Could not create task: %v", err)
	}

	info, err := hTaskClient.Enqueue(task)
	if err != nil {
		log.Fatalf("Could not enqueue task: %v", err)
	}

	log.Printf("Enqueued task: id=%s queue=%s videoId=%s", info.ID[:8], info.Queue, input.Video[:8])
	return nil
}
