// Package getpages used to get pages via URL list
package getpages

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	// ErrOpenSource indicates error with source file opening.
	ErrOpenSource = "open source: %w"
	// ErrCloseSourceStream indicates error with source stream closing.
	ErrCloseSourceStream = "close source stream: %w"
	// ErrParseURI raised on unparseable URL.
	ErrParseURI = "parse URL: %w"
	// ErrNewRequest indicates error with request creating.
	ErrNewRequest = "create request: %w"
	// ErrDoRequest indicates error with request doing.
	ErrDoRequest = "do request: %w"
	// ErrReadRequest indicates error with request reading.
	ErrReadRequest = "read request: %w"
)

// ResponseMeta holds response metadata.
type ResponseMeta struct {
	Length int64
}

// Response holds response attributes.
type Response struct {
	Length  int64
	URL     string
	Error   error
	Elapsed time.Duration
}

// Service holds service attributes.
type Service struct {
	sourceFile   string
	client       *http.Client
	nameStream   chan string
	resultStream chan Response
	quit         chan struct{}
}

// New returns initialized *Service.
func New(sourceFile string, reqTimeout time.Duration, sourceChanLen int) *Service {
	return &Service{
		sourceFile:   sourceFile,
		client:       &http.Client{Timeout: reqTimeout},
		nameStream:   make(chan string, sourceChanLen),
		resultStream: make(chan Response),
		quit:         make(chan struct{}),
	}
}

// ProcessSource reads sourceStream, sends to nameStream and closes nameStream when done.
func (gps Service) ProcessSource() error {
	var err error
	var sourceStream io.ReadCloser
	if gps.sourceFile == "" {
		sourceStream = os.Stdin
	} else {
		sourceStream, err = os.Open(filepath.Clean(gps.sourceFile))
		if err != nil {
			return fmt.Errorf(ErrOpenSource, err)
		}
		defer func() {
			errClose := sourceStream.Close()
			if errClose != nil {
				errClose = fmt.Errorf(ErrCloseSourceStream, errClose)
			}
			err = errors.Join(err, errClose)
		}()
	}
	scanner := bufio.NewScanner(sourceStream)
	var lineNo int64
	for scanner.Scan() {
		address := scanner.Text()
		_, err := url.ParseRequestURI(address)
		if err != nil {
			slog.Error("ParseURI", "err", err, "line", lineNo)
			gps.resultStream <- Response{Error: fmt.Errorf(ErrParseURI, err)}
		} else {
			gps.nameStream <- address
		}
		lineNo++
	}
	close(gps.nameStream) // Stop workers
	return scanner.Err()
}

// Close waits for WriteResults completion.
func (gps Service) Close() {
	close(gps.resultStream) // stop WriteResults
	<-gps.quit              // wait for WriteResults ends
}

// ResultChan returns channel with service results.
func (gps Service) ResultChan() <-chan Response {
	return gps.resultStream
}

// ResultIsProcessed called by result processor as job done signal.
func (gps Service) ResultIsProcessed() {
	gps.quit <- struct{}{} // Signal to parent: write is done
}

// StatusError raised when Response status is not OK(200).
type StatusError struct {
	Status int
}

// Error returns status as error string.
func (e StatusError) Error() string {
	return fmt.Sprintf("Status is not OK (%d)", e.Status)
}

// ProcessStream listens channel and process received addresses.
func (gps Service) ProcessStream(ctx context.Context) {
	for reqURL := range gps.nameStream {
		start := time.Now()
		meta, err := Get(ctx, gps.client, reqURL)
		rv := Response{
			URL:     reqURL,
			Error:   err,
			Elapsed: time.Since(start),
		}
		if err == nil {
			rv.Length = meta.Length
		}
		gps.resultStream <- rv
	}
}

// Get fetches resource by address and returns response body metadata.
func Get(ctx context.Context, client *http.Client, address string) (*ResponseMeta, error) {
	ctx, cancel := context.WithTimeout(ctx, client.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf(ErrNewRequest, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(ErrDoRequest, err)
	}
	respLength, err := io.Copy(io.Discard, resp.Body)

	errClose := resp.Body.Close()
	err = errors.Join(err, errClose)
	if err != nil {
		return nil, fmt.Errorf(ErrReadRequest, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, StatusError{resp.StatusCode}
	}
	return &ResponseMeta{Length: respLength}, nil
}
