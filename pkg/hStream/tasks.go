package hStream

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
)

// A list of task types.
const (
	TypeVideoEncode = "video:encode"
	// TypeImageResize   = "image:resize"
)

type VideoEncodePayload struct {
	UUID string
	ResX int
	ResY int
}

// type ImageResizePayload struct {
// 	SourceURL string
// }

//----------------------------------------------
// Write a function NewXXXTask to create a task.
// A task consists of a type and a payload.
//----------------------------------------------

func NewVideoEncodeTask(uuid string, resX int, resY int) (*asynq.Task, error) {
	payload, err := json.Marshal(VideoEncodePayload{UUID: uuid, ResX: resX, ResY: resY})
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
	log.Printf("Encoding %s with resolution of %dp", p.UUID[:8], p.ResY)
	// TODO: Get Video PATH from UUID
	// TODO: Encode video with ffmpeg

	// ```
	// ffmpeg -i jwb-096_MG_01_r360P.mp4 -profile:v baseline -level 3.0 -s 640x360 -start_number 0 -hls_time 10 -hls_list_size 0 -f hls index.m3u8
	// ```

	var vid Video
	db.Where(&Video{ID: p.UUID}).First(&vid)
	if vid.ID == "" {
		log.Fatalf("Video with id=%s not found", p.UUID[:8])
	}
	err := vid.Encode("", p.ResX, p.ResY)
	if err != nil {
		log.Fatal(err)
	}
	vid.IsReady = true
	db.Save(&vid)
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
