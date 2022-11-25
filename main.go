package main

import (
	
	"flag"
	"fmt"
	// "golang.org/x/net/netutil"
	"html/template"
	"io"
	// "log"
	// "net"
	"net/http"
	"os"
	
	"path/filepath"
	
	"time"
)

type Progress struct {
	TotalSize int64
	BytesRead int64
}

const (
	defaultBindAddr = ":9990"
	MAX_UPLOAD_SIZE = 1024 * 1024
	defaultMaxConn  = 0
)

func main() {
	var (
		bindAddr string
		
	)

	flag.StringVar(&bindAddr, "b", defaultBindAddr, "TCP address the server will bind to")
	flag.Parse()
	file := http.FileServer(http.Dir("uploads"))

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest)
	mux.HandleFunc("/upload", uploadHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", file))

	server := &http.Server{
		Addr: bindAddr,
		Handler: mux,
		}
	server.ListenAndServe() 

	// listener, err := net.Listen("tcp", ":7070")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// srv.Serve(listener)
	

	
}

func (pr *Progress) Write(p []byte) (n int, err error) {
	n, err = len(p), nil
	pr.BytesRead += int64(n)
	pr.Print()
	return
}

// Print displays the current progress of the file upload
func (pr *Progress) Print() {
	if pr.BytesRead == pr.TotalSize {
		fmt.Println("DONE!")
		return
	}

	fmt.Printf("File upload in progress: %d\n", pr.BytesRead)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t, _ := template.ParseFiles("upload.html")
		t.Execute(w, nil)
	} else {
		// 32 MB is the default used by FormFile
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// get a reference to the fileHeaders
		files := r.MultipartForm.File["file"]

		for _, fileHeader := range files {
			if fileHeader.Size > MAX_UPLOAD_SIZE {
				http.Error(w, fmt.Sprintf("The uploaded image is too big: %s. Please use an image less than 1MB in size", fileHeader.Filename), http.StatusBadRequest)
				return
			}

			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer file.Close()

			buff := make([]byte, 512)
			_, err = file.Read(buff)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			filetype := http.DetectContentType(buff)
			if filetype != "image/jpeg" && filetype != "image/png" && filetype != "image/gif" && filetype != "text/html" && filetype != "text/plain" && filetype != "text/css"{
				http.Error(w, "The provided file format is not allowed. Please upload text/html, text/plain, image/gif, image/jpeg, image/jpeg, or text/css", http.StatusBadRequest)
				return
			}

			_, err = file.Seek(0, io.SeekStart)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = os.MkdirAll("./uploads", os.ModePerm)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			f, err := os.Create(fmt.Sprintf("./uploads/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			defer f.Close()

			pr := &Progress{
				TotalSize: fileHeader.Size,
			}

			_, err = io.Copy(f, io.TeeReader(file, pr))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		http.Redirect(w, r, "/", http.StatusFound)
		fmt.Fprintf(w, "Upload successful")
	}

}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, nil)
	// log.Printf("got request from %s\n", r.RemoteAddr)

	// w.WriteHeader(http.StatusOK)
	// w.Write([]byte("you got it"))
}
