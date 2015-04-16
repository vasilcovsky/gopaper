package github

import (
	"testing"
)

func TestTimeConsuming(t *testing.T) {
	url := DownloadURL("vasilcovsky/gopaper/blob/master/blob/f.go")
	expected := "https://raw.githubusercontent.com/vasilcovsky/gopaper/master/blob/f.go"
	if url != expected {
		t.Fatalf("Broken raw url: %s", url)
	}
}
