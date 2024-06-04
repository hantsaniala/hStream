package hStream

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

	f, err := os.OpenFile(GetEnv("UPLOAD_ROOT")+"/original/"+newFileName, os.O_WRONLY|os.O_CREATE, 0666)

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

	keyinfoPath := path.Join("enc.keyinfo")

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

	db.Delete(&video, "id = ?", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("video deleted successfully")
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type DownloadRequestInput struct {
	PublicKey    string `json:"public_key"`
	PlaylistData string `json:"playlist_data"`
	Video        string `json:"video"`
	Resolution   int    `json:"resolution"`
}

type DownloadRequestResponse struct {
	UUID string `json:"uuid"`
}

func PrepareDownloadVideo(w http.ResponseWriter, r *http.Request) {
	var input DownloadRequestInput
	var resp DownloadRequestResponse

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

	resp.UUID = input.Video
	//TODO: Handle error

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type DownloadStatusResponse struct {
	Ready bool `json:"ready"`
}

func CheckDownloadStatus(w http.ResponseWriter, r *http.Request) {
	var stat DownloadStatusResponse
	id := mux.Vars(r)["id"]

	// TODO: Use better check
	fileP := path.Join(GetEnv("UPLOAD_ROOT"), "download", fmt.Sprintf("%s.tar.gz", id))
	if _, err := os.Stat(fileP); !errors.Is(err, os.ErrNotExist) {
		stat.Ready = true
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stat)
}
