package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	extJPG  = "jpg"
	extJPEG = "jpeg"
	extPNG  = "png"
)

// UploadResponse holds response to the upload image endpoint
type UploadResponse struct {
	ID string `json:"id,omitempty"`
}

// Service holds the logic for uploading and accessing files
type Service struct {
	imageDir string
}

func (s Service) uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeTextResponse(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	// check size
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		writeTextResponse(
			w,
			fmt.Sprintf("Bad Request: %v", err),
			http.StatusBadRequest,
		)
		return
	}

	file, handler, err := r.FormFile("fileupload")
	if err != nil {
		writeTextResponse(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)
		return
	}
	defer file.Close()

	// allowed extensions png, jpg/jpeg
	fileSplit := strings.Split(handler.Filename, ".")
	extension := fileSplit[len(fileSplit)-1]

	if extension != extPNG && extension != extJPEG && extension != extJPG {
		writeTextResponse(
			w,
			"Bad Request: Extension not allowed",
			http.StatusBadRequest,
		)
		return
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		writeTextResponse(
			w,
			"Internal Server Error: File can not be created",
			http.StatusInternalServerError,
		)
		return
	}

	fileID := rand.Int()
	fileIDString := strconv.Itoa(fileID)

	err = os.WriteFile(fmt.Sprintf(
		"%s/%s.%s",
		s.imageDir,
		fileIDString,
		extension,
	), fileBytes, 0755)
	if err != nil {
		writeTextResponse(
			w,
			"Internal Server Error: File can not be created",
			http.StatusInternalServerError,
		)
		return
	}
	resp := UploadResponse{ID: fileIDString}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
	return
}

func (s Service) accessFile(w http.ResponseWriter, r *http.Request) {
	var resultFileName, extension string

	path := r.URL.Path
	fileID := strings.TrimPrefix(path, "/image/")

	dir := os.DirFS(s.imageDir)
	entries, err := fs.ReadDir(dir, ".")
	if err != nil {
		writeTextResponse(
			w,
			"Internal Server Error: Directory can not be read",
			http.StatusInternalServerError,
		)
	}

	for _, e := range entries {
		fileSplit := strings.Split(e.Name(), ".")
		fileWithoutExtension := fileSplit[0]
		extension = fileSplit[len(fileSplit)-1]
		if fileWithoutExtension == fileID {
			resultFileName = e.Name()
		}
	}
	if resultFileName == "" {
		writeTextResponse(
			w,
			"Not Found: File not found",
			http.StatusNotFound,
		)
		return
	}
	file, err := dir.Open(resultFileName)

	if err != nil {
		writeTextResponse(
			w,
			"Internal Server Error: File can not be created",
			http.StatusInternalServerError,
		)
		return
	}
	defer file.Close()

	image, err := io.ReadAll(file)
	if err != nil {
		writeTextResponse(
			w,
			"Internal Server Error: File can not be read",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", extension))
	w.WriteHeader(http.StatusOK)
	w.Write(image)
	return
}

func writeTextResponse(
	w http.ResponseWriter,
	message string,
	// contentType string,
	status int,
) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, message+"\n")
}

func main() {

	mux := http.NewServeMux()

	service := Service{imageDir: "images"}
	mux.HandleFunc("/image", service.uploadFile)
	mux.HandleFunc("/image/", service.accessFile)

	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	s.ListenAndServe()
}
