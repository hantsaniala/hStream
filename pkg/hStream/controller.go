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
	r.ParseMultipartForm(100 << 20)           // Max file size: 100Mo
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

	encodeVideo(currUUID4)

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

	var video Video
	db.First(&video, "id = ?", input.Video)
	resp.UUID = video.ID

	// TODO: Move value directly to .env for download folder
	downloadDir := path.Join(GetEnv("UPLOAD_ROOT"), "download")
	destDir := path.Join(downloadDir, video.ID)
	if _, err := os.Stat(filepath.Join(destDir, "index.m3u8")); os.IsNotExist(err) {
		os.MkdirAll(destDir, 0644)
	}

	err = video.GenVideoKey(destDir)
	if err != nil {
		log.Fatal(err)
	}

	video.GenVideoKeyinfo(destDir)
	video.GenMetadata(filepath.Join(destDir, "metadata-playlist.json"), input.PlaylistData)
	video.GenMetadata(filepath.Join(destDir, "metadata.json"), input.PlaylistData)
	video.Encode2(input.Resolution, filepath.Join(destDir, "enc.keyinfo"), destDir)

	files, err := os.ReadDir(filepath.Join(destDir, fmt.Sprint(input.Resolution)))
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		err := MoveFile(filepath.Join(destDir, fmt.Sprint(input.Resolution), f.Name()), filepath.Join(destDir, f.Name()))
		if err != nil {
			log.Fatal(err)
		}
	}

	os.Remove(filepath.Join(destDir, fmt.Sprintf("index-%d.m3u8", input.Resolution)))
	os.RemoveAll(filepath.Join(destDir, fmt.Sprint(input.Resolution)))

	err = video.ArchiveAndCompress(destDir, downloadDir)
	if err != nil {
		log.Fatal(err)
	}

	video.RemoveFolder(destDir)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
