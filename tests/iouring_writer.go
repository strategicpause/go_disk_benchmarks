package tests

import (
	"fmt"
	"github.com/iceber/iouring-go"
	"os"
	"sync"
)

const (
	Entries uint = 1024
)

type IOURingWriter struct {
	file    *os.File
	iouring *iouring.IOURing
	ch      chan iouring.Result
	wg      *sync.WaitGroup
	offset  uint64
}

func NewIOURingWriter(file *os.File) (*IOURingWriter, error) {
	iour, err := iouring.New(Entries, iouring.WithDrain())
	if err != nil {
		return nil, fmt.Errorf("new IOURing error: %v", err)
	}
	iour.RegisterFile(file)

	r := IOURingWriter{
		file:    file,
		iouring: iour,
		ch:      make(chan iouring.Result, Entries),
		wg:      &sync.WaitGroup{},
	}
	r.start()

	return &r, nil
}

func (r *IOURingWriter) start() {
	go func() {
		for result := range r.ch {
			if err := result.Err(); err != nil {
				panic(err)
			}
			r.wg.Done()
		}
	}()
}

func (r *IOURingWriter) Write(b []byte) (int, error) {
	r.wg.Add(1)
	_, err := r.iouring.Pwrite(r.file, b, r.offset, r.ch)
	if err != nil {
		return 0, err
	}
	l := len(b)
	r.offset += uint64(l)
	return l, nil
}

func (r *IOURingWriter) Wait() {
	r.wg.Wait()
	close(r.ch)
}

func (r *IOURingWriter) Close() {
	r.iouring.Close()
}
