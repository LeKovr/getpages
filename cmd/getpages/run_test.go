package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/LeKovr/getpages/pkgtest"
)

func TestMain(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(pkgtest.Handler),
	)
	defer srv.Close()

	file := pkgtest.FillSource(t, srv.URL)
	defer func() {
		err := os.Remove(file)
		if err != nil {
			t.Error(err)
		}
	}()
	ctx := context.Background()
	var c int
	a := os.Args
	os.Args = []string{"test", "-source", file, "-timeout", "5ms"}
	Run(ctx, func(code int) { c = code })
	if c != 0 {
		t.Errorf("expected 0, got %d", c)
	}
	os.Args = []string{"test", "-source", "::"}
	Run(ctx, func(code int) { c = code })
	if c != 1 {
		t.Errorf("expected 0, got %d", c)
	}
	os.Args = a
}
