package main

import (
	"log"
	"net/http"
	"os"
)

type Page struct {
	Path string
	Body []byte
}

var (
	// this is real annoying... should find a better way to do this but it works for now
	indexHtml *Page
	indexCss  *Page
)

func main() {
	indexHtml = loadPage("www/static/index.html")
	indexCss = loadPage("www/static/css/index.css")
	http.HandleFunc("/", handleRoot)

	log.Println("Web server up and running!")
	port := ":" + os.Getenv("port")
	http.ListenAndServe(port, nil)
}

func handleRoot(writer http.ResponseWriter, request *http.Request) {
	// ya, ya this is real bad but it's only 2 pages and I'll update it soon
	if request.RequestURI == "/" || request.RequestURI == "/index.html" {
		log.Println("Loading: (root)", request.RequestURI)
		writer.Write(indexHtml.Body)
	} else if request.RequestURI == "/css/index.css" {
		log.Println("Loading: (root)", request.RequestURI)
		writer.Header().Add("Content-Type", "text/css")
		writer.Write(indexCss.Body)
	}
}

func loadPage(path string) *Page {
	data, err := Asset(path)
	if err != nil {
		log.Println("Failed to find asset: ", path)
		return nil
	}
	return &Page{
		Path: path,
		Body: data,
	}
}
