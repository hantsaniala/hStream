package hStream

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hantsaniala/hStream/pkg/utils"
)

func PostVideo(w http.ResponseWriter, r *http.Request) {
	// TODO: handle FormFile input
	var video Video
	currUUID4 := uuid.NewString()
	r.ParseMultipartForm(1000 << 20)          // Max file size: 100Mo
	file, handler, err := r.FormFile("video") // retrieve the file from form data
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	newFileName := currUUID4 + "." + getFileExt(handler.Filename)

	f, err := os.OpenFile(utils.GetEnv("UPLOAD_ROOT")+"/original/"+newFileName, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	io.Copy(f, file)

	video = Video{
		ID:       currUUID4,
		FileName: handler.Filename,
		Title:    r.Form["title"][0],
	}

	err = video.SetDuration()
	if err != nil {
		log.Println(err)
	}

	// json.NewDecoder(r.Body).Decode(&video)
	db.Create(&video)

	video.CopyKey()

	var keyinfoPath string
	encrypt := r.Form["encrypt"][0]
	if encrypt != "" && encrypt == "true" {
		keyinfoPath = filepath.Join(utils.GetEnv("KEY_FOLDER"), video.ID, utils.GetEnv("KEYINFO"))
	}

	EnqueueEncodeVideoTask(currUUID4, keyinfoPath)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(video)
}

func GetVideo(w http.ResponseWriter, r *http.Request) {
	var video Video
	id := mux.Vars(r)["id"]
	db.First(&video, "id = ?", id)
	if video.ID == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	SetStreamURL(&video, r)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(video)
}

func GetVideos(w http.ResponseWriter, r *http.Request) {
	var videos []*Video
	db.Find(&videos)
	for _, v := range videos {
		SetStreamURL(v, r)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(videos)
}

func FullUpdateVideo(w http.ResponseWriter, r *http.Request) {
	var video Video
	id := mux.Vars(r)["id"]
	db.First(&video, "id = ?", id)
	if video.ID == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	json.NewDecoder(r.Body).Decode(&video)
	db.Save(&video)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(video)
}

func PartialUpdateVideo(w http.ResponseWriter, r *http.Request) {
	var video, partialVideo Video
	id := mux.Vars(r)["id"]
	db.First(&video, "id = ?", id)
	if video.ID == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&partialVideo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	if partialVideo.ID != "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Video ID can't be manualy set")
		return
	}

	if partialVideo.StreamURL != "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Video Stream URL can't be manualy set")
		return
	}

	if !partialVideo.CreatedAt.Equal(time.Time{}) && !partialVideo.CreatedAt.Equal(video.CreatedAt) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Video CreatedAt can't be manualy set")
		return
	}

	if !partialVideo.UpdatedAt.Equal(time.Time{}) && !partialVideo.UpdatedAt.Equal(video.UpdatedAt) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Video UpdatedAt can't be manualy set")
		return
	}

	partialVideo.UpdatedAt = time.Time{}

	db.Model(&video).Updates(partialVideo)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(video)
}

func DeleteVideo(w http.ResponseWriter, r *http.Request) {
	var video Video
	id := mux.Vars(r)["id"]
	db.First(&video, "id = ?", id)
	if video.ID == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	originalFile := video.GetOriginalFilePath()
	err := os.Remove(originalFile)
	if err != nil {
		log.Println(err)
	}

	encodedFolder := video.GetEncodedDestinationPath()
	err = os.RemoveAll(encodedFolder)
	if err != nil {
		log.Println(err)
	}

	keyFolder := filepath.Join(utils.GetEnv("KEY_FOLDER"), video.ID)
	err = os.RemoveAll(keyFolder)
	if err != nil {
		log.Println(err)
	}

	db.Delete(&video, "id = ?", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("video deleted successfully")
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type DownloadRequestInput struct {
	PublicKey     string                 `json:"public_key"`
	PlaylistData  map[string]interface{} `json:"playlist_data"`
	VideoData     map[string]interface{} `json:"video_data"`
	Video         string                 `json:"video"`
	Resolution    int                    `json:"resolution"`
	VID           string                 `json:"v_id"`
	WebhookReady  string                 `json:"webhook_ready"`
	WebhookDelete string                 `json:"webhook_delete"`
}

type DownloadRequestResponse struct {
	URL string `json:"url"`
}

func PrepareDownloadVideo(w http.ResponseWriter, r *http.Request) {
	var input DownloadRequestInput
	var resp DownloadRequestResponse
	input.VID = uuid.NewString()

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	err = EnqueueDownloadVideoTask(input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	resp.URL = fmt.Sprintf("%s/api/v1/file/%s", utils.GetEnv("HOST"), input.VID)
	//TODO: Handle error

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if strings.Contains(id, "/") || strings.Contains(id, "\\") || strings.Contains(id, "..") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("invalid id")
		return
	}

	filename := fmt.Sprintf("%s.%s", id, ARCHIVE_EXT)
	filepath := filepath.Join(utils.GetEnv("UPLOAD_ROOT"), "download", filename)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	http.ServeFile(w, r, filepath)
}

type DownloadStatusResponse struct {
	Ready bool `json:"ready"`
}

func CheckDownloadStatus(w http.ResponseWriter, r *http.Request) {
	var stat DownloadStatusResponse
	id := mux.Vars(r)["id"]

	if strings.Contains(id, "/") || strings.Contains(id, "\\") || strings.Contains(id, "..") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("invalid id")
		return
	}

	// TODO: Use better check
	fileP := path.Join(utils.GetEnv("UPLOAD_ROOT"), "download", fmt.Sprintf("%s.%s", id, ARCHIVE_EXT))
	if _, err := os.Stat(fileP); !errors.Is(err, os.ErrNotExist) {
		stat.Ready = true
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stat)
}

func ServeKey(w http.ResponseWriter, r *http.Request) {
	key, err := os.ReadFile(filepath.Join(utils.GetEnv("KEYMASTER_FOLDER"), utils.GetEnv("KEY")))
	if err != nil {
		http.Error(w, "Unable to read key file", http.StatusInternalServerError)
		return
	}
	// Set CORS headers (if necessary)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Set appropriate headers for security
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.Write(key)
}

func DeleteDownloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if strings.Contains(id, "/") || strings.Contains(id, "\\") || strings.Contains(id, "..") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("invalid id")
		return
	}

	fileP := path.Join(utils.GetEnv("UPLOAD_ROOT"), "download", fmt.Sprintf("%s.%s", id, ARCHIVE_EXT))
	err := os.Remove(fileP)
	if err != nil && errors.Is(err, &fs.PathError{}) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "file deleted successfully"})
}

func RebuildVideo(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var video Video
	db.Where(&Video{ID: id}).First(&video)
	if video.ID == "" {
		log.Fatalf("Video with id=%s not found", id[:8])
	}

	keyinfoPath := filepath.Join(utils.GetEnv("KEY_FOLDER"), video.ID, utils.GetEnv("KEYINFO"))

	EnqueueRebuildVideoTask(id, keyinfoPath)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(video)
}
