package main

import (
	"errors"
	"fmt"
	"github"
	"gocco"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var (
	ErrLangNotSupportedYet = errors.New("Sorry, this language is not supported yet")
)

type AppContext struct {
	// path to template directory
	tplPath string
	// path to static directory
	staticPath string
	// flag to check ETag header in http request
	allowETag bool
	// parsed template for view code
	contentTpl *template.Template
}

// Create new application context from Environment variables
func NewFromEnv() (*AppContext, error) {
	tplPath := os.Getenv("GOPAPERD_TPL_PATH")
	staticPath := os.Getenv("GOPAPERD_STATIC_PATH")
	allowETag, err := strconv.ParseBool(os.Getenv("GOPAPERD_ALLOW_ETAG"))

	if err != nil {
		return nil, err
	}

	tplContent := filepath.Join(tplPath, "content.tmpl")
	tpl, err := template.ParseFiles(tplContent)
	if err != nil {
		return nil, err
	}

	ctx := &AppContext{
		tplPath: tplPath,
		staticPath: staticPath,
		allowETag: allowETag,
		contentTpl: tpl,
	}

	return ctx, nil
}

// Handler for Github URL
func (ctx AppContext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	if ctx.allowETag && r.Header.Get("If-None-Match") == file.Header.Get("ETag") {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	content := &gocco.SourceFile{
		Path:    file.Path,
		Content: file.Content,
	}

	doc := gocco.GenerateDocumentation(content, ctx.contentTpl)

	w.Header().Set("ETag", file.Header.Get("ETag"))
	w.Header().Set("Expires", file.Header.Get("Expires"))
	w.Write(doc)
}

func main() {
	ctx, err := NewFromEnv()

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Template path: %s", ctx.tplPath)
	log.Printf("Static path: %s", ctx.staticPath)
	log.Printf("Allow ETag: %t", ctx.allowETag)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(ctx.staticPath))))
	http.Handle("/", ctx)

	log.Fatalln(http.ListenAndServe("0.0.0.0:8080", nil))
}
