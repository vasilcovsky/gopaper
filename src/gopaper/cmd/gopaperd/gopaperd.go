package main

import (
	"errors"
	"fmt"
	"github"
	"gocco"
	"net/http"
)

var (
	ErrLangNotSupportedYet = errors.New("Sorry, this kind of language is not supported yet")
)

type gendocReq struct {
	source   *gocco.SourceFile
	outputCh chan []byte
}

var (
	goccoCh = make(chan *gendocReq)
)

func goccoHandler(w http.ResponseWriter, r *http.Request) {
	candidate := r.RequestURI

	if ok := gocco.Allowed(candidate); !ok {
		fmt.Fprint(w, ErrLangNotSupportedYet)
		return
	}

	client := &http.Client{}
	crawler := github.NewCrawler(client)

	file, err := crawler.GetFile(candidate)

	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	if r.Header.Get("If-None-Match") == file.Header.Get("ETag") {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	content := &gocco.SourceFile{
		Path:    file.Path,
		Content: file.Content,
	}

	w.Header().Set("ETag", file.Header.Get("ETag"))
	w.Header().Set("Expires", file.Header.Get("Expires"))
	w.Write(gocco.GenerateDocumentation(content))
}

func goccoListener(ch <-chan *gendocReq) {
	for req := range ch {
		content := Generate(req.source)
		req.outputCh <- content
		close(req.outputCh)
	}
}

func Generate(content *gocco.SourceFile) []byte {
	var (
		ch      chan []byte = make(chan []byte)
		request *gendocReq  = &gendocReq{content, ch}
	)

	goccoCh <- request

	return <-ch
}

func main() {
	go goccoListener(goccoCh)

	http.HandleFunc("/", goccoHandler)
	http.ListenAndServe("0.0.0.0:8080", nil)
}
