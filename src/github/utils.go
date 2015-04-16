package github

import (
	"strings"
)

// Converts path like /:user/:repo/blob/:branch/file...
// into actual URL with file content
func DownloadURL(blobURL string) string {
	s := strings.Replace(blobURL, "/blob/", "/", 1)
	return "https://raw.githubusercontent.com/" + s
}
