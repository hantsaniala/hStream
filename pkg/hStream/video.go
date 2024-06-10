package hStream

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hantsaniala/hStream/pkg/gen"
	"github.com/hantsaniala/hStream/pkg/utils"
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
	Duration int `json:"-"`

	Duration2 float64 `json:"duration"`

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

// Get vertical resolution.
func (v *Video) GetResY() (int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=height",
		"-of", "csv=p=0",
		v.GetOriginalFilePath())

	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	sOut := string(out)
	sOut = strings.TrimSpace(sOut)
	iOut, err := strconv.Atoi(sOut)
	if err != nil {
		return 0, err
	}

	return iOut, nil
}

func (v *Video) GetOriginalFilePath() string {
	return path.Join(utils.GetEnv("UPLOAD_ROOT"), "original", v.ID+"."+getFileExt(v.FileName))
}

func (v *Video) GetEncodedDestinationPath() string {
	return path.Join(utils.GetEnv("MEDIA_ROOT"), v.ID)
}

// Set Video duration using ffprobe.
func (v *Video) SetDuration() error {
	filePath := v.GetOriginalFilePath()
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	sOut := string(out)
	sOut = strings.TrimSpace(sOut)

	d, err := strconv.ParseFloat(sOut, 64)
	if err != nil {
		return err
	}

	v.Duration2 = d

	return nil
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

	destDir := v.GetEncodedDestinationPath()

	if _, err := os.Stat(path.Join(destDir, "index.m3u8")); os.IsNotExist(err) {
		os.MkdirAll(destDir, 0644)
	}

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

// New version of `Encode()` that split step by encoding resolution.
func (v *Video) Encode2(res int, keyinfoPath string, destDir string) error {
	log.Printf("Encoding %s with resolution of %dp", v.ID[:8], res)

	if _, err := os.Stat(path.Join(destDir, fmt.Sprintf("index-%d.m3u8", res))); os.IsNotExist(err) {
		os.MkdirAll(destDir, 0644)
	}

	availOptions := GetEncodeOption()
	op := availOptions[res]

	desiredWidth := (res * 16) / 9

	encodingArgs := []string{
		"-i", v.GetOriginalFilePath(),
		"-vf", fmt.Sprintf("scale=w=%d:h=%d", desiredWidth, res),
		"-c:v", "libx264",
		"-x264-params", "nal-hrd=cbr:force-cfr=1",
		"-b:v", op.VideoBitrate,
		"-maxrate:v", op.VideoMaxRate,
		"-minrate:v", op.VideoMinRate,
		"-bufsize:v", op.VideoBufSize,
		"-preset", "slow",
		"-g", "48",
		"-sc_threshold", "0",
		"-keyint_min", "48",
		"-c:a", "aac",
		"-b:a", op.AudioBitrate,
		"-ac", "2",
		"-f", "hls",
		"-hls_time", "10",
		"-start_number", "0",
		"-hls_list_size", "0",
		"-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
		"-hls_segment_type", "mpegts",
		"-hls_key_info_file", keyinfoPath,
		"-master_pl_name", fmt.Sprintf("index-%d.m3u8", res),
		"-hls_segment_filename", path.Join(destDir, "%v/index%02d.ts"),
		"-var_stream_map", fmt.Sprintf("v:0,a:0,name:%d", res),
		path.Join(destDir, "%v/plist.m3u8"),
	}

	cmd := exec.Command("ffmpeg", encodingArgs...)

	out, err := cmd.CombinedOutput()

	if out != nil {
		log.Println(string(out))
	}

	if err != nil {
		log.Println(err.Error())
	}

	if err != nil && err.Error() != "exit status 1" {
		return err
	}
	return nil
}

// Generate master playlist from multiple master playlist
func (v *Video) MergeMasterPlaylist(resList []int) error {
	var commonS []string

	destDir := v.GetEncodedDestinationPath()
	masterFile := path.Join(destDir, "index.m3u8")

	for i, res := range resList {
		var uniq []string
		f, err := os.Open(path.Join(destDir, fmt.Sprintf("index-%d.m3u8", res)))
		if err != nil {
			log.Println("Error opening playlist")
			continue
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		line := 0

		for sc.Scan() {
			l := sc.Text()
			if line < 2 {
				commonS = append(commonS, l)
			} else {
				uniq = append(uniq, l)
			}
			line++
		}
		if err := sc.Err(); err != nil {
			log.Println("Error reading file", err)
			continue
		}

		if i == 0 {
			utils.WriteToFile(masterFile, commonS)
		}
		utils.WriteToFile(masterFile, []string{"\n"})
		utils.WriteToFile(masterFile, uniq)

		// Remove file after processing
		os.Remove(path.Join(destDir, fmt.Sprintf("index-%d.m3u8", res)))
	}

	return nil
}

func (v *Video) CopyKey() error {
	keyFile := filepath.Join(utils.GetEnv("KEYMASTER_FOLDER"), utils.GetEnv("KEY"))
	destDir := filepath.Join(utils.GetEnv("KEY_FOLDER"), v.ID)
	destFile := filepath.Join(destDir, utils.GetEnv("KEY"))

	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		os.MkdirAll(destDir, 0644)
	}

	err := CopyFile(keyFile, destFile)
	if err != nil {
		return err
	}
	gen.GenKeyinfo(destDir, fmt.Sprintf("%s/api/v1/key/%s/%s", utils.GetEnv("HOST"), v.ID, utils.GetEnv("KEY")))
	return nil
}

func (v *Video) GenMetadata(destPath, data string) error {
	f, err := os.OpenFile(destPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file for writing, err: %s", err)
	}
	defer f.Close()

	_, err = f.WriteString(data)
	if err != nil {
		return fmt.Errorf("error writing to file, err: %s", err)
	}

	return nil
}

func (v *Video) ArchiveAndCompress(source, target string) error {
	outputFile, err := os.Create(filepath.Join(target, fmt.Sprintf("%s.mp4", v.ID)))
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	// Create a gzip writer
	gzipWriter, _ := gzip.NewWriterLevel(outputFile, gzip.BestCompression)
	defer gzipWriter.Close()

	// Create a tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	// Walk through the folder and add files to the tar archive
	err = filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Create a new tar header
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			// Write the header to the tar archive
			err = tarWriter.WriteHeader(header)
			if err != nil {
				return err
			}

			// If the file is not a directory, write its contents to the tar archive
			if !info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				_, err = io.Copy(tarWriter, file)
				if err != nil {
					return err
				}
			}

			return nil
		})

	if err != nil {
		return err
	}
	return nil
}

func (v *Video) RemoveFolder(source string) error {
	err := os.RemoveAll(source)
	if err != nil {
		return err
	}
	return nil
}

func (v *Video) RemoveFile(source string) error {
	err := os.Remove(source)
	if err != nil {
		return err
	}
	return nil
}
