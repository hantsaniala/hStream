package hStream

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hibiken/asynq"
)

var (
	// ADDR    = "localhost"
	PORT        = "5480"
	hStream     *http.Server
	hTaskClient *asynq.Client
)

type WebData struct {
	ID string
}

// Start API server.
func StartServer() {
	listener, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		log.Fatal(err)
	}

	addr := listener.Addr().(*net.TCPAddr)
	log.Printf("Runing server on %s\n", addr.String())

	hStream = &http.Server{
		Addr:    addr.String(),
		Handler: handlers.LoggingHandler(os.Stdout, registerHandlers()),
	}

	err = hStream.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}

func registerHandlers() *mux.Router {
	router := mux.NewRouter()

	// media
	router.HandleFunc("/media/{mId}/stream/", streamMasterHandler)                               //.Methods("GET", "OPTIONS")
	router.HandleFunc("/media/{mId}/stream/{folder}/plist.m3u8", streamFolderHandler)            //.Methods("GET")
	router.HandleFunc("/media/{mId}/stream/{folder}/{segName:index[0-9]+.ts}", streamSegHandler) //.Methods("GET")

	// api
	router.HandleFunc("/api/v1/videos", PostVideo).Methods("POST")
	router.HandleFunc("/api/v1/videos", GetVideos).Methods("GET")
	router.HandleFunc("/api/v1/videos/{id}", GetVideo).Methods("GET")
	router.HandleFunc("/api/v1/videos/{id}", FullUpdateVideo).Methods("PUT")
	router.HandleFunc("/api/v1/videos/{id}", PartialUpdateVideo).Methods("PATCH")
	router.HandleFunc("/api/v1/videos/{id}", DeleteVideo).Methods("DELETE")

	// static
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./ui/dist/")))

	return router
}

func streamMasterHandler(response http.ResponseWriter, request *http.Request) {
	if (*request).Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(request)
	mId := vars["mId"]

	mediaBase := getMediaBase(mId)
	serveHlsM3u8(response, request, mediaBase, "", "index.m3u8")
}

func streamFolderHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	mId := vars["mId"]

	folder := vars["folder"]

	mediaBase := getMediaBase(mId)
	serveHlsM3u8(response, request, mediaBase, folder, "plist.m3u8")
}

func streamSegHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	mId := vars["mId"]

	segName, hasNoSegName := vars["segName"]
	folder, hasNoFolder := vars["folder"]

	log.Printf("hasNoSegName: %t\thasNoFolder: %t", hasNoSegName, hasNoFolder)
	mediaBase := getMediaBase(mId)
	serveHlsTs(response, request, mediaBase, folder, segName)
}

func serveHlsM3u8(w http.ResponseWriter, r *http.Request, mediaBase string, folder string, m3u8Name string) {
	mediaFile := fmt.Sprintf("%s/%s", mediaBase, m3u8Name)
	if folder != "" {
		mediaFile = fmt.Sprintf("%s/%s/%s", mediaBase, folder, m3u8Name)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Content-Type", "application/x-mpegURL")
	http.ServeFile(w, r, mediaFile)
}

func serveHlsTs(w http.ResponseWriter, r *http.Request, mediaBase string, folder string, segName string) {
	mediaFile := fmt.Sprintf("%s/%s/%s", mediaBase, folder, segName)
	w.Header().Set("Content-Type", "video/MP2T")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Content-Type", "application/x-mpegURL")
	http.ServeFile(w, r, mediaFile)
}
