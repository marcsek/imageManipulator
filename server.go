package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gonum.org/v1/gonum/mat"
)

func StreamFile(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/octet-stream")

	switch t := v.(type) {
	case []byte:
		_, err := w.Write(t)
		return err
	case ApiError:
		return fmt.Errorf(t.Error)
	default:
		return fmt.Errorf("server error")
	}
}

type ApiError struct {
	Error string
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			StreamFile(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

type ApiServer struct {
	listenAddr   string
	imageHandler ImageHandler
}

func NewApiServer(listerAddr string, imageHandler ImageHandler) *ApiServer {
	return &ApiServer{
		listenAddr:   listerAddr,
		imageHandler: imageHandler,
	}
}

func (s *ApiServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/processImage", makeHTTPHandleFunc(s.handleRequest))

	log.Println("Api server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *ApiServer) handleRequest(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		return s.processImage(w, r)
	}

	return fmt.Errorf("method not allow %s", r.Method)
}

func (s *ApiServer) processImage(w http.ResponseWriter, r *http.Request) error {
	rotate := r.URL.Query().Get("rotate")
	graysale := r.URL.Query().Get("grayscale")
	blur := r.URL.Query().Get("blur")

	image, err := png.Decode(r.Body)
	if err != nil {
		return fmt.Errorf("image could not be processed %s", "daco")
	}

	imageBuffer := s.imageHandler.CreateTensor(image)
	if rotate == "TRUE" {
		s.imageHandler.RotateImage(&imageBuffer)
	}
	if graysale == "TRUE" {
		s.imageHandler.GrayScaleImage(&imageBuffer)
	}
	if blur == "TRUE" {
		s.imageHandler.BlurImage(&imageBuffer, mat.NewDense(3, 3, []float64{
			1.0 / 9, 1.0 / 9, 1.0 / 9,
			1.0 / 9, 1.0 / 9, 1.0 / 9,
			1.0 / 9, 1.0 / 9, 1.0 / 9,
		}))
	}

	newImage := s.imageHandler.DecodeTensor(imageBuffer)

	bufferToSend := new(bytes.Buffer)
	error := jpeg.Encode(bufferToSend, newImage, nil)
	if error != nil {
		return fmt.Errorf("image could not be transformed")
	}

	return StreamFile(w, http.StatusOK, bufferToSend.Bytes())
}
