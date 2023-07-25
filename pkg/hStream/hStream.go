package hStream

import (
	"fmt"
	"html/template"
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
	PORT        = "9000"
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
	// TODO: Remove `/public/` from assets URL
	assetsDir := "/public/assets"
	router.PathPrefix(assetsDir).Handler(http.StripPrefix(assetsDir, http.FileServer(http.Dir("."+assetsDir))))

	router.HandleFunc("/", indexPage).Methods("GET")
	router.HandleFunc("/{id}", indexPage).Methods("GET")
	router.HandleFunc("/media/{mId}/stream/", streamMasterHandler).Methods("GET")
	router.HandleFunc("/media/{mId}/stream/{folder}/plist.m3u8/", streamFolderHandler).Methods("GET")
	router.HandleFunc("/media/{mId}/stream/{folder}/plist.m3u8/{segName:index[0-9]+.ts}", streamSegHandler).Methods("GET")

	router.HandleFunc("/api/v1/videos", PostVideo).Methods("POST")
	router.HandleFunc("/api/v1/videos", GetVideos).Methods("GET")
	router.HandleFunc("/api/v1/videos/{id}", GetVideo).Methods("GET")
	router.HandleFunc("/api/v1/videos/{id}", UpdateVideo).Methods("PUT")
	router.HandleFunc("/api/v1/videos/{id}", DeleteVideo).Methods("DELETE")

	return router
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("public/index.html")
	// log.Println(r.FormValue("id"))
	id := mux.Vars(r)["id"]
	wd := WebData{
		ID: id,
	}
	// log.Println(wd.ID)
	// http.ServeFile(w, r, "public/index.html")
	tmpl.Execute(w, &wd)
}

func streamMasterHandler(response http.ResponseWriter, request *http.Request) {
	log.Println("streamMasterHandler")
	vars := mux.Vars(request)
	// mId, err := strconv.Atoi(vars["mId"])
	// if err != nil {
	// 	response.WriteHeader(http.StatusNotFound)
	// 	return
	// }
	mId := vars["mId"]

	mediaBase := getMediaBase(mId)
	serveHlsM3u8(response, request, mediaBase, "", "index.m3u8")
}

func streamFolderHandler(response http.ResponseWriter, request *http.Request) {
	log.Println("streamFolderHandler")
	vars := mux.Vars(request)
	mId := vars["mId"]

	folder := vars["folder"]

	mediaBase := getMediaBase(mId)
	serveHlsM3u8(response, request, mediaBase, folder, "plist.m3u8")
}

func streamSegHandler(response http.ResponseWriter, request *http.Request) {
	log.Println("streamSegHandler")
	vars := mux.Vars(request)
	mId := vars["mId"]

	segName, hasNoSegName := vars["segName"]
	folder, hasNoFolder := vars["folder"]

	log.Printf("hasNoSegName: %t\thasNoFolder: %t", hasNoSegName, hasNoFolder)
	mediaBase := getMediaBase(mId)
	serveHlsTs(response, request, mediaBase, folder, segName)
}

func getMediaBase(mId string) string {
	mediaRoot := GetEnv("MEDIA_ROOT")
	return fmt.Sprintf("%s/%s", mediaRoot, mId)
}

func serveHlsM3u8(w http.ResponseWriter, r *http.Request, mediaBase string, folder string, m3u8Name string) {
	mediaFile := fmt.Sprintf("%s/%s", mediaBase, m3u8Name)
	log.Println("folder", folder)
	if folder != "" {
		mediaFile = fmt.Sprintf("%s/%s/%s", mediaBase, folder, m3u8Name)
	}
	log.Println(mediaFile)
	http.ServeFile(w, r, mediaFile)
	w.Header().Set("Content-Type", "application/x-mpegURL")
}

func serveHlsTs(w http.ResponseWriter, r *http.Request, mediaBase string, folder string, segName string) {
	mediaFile := fmt.Sprintf("%s/%s/%s", mediaBase, folder, segName)
	log.Println(mediaFile)
	http.ServeFile(w, r, mediaFile)
	w.Header().Set("Content-Type", "video/MP2T")
}
