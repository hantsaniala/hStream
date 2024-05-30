package hStream

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/hibiken/asynq"
)

// A list of task types.
const (
	TypeVideoEncode = "video:encode"
	// TypeImageResize   = "image:resize"
)

type VideoEncodePayload struct {
	UUID        string
	KeyInfoPath string
}

// type ImageResizePayload struct {
// 	SourceURL string
// }

//----------------------------------------------
// Write a function NewXXXTask to create a task.
// A task consists of a type and a payload.
//----------------------------------------------

func NewVideoEncodeTask(uuid string, keyinfoPath string) (*asynq.Task, error) {
	payload, err := json.Marshal(VideoEncodePayload{UUID: uuid, KeyInfoPath: keyinfoPath})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeVideoEncode, payload), nil
}

// func NewImageResizeTask(src string) (*asynq.Task, error) {
// 	payload, err := json.Marshal(ImageResizePayload{SourceURL: src})
// 	if err != nil {
// 		return nil, err
// 	}
// 	// task options can be passed to NewTask, which can be overridden at enqueue time.
// 	return asynq.NewTask(TypeImageResize, payload, asynq.MaxRetry(5), asynq.Timeout(20*time.Minute)), nil
// }

//---------------------------------------------------------------
// Write a function HandleXXXTask to handle the input task.
// Note that it satisfies the asynq.HandlerFunc interface.
//
// Handler doesn't need to be a function. You can define a type
// that satisfies asynq.Handler interface. See examples below.
//---------------------------------------------------------------

func HandleVideoEncodeTask(ctx context.Context, t *asynq.Task) error {
	var p VideoEncodePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	var vid Video
	db.Where(&Video{ID: p.UUID}).First(&vid)
	if vid.ID == "" {
		log.Fatalf("Video with id=%s not found", p.UUID[:8])
	}

	resX, err := vid.GetResY()
	if err != nil {
		log.Println(err)
	}

	availRes := []int{
		1080,
		720,
		540,
		360,
	}

	var outRes []int
	for _, r := range availRes {
		if resX > r {
			outRes = append(outRes, r)
		}
	}

	// Force resolution to be 360p if lower than all available resolution
	if len(outRes) == 0 {
		outRes = append(outRes, 360)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 1)

	destDir := vid.GetEncodedDestinationPath()

	for _, r := range outRes {
		wg.Add(1)
		go func(r int) {
			defer wg.Done()
			errs <- vid.Encode2(r, p.KeyInfoPath, destDir)
		}(r)
	}

	if err := <-errs; err != nil {
		log.Fatal(err)
	}

	wg.Wait()
	vid.IsReady = true
	db.Save(&vid)
	vid.MergeMasterPlaylist(outRes)

	return nil
}

// ImageProcessor implements asynq.Handler interface.
// type ImageProcessor struct {
// 	// ... fields for struct
// }

// func (processor *ImageProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
// 	var p ImageResizePayload
// 	if err := json.Unmarshal(t.Payload(), &p); err != nil {
// 		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
// 	}
// 	log.Printf("Resizing image: src=%s", p.SourceURL)
// 	// Image resizing code ...
// 	return nil
// }

// func NewImageProcessor() *ImageProcessor {
// 	return &ImageProcessor{}
// }
