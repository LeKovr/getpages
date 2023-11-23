package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(testingHandler),
	)
	defer srv.Close()

	file := fillSource(t, srv.URL)
	defer func() {
		err := os.Remove(file)
		if err != nil {
			t.Error(err)
		}
	}()
	a := os.Args
	os.Args = []string{"test", "-source", file, "-timeout", "5ms"}
	main()
	os.Args = a
}

// fillSource creates temp file and fill it with URIs from []tests.
func fillSource(t *testing.T, server string) string {
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

	for _, tt := range tests {
		addr := tt.suffix
		if tt.addPrefix {
			addr = server + "/" + addr
		}
		_, err = f.WriteString(addr + "\n")
		if err != nil {
			t.Error(err)
		}
	}
	return f.Name()
}
