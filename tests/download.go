package tests

import (
	"fmt"
	"github.com/brk0v/directio"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"syscall"
)

const (
	// ImageUri is 1.4 GB in size, & located in the PNW. Using wget the ISO was downloaded in 20s at ~94.0MB/s
	ImageUri         = "http://mirror.pnl.gov/releases/22.04/ubuntu-22.04.1-live-server-amd64.iso"
	ExpectedSize     = 1474873344
	FilenameTemplate = "/mnt/drive/%s"
)

func DownloadSetup() (string, error) {
	filename := fmt.Sprintf(FilenameTemplate, uuid.New().String())

	return filename, nil
}

func AssertSize(filename string) error {
	info, _ := os.Stat(filename)
	if info.Size() != ExpectedSize {
		return fmt.Errorf("file is corrupted")
	}
	return nil
}

func ControlDownload(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Downloading iso image
	resp, err := http.Get(ImageUri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(f, resp.Body)
	return err
}

func ODirectDownload(filename string) error {
	// Open file with O_DIRECT
	flags := os.O_WRONLY | os.O_EXCL | os.O_CREATE | syscall.O_DIRECT
	f, err := os.OpenFile(filename, flags, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Use directio writer
	dio, err := directio.New(f)
	if err != nil {
		return err
	}
	defer dio.Flush()

	// Downloading iso image
	resp, err := http.Get(ImageUri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(dio, resp.Body)
	return err
}

func IOUringDownload(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	iour, err := NewIOURingWriter(f)
	if err != nil {
		return err
	}
	defer iour.Close()

	req, err := http.Get(ImageUri)
	if err != nil {
		return err
	}
	_, err = io.Copy(iour, req.Body)
	if err != nil {
		return err
	}

	fmt.Println("Waiting...")
	iour.Wait()
	return nil
}
