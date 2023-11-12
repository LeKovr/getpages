package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

var (
	sourceFile  string
	workerCount int
	reqTimeout  time.Duration

	errNoContent = errors.New("No content response")
)

func init() {
	flag.StringVar(&sourceFile, "source", "", "file with URLs (STDIN used if empty)")
	flag.DurationVar(&reqTimeout, "timeout", 5*time.Second, "request timeout")
	flag.IntVar(&workerCount, "workers", 2, "workers count")
}

func main() {
	flag.Parse()
	var err error
	defer func() {
		if err != nil {
			slog.Error("Exit", "error", err)
		}
	}()
	var sourceStream io.ReadCloser
	if sourceFile == "" {
		sourceStream = os.Stdin
	} else {
		sourceStream, err = os.Open(sourceFile)
		if err != nil {
			return
		}
		defer func() {
			err = errors.Join(err, sourceStream.Close())
		}()
	}
	client := &http.Client{Timeout: reqTimeout}
	nameStream := make(chan string, workerCount)
	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			ProcessStream(ctx, client, nameStream)
		}()
	}
	scanner := bufio.NewScanner(sourceStream)
	for scanner.Scan() {
		nameStream <- scanner.Text()
	}
	err = scanner.Err() // Printed on exit
	close(nameStream) // Stop workers
	wg.Wait()
}

// StatusError raised when Response status is not OK(200)
type StatusError struct {
	Status int
}

func (e StatusError) Error() string {
	return fmt.Sprintf("Status is not OK (%d)", e.Status)
}

// ResponseMeta holds response metadata
type ResponseMeta struct {
	Length int64
}

// ProcessStream listens channel and process received addresses.
func ProcessStream(ctx context.Context, client *http.Client, stream <-chan string) {
	for {
		reqURL, ok := <-stream
		if !ok { // channel closed, stop work
			return
		}
		start := time.Now()
		meta, err := Get(ctx, client, reqURL)
		elapsed := time.Since(start)
		var level slog.Level
		contextAttrs := make([]slog.Attr, 2, 3)
		if err != nil {
			level = slog.LevelError
			contextAttrs = append(contextAttrs, slog.String("error", err.Error()))
		} else {
			level = slog.LevelInfo
			contextAttrs = append(contextAttrs, slog.Int64("length", meta.Length))
		}
		contextAttrs = append(contextAttrs, slog.Duration("elapsed", elapsed))
		slog.LogAttrs(ctx, level, reqURL, contextAttrs...)
	}
}

// Get fetches resource by address and returns response body metadata.
func Get(ctx context.Context, client *http.Client, address string) (*ResponseMeta, error) {
	ctx, cancel := context.WithTimeout(ctx, client.Timeout)
	defer cancel()
	req, err := mkRequest(ctx, address)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	respLength, err := io.Copy(io.Discard, resp.Body)
	errClose := resp.Body.Close()
	err = errors.Join(err, errClose)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, StatusError{resp.StatusCode}
	}
	return &ResponseMeta{Length: respLength}, nil
}

// mkRequest valdates address and makes http.Request.
func mkRequest(ctx context.Context, address string) (*http.Request, error) {
	_, err := url.ParseRequestURI(address)
	if err != nil {
		return nil, err
	}
	return http.NewRequestWithContext(ctx, http.MethodGet, address, http.NoBody)
}
