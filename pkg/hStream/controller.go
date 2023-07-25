package hStream

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func PostVideo(w http.ResponseWriter, r *http.Request) {
	// TODO: handle FormFile input
	w.Header().Set("Content-Type", "application/json")
	var video Video
	currUUID4 := uuid.NewString()
	r.ParseMultipartForm(100 << 20)           // Max file size: 100Mo
	file, handler, err := r.FormFile("video") // retrieve the file from form data
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	newFileName := currUUID4 + "." + strings.Split(handler.Filename, ".")[1]

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

	// json.NewDecoder(r.Body).Decode(&video)
	db.Create(&video)

	encodeVideo(currUUID4)

	json.NewEncoder(w).Encode(video)
}

func GetVideo(w http.ResponseWriter, r *http.Request) {
	var video Video
	id := mux.Vars(r)["id"]
	db.First(&video, id)
	if video.ID == "" {
		json.NewEncoder(w).Encode("video not found!")
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

func UpdateVideo(w http.ResponseWriter, r *http.Request) {
	var video Video
	id := mux.Vars(r)["id"]
	db.First(&video, id)
	if video.ID == "" {
		json.NewEncoder(w).Encode("video not found!")
		return
	}
	json.NewDecoder(r.Body).Decode(&video)
	db.Save(&video)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(video)
}

func DeleteVideo(w http.ResponseWriter, r *http.Request) {
	var video Video
	id := mux.Vars(r)["id"]
	db.First(&video, id)
	if video.ID == "" {
		json.NewEncoder(w).Encode("video not found!")
		return
	}
	db.Delete(&video, id)
	// TODO: Delete existing file too
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("video deleted successfully")
}
