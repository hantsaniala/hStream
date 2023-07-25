package hStream

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

// Model
//
// TODO: Use uuid instead of auto incremented uint
type Video struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	// DeletedAt *time.Time `sql:"index"`

	// Manual primary key with uuid4 as value.
	ID string `gorm:"primaryKey"`

	// Video original filename after update.
	FileName string `json:"filename"`

	// Video title.
	Title string `json:"title"`

	// // Video author.
	// Author string `json:"author"`

	// Video duration.
	Duration int `json:"duration"`

	// // Default location path after video upload.
	// OriginalPath string `json:"originalPath"`

	// // Path of the encoded video.
	// EncodedPath string `json:"encodedPath"`

	// By default `false`, set as `true` when video is fully encoded.
	IsReady bool `json:"isReady"`

	// Dynamic field that will be filled with the stream url of th the video.
	StreamURL string `json:"streamURL"`
}

// Set uuid4 as `Video.Id` value.
//
// Source: https://stackoverflow.com/a/68370363/5527968
// func (v *Video) BeforeCreate(tx *gorm.DB) (err error) {
// 	// Note: Gorm will fail if the function signature
// 	//  does not include `*gorm.DB` and `error`

// 	// UUID version 4
// 	v.ID = uuid.NewString()
// 	return
// }

// Set Video.StreamURL manually to generate URL.
func SetStreamURL(v *Video, r *http.Request) error {
	prtcl := "http://"
	if r.TLS != nil {
		prtcl = "https://"
	}
	// v.StreamURL = prtcl + r.URL.Scheme + r.Host + "/media/" + strconv.FormatUint(uint64(v.ID), 10) + "/stream/"
	v.StreamURL = prtcl + r.URL.Scheme + r.Host + "/media/" + v.ID + "/stream/"
	return nil
}

func (v *Video) GetOriginalFilePath() string {
	return path.Join(GetEnv("UPLOAD_ROOT"), "original", v.ID+"."+strings.Split(v.FileName, ".")[1])
}

func (v *Video) GetEncodedDestinationPath(format string, resX int, resY int) string {
	// return path.Join(GetEnv("MEDIA_ROOT"), v.ID, format, strconv.Itoa(resY))
	return path.Join(GetEnv("MEDIA_ROOT"), v.ID)
}

// Create a folder with the video UUID as name.
func (v *Video) PrepareFolder() error {
	return nil
}

// Encode video to given format.
func (v *Video) Encode(format string, resX int, resY int) error {
	if format == "" {
		format = "hls"
	}

	destDir := v.GetEncodedDestinationPath(format, resX, resY)

	if _, err := os.Stat(path.Join(destDir, "index.m3u8")); os.IsNotExist(err) {
		os.MkdirAll(destDir, 0644)
	}

	// out, err := exec.Command("ffmpeg",
	// 	"-i", v.GetOriginalFilePath(),
	// 	"-profile:v", "baseline",
	// 	"-level", "3.0",
	// 	"-s", strconv.Itoa(resX)+"x"+strconv.Itoa(resY),
	// 	"-start_number", "0",
	// 	"-hls_time", "10",
	// 	"-hls_list_size", "0",
	// 	"-f", "hls",
	// 	path.Join(destDir, "index.m3u8")).Output()

	cmd := exec.Command("ffmpeg",
		"-i", v.GetOriginalFilePath(),
		"-filter_complex",
		"[0:v]split=3[v1][v2][v3]; [v1]copy[v1out]; [v2]scale=w=1280:h=720[v2out]; [v3]scale=w=640:h=360[v3out]",

		"-map", "[v1out]", "-c:v:0", "libx264", "-x264-params", "nal-hrd=cbr:force-cfr=1", "-b:v:0", "5M", "-maxrate:v:0", "5M", "-minrate:v:0", "5M", "-bufsize:v:0", "10M", "-preset", "slow", "-g", "48", "-sc_threshold", "0", "-keyint_min", "48",
		"-map", "[v2out]", "-c:v:1", "libx264", "-x264-params", "nal-hrd=cbr:force-cfr=1", "-b:v:1", "3M", "-maxrate:v:1", "3M", "-minrate:v:1", "3M", "-bufsize:v:1", "3M", "-preset", "slow", "-g", "48", "-sc_threshold", "0", "-keyint_min", "48",
		"-map", "[v3out]", "-c:v:2", "libx264", "-x264-params", "nal-hrd=cbr:force-cfr=1", "-b:v:2", "1M", "-maxrate:v:2", "1M", "-minrate:v:2", "1M", "-bufsize:v:2", "1M", "-preset", "slow", "-g", "48", "-sc_threshold", "0", "-keyint_min", "48",

		"-map", "a:0", "-c:a:0", "aac", "-b:a:0", "96k", "-ac", "2",
		"-map", "a:0", "-c:a:1", "aac", "-b:a:1", "96k", "-ac", "2",
		"-map", "a:0", "-c:a:2", "aac", "-b:a:2", "48k", "-ac", "2",

		"-f", "hls",
		"-hls_time", "10",
		"-start_number", "0",
		"-hls_list_size", "0",
		"-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
		"-hls_segment_type", "mpegts",
		"-master_pl_name", "index.m3u8",
		"-hls_segment_filename", path.Join(destDir, "%v/index%02d.ts"),
		"-var_stream_map", "v:0,a:0,name:1080 v:1,a:1,name:720 v:2,a:2,name:360", path.Join(destDir, "%v/plist.m3u8"),
	)
	out, err := cmd.CombinedOutput()

	if out != nil {
		log.Println(string(out))
	}

	if err != nil && err.Error() != "exit status 1" {
		log.Fatal(err)
	}
	return nil
}
