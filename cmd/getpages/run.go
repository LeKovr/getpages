package main

import (
	"context"
	"flag"
	"log/slog"
	"sync"
	"time"

	"github.com/LeKovr/getpages"
)

var (
	sourceFile  string
	workerCount int
	reqTimeout  time.Duration
)

func init() {
	flag.StringVar(&sourceFile, "source", "", "file with URLs (STDIN used if empty)")
	flag.DurationVar(&reqTimeout, "timeout", 5*time.Second, "request timeout")
	flag.IntVar(&workerCount, "workers", 2, "workers count")
}

// Run app and exit via given exitFunc.
func Run(ctx context.Context, exitFunc func(code int)) {
	flag.Parse()
	var err error
	defer func() {
		var code int
		if err != nil {
			slog.Error("Exit", "error", err)
			code = 1
		}
		exitFunc(code)
	}()
	gps := getpages.New(sourceFile, reqTimeout, workerCount)
	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			gps.ProcessStream(ctx)
		}()
	}
	go func() {
		WriteResults(ctx, gps)
	}()
	err = gps.ProcessSource() // read sourceFile and send to workers
	wg.Wait()                 // wait for ProcessStream workers ends
	gps.Close()               // wait for WriteResults ends
}

// WriteResults writes get results as logs.
func WriteResults(ctx context.Context, gps *getpages.Service) {
	var level slog.Level
	contextAttrs := make([]slog.Attr, 2)
	for result := range gps.ResultChan() {
		if result.Error != nil {
			level = slog.LevelError
			contextAttrs[0] = slog.String("error", result.Error.Error())
		} else {
			level = slog.LevelInfo
			contextAttrs[0] = slog.Int64("length", result.Length)
		}
		contextAttrs[1] = slog.Duration("elapsed", result.Elapsed)
		slog.LogAttrs(ctx, level, result.URL, contextAttrs...)
	}
	gps.ResultIsProcessed()
}
