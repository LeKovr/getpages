package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	PageOK       = "page"
	PageNotFound = "404"
	PageTimeout  = "sleep"
	PageNotValid = "::"
)

var (
	// WaitDuration is a request timeout
	WaitDuration = 5 * time.Millisecond
	// ExtraDuration must be longer than WaitDuration
	ExtraDuration = 10 * time.Millisecond

	testBody = []byte(`sometext`)

	tests = []struct {
		suffix    string
		addPrefix bool
		isTimeout bool
		wantErr   string
	}{
		{PageOK, true, false, ""},
		{PageNotFound, true, false, "Status is not OK"},
		{PageNotValid, false, false, "missing protocol scheme"},
		{PageTimeout, true, true, "context deadline exceeded"},
	}
)

// fillSource creates temp file and fill it with URIs from []tests
func fillSource(t *testing.T, server string) string {
	f, err := os.CreateTemp("", "tmpfile-")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	for _, tt := range tests {
		addr := tt.suffix
		if tt.addPrefix {
			addr = server + "/" + addr
		}
		f.WriteString(addr + "\n")
	}

	return f.Name()
}
func TestRun(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(testingHandler),
	)
	defer srv.Close()

	file := fillSource(t, srv.URL)
	defer os.Remove(file)
	a := os.Args
	os.Args = []string{"test", "-source", file, "-timeout", "5ms"}
	main()
	os.Args = a
}

func TestGet(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(testingHandler),
	)
	defer srv.Close()
	client := &http.Client{
		Timeout: WaitDuration,
	}
	ctx := context.Background()

	for _, tt := range tests {
		addr := tt.suffix
		if tt.addPrefix {
			addr = srv.URL + "/" + addr
		}
		meta, err := Get(ctx, client, addr)
		if tt.isTimeout && !os.IsTimeout(err) {
			t.Errorf("expected timeout, got %v", err)
		} else if err != nil {
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		} else {
			if meta.Length != int64(len(testBody)) {
				t.Errorf("expected %d, got %d", len(testBody), meta.Length)
			}
		}
	}
}

func testingHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/" + PageTimeout:
		time.Sleep(ExtraDuration)
	case "/" + PageNotFound:
		http.Error(w, "404", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(testBody)
}
