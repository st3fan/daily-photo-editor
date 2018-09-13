package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rwcarlsen/goexif/exif"
)

const (
	availablePhotosPath = "/home/stefan/lib/daily-photo/photos/available"
	postedPhotosPath    = "/home/stefan/lib/daily-photo/photos/posted"
	failedPhotosPath    = "/home/stefan/lib/daily-photo/photos/failed"
)

var images []string

type editImageData struct {
	Index    int
	Name     string
	Previous int
	Next     int
	Time     time.Time
	Caption  string
}

func imageIndex(r *http.Request) int {
	vars := mux.Vars(r)
	index, err := strconv.Atoi(vars["index"])
	if err != nil {
		panic(err)
	}
	return index
}

func handleEdit(w http.ResponseWriter, r *http.Request) {
	index := imageIndex(r)

	f, err := os.Open(availablePhotosPath + "/" + images[index] + ".jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	t, err := x.DateTime()
	if err != nil {
		log.Fatal(err)
	}

	caption := ""
	captionPath := availablePhotosPath + "/" + images[index] + ".txt"
	if b, err := ioutil.ReadFile(captionPath); err == nil {
		caption = string(b)
	}

	tmpl := template.Must(template.ParseFiles("edit.html"))
	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, editImageData{
		Index:    index,
		Name:     images[index],
		Previous: index - 1,
		Next:     index + 1,
		Time:     t,
		Caption:  caption,
	})
}

func handleSave(w http.ResponseWriter, r *http.Request) {
	index := imageIndex(r)

	if err := r.ParseForm(); err != nil {
		log.Panic(err)
	}

	caption := r.Form.Get("caption")

	captionPath := availablePhotosPath + "/" + images[index] + ".txt"
	if err := ioutil.WriteFile(captionPath, []byte(caption), 0644); err != nil {
		log.Panic(err)
	}

	url := fmt.Sprintf("http://%s/edit/%d", r.Host, index)
	log.Println("Redirecting to: ", url)
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func handlePhoto(w http.ResponseWriter, r *http.Request) {
	index := imageIndex(r)
	image, err := ioutil.ReadFile(availablePhotosPath + "/" + images[index] + ".jpg")
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(image)
}

func main() {
	matches, err := filepath.Glob(availablePhotosPath + "/*.jpg")
	if err != nil {
		panic(err)
	}

	for _, path := range matches {
		_, file := filepath.Split(path)
		name := file[0 : len(file)-4]
		images = append(images, name)
	}

	r := mux.NewRouter()
	r.HandleFunc("/edit/{index}", handleEdit).Methods("GET")
	r.HandleFunc("/save/{index}", handleSave).Methods("POST")
	r.HandleFunc("/photo/{index}", handlePhoto)
	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
