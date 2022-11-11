package main

import (
	"context"
	"diskio/tests"
	"fmt"
	"github.com/mackerelio/go-osstat/memory"
	"log"
	"os"
	"time"
)

type Execution func(string) error

type Setup func() (string, error)

type Validation func(string) error

func main() {
	if err := Run(tests.IOUringDownload, "io_uring", "download", tests.DownloadSetup, tests.AssertSize); err != nil {
		log.Fatal(err)
		return
	}
	if err := Run(tests.IOUringOverwrite, "io_uring", "overwrite", tests.CreateSparseFile, tests.AssertFile); err != nil {
		log.Fatal(err)
		return
	}
	if err := Run(tests.IOUringDirectOverwrite, "io_uring-direct", "overwrite", tests.CreateSparseFile, tests.AssertFile); err != nil {
		log.Fatal(err)
		return
	}
	if err := Run(tests.ControlDownload, "control", "download", tests.DownloadSetup, tests.AssertSize); err != nil {
		log.Fatal(err)
		return
	}
	if err := Run(tests.ControlOverwrite, "control", "overwrite", tests.CreateSparseFile, tests.AssertFile); err != nil {
		log.Fatal(err)
		return
	}
	if err := Run(tests.ODirectDownload, "o_direct", "download", tests.DownloadSetup, tests.AssertSize); err != nil {
		log.Fatal(err)
		return
	}
	if err := Run(tests.ODirectOverwrite, "o_direct", "overwrite", tests.CreateSparseFile, tests.AssertFile); err != nil {
		log.Fatal(err)
		return
	}
	if err := Run(tests.DDOverwrite, "dd", "overwrite", tests.CreateSparseFile, tests.AssertFile); err != nil {
		log.Fatal(err)
		return
	}
}

func Run(f Execution, name string, test string, setup Setup, validate Validation) error {
	fmt.Printf("Preparing test %s %s\n", name, test)
	filename, err := setup()
	defer os.Remove(filename)
	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	startMeasuring(ctx, name, test)
	fmt.Printf("Starting %s %s\n", name, test)
	if err = f(filename); err != nil {
		return err
	}
	timeElapsed := time.Since(start).Milliseconds()
	cancel()
	fmt.Printf("%s %s finished in %d ms.\n", filename, test, timeElapsed)
	logName := fmt.Sprintf("%s-%s.log", name, test)
	logFile, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer logFile.Close()
	fmt.Fprintln(logFile, timeElapsed)

	if err = validate(filename); err != nil {
		return err
	}
	return nil
}

func startMeasuring(ctx context.Context, name string, test string) {
	timer := time.Tick(time.Second)
	go func() {
		filename := fmt.Sprintf("%s-%s-mem.csv", name, test)
		memFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer memFile.Close()
		for {
			select {
			case <-timer:
				timeStamp := time.Now().Format(time.RFC3339)
				stats, err := memory.Get()
				if err != nil {
					log.Fatal(err)
					continue
				}
				fmt.Fprintf(memFile, "%s,%d,%d\n", timeStamp, int64(stats.Used)/tests.MiB, int64(stats.Cached)/tests.MiB)
			case <-ctx.Done():
				return
			}
		}
	}()
}
