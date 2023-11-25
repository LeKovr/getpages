package getpages_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LeKovr/getpages"
	"github.com/LeKovr/getpages/pkgtest"
)

func TestGet(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(pkgtest.Handler),
	)
	defer srv.Close()
	client := &http.Client{
		Timeout: pkgtest.WaitDuration,
	}
	ctx := context.Background()

	for _, tt := range pkgtest.Tests {
		addr := tt.Suffix
		if tt.AddPrefix {
			addr = srv.URL + "/" + addr
		}
		meta, err := getpages.Get(ctx, client, addr)
		switch {
		case tt.IsTimeout:
			e := errors.Unwrap(err)
			if !os.IsTimeout(e) {
				t.Errorf("expected timeout, got %#v", err)
			}
		case err != nil:
			if !strings.Contains(err.Error(), tt.WantErr) {
				t.Errorf("expected error %v, got %v", tt.WantErr, err)
			}
		default:
			if meta.Length != int64(len(pkgtest.TestBody)) {
				t.Errorf("expected %d, got %d", len(pkgtest.TestBody), meta.Length)
			}
		}
	}
}

func TestWork(t *testing.T) {
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
	wc := 2
	gps := getpages.New(file, 5000*time.Microsecond, wc)
	var wg sync.WaitGroup
	wg.Add(wc)
	for i := 0; i < wc; i++ {
		go func() {
			defer wg.Done()
			gps.ProcessStream(ctx)
		}()
	}
	var totalLen int64
	const waitLen int64 = 8
	go func() {
		WriteResults(ctx, gps, &totalLen)
	}()
	err := gps.ProcessSource() // read sourceFile and send to workers
	wg.Wait()                  // wait for ProcessStream workers ends
	gps.Close()                // wait for WriteResults ends
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if totalLen != waitLen {
		t.Errorf("total length not eq, want %d, got %d", waitLen, totalLen)
	}
}

func WriteResults(_ context.Context, gps *getpages.Service, length *int64) {
	for result := range gps.ResultChan() {
		if result.Error == nil {
			*length += result.Length
		}
	}
	gps.ResultIsProcessed()
}
