package github

import (
	"io/ioutil"
	"log"
	"net/http"
)

type Crawler struct {
	client *http.Client
}

type File struct {
	Path    string
	URL     string
	Content []byte
	Header  http.Header
}

func NewCrawler(client *http.Client) *Crawler {
	return &Crawler{client}
}

func (c Crawler) GetFile(uri string) (*File, error) {
	rawURL := DownloadURL(uri)
	res, err := c.client.Get(rawURL)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, http.ErrMissingFile
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	file := &File{
		Path:    uri,
		URL:     rawURL,
		Content: content,
		Header:  res.Header,
	}

	return file, nil
}
