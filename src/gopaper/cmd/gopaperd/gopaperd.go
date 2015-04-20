package main

import (
	"net/http"
	"gocco"
	"fmt"
	"html/template"
	"github"
	"errors"
	"os"
	"strconv"
	"path/filepath"
	"log"
)

var (
	ErrLangNotSupportedYet = errors.New("Sorry, this language is not supported yet")
)

type AppContext struct {
	// path to template directory
	tplPath string
	// flag to check ETag header in http request
	allowETag bool
	// parsed template for view code
	contentTpl *template.Template
}

// Create new application context from Environment variables
func NewFromEnv() (*AppContext, error) {
	tplPath := os.Getenv("GOPAPERD_TPL_PATH")
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
		tplPath,
		allowETag,
		tpl,
	}

	return ctx, nil
}

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

	w.Header().Set("ETag", file.Header.Get("ETag"))
	w.Header().Set("Expires", file.Header.Get("Expires"))
	w.Write(gocco.GenerateDocumentation(content, ctx.contentTpl))
}

func main() {
	ctx, err := NewFromEnv()

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Template path: %s", ctx.tplPath)
	log.Printf("Allow ETag: %t", ctx.allowETag)

	http.Handle("/", ctx)
	log.Fatalln(http.ListenAndServe("0.0.0.0:8080", nil))
}
