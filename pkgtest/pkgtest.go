// //go:build test

package pkgtest

import (
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	// PageOK holds URL suffix for normal page.
	PageOK = "page"
	// PageNotFound holds URL suffix for page which not exists.
	PageNotFound = "404"
	// PageTimeout holds URL suffix for normal page with delayed response.
	PageTimeout = "sleep"
	// PageNotValid holds URL suffix for incorrect URL.
	PageNotValid = "::"
)

var (
	// WaitDuration is a request timeout.
	WaitDuration = 5 * time.Millisecond
	// ExtraDuration must be longer than WaitDuration.
	ExtraDuration = 10 * time.Millisecond

	// TestBody holds sample test body.
	TestBody = []byte(`sometext`)

	// Tests holds test data.
	Tests = []struct {
		Suffix    string
		AddPrefix bool
		IsTimeout bool
		WantErr   string
	}{
		{PageOK, true, false, ""},
		{PageNotFound, true, false, "Status is not OK"},
		{PageNotValid, false, false, "missing protocol scheme"},
		{PageTimeout, true, true, "context deadline exceeded"},
	}
)

// Handler serves testing http responses.
func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/" + PageTimeout:
		time.Sleep(ExtraDuration)
	case "/" + PageNotFound:
		http.Error(w, "404", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(TestBody)
}

// FillSource creates temp file and fill it with URIs from []Tests.
func FillSource(t *testing.T, server string) string {
	t.Helper()
	f, err := os.CreateTemp("", "tmpfile-")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		e := f.Close()
		if e != nil {
			t.Error(e)
		}
	}()

	for _, tt := range Tests {
		addr := tt.Suffix
		if tt.AddPrefix {
			addr = server + "/" + addr
		}
		_, err = f.WriteString(addr + "\n")
		if err != nil {
			t.Error(err)
		}
	}
	return f.Name()
}
