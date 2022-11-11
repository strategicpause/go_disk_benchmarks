package tests

import (
	"crypto/sha256"
	"fmt"
	"github.com/google/uuid"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const (
	KiB          int64 = 1024
	MiB                = 1024 * KiB
	GiB                = 1024 * MiB
	Size               = 5 * GiB
	BlockSize          = 512 * KiB
	NumBocks           = Size / BlockSize
	ExpectedHash       = "7f06c62352aebd8125b2a1841e2b9e1ffcbed602f381c3dcb3200200e383d1d5"
)

func CreateSparseFile() (string, error) {
	filename := fmt.Sprintf(FilenameTemplate, uuid.New().String())
	fd, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer fd.Close()
	if _, err = fd.Seek(Size-1, 0); err != nil {
		return "", err
	}
	if _, err = fd.Write([]byte{0xff}); err != nil {
		return "", err
	}
	return filename, nil
}

func AssertFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	hash := fmt.Sprintf("%x", h.Sum(nil))
	if hash != ExpectedHash {
		return fmt.Errorf("file is corrupt: %s", hash)
	}

	return nil
}

func DDOverwrite(filename string) error {
	args := []string{
		"if=/dev/zero",
		"of=" + filename,
		fmt.Sprintf("bs=%d", BlockSize),
		fmt.Sprintf("count=%d", NumBocks),
		"conv=notrunc",
	}
	cmd := exec.Command("dd", args...)
	return cmd.Run()
}

func ControlOverwrite(filename string) error {
	fd, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer fd.Close()

	reader := NewZeroReader(Size)
	io.Copy(fd, reader)

	return nil
}

func ODirectOverwrite(filename string) error {
	fd, err := os.OpenFile(filename, os.O_RDWR|syscall.O_DIRECT, 0666)
	if err != nil {
		return err
	}
	defer fd.Close()

	reader := NewZeroReader(Size)
	io.Copy(fd, reader)

	return nil
}

func IOUringOverwrite(filename string) error {
	start := time.Now()
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	iour, err := NewIOURingWriter(f)
	if err != nil {
		return err
	}
	defer iour.Close()

	reader := NewZeroReader(Size)
	io.Copy(iour, reader)

	fmt.Printf("writes finished in %d ms.\n", time.Since(start).Milliseconds())
	iour.Wait()
	fmt.Printf("waiting finished in %d ms.\n", time.Since(start).Milliseconds())

	return nil
}

func IOUringDirectOverwrite(filename string) error {
	start := time.Now()
	f, err := os.OpenFile(filename, os.O_RDWR|syscall.O_DIRECT, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	iour, err := NewIOURingWriter(f)
	if err != nil {
		return err
	}
	defer iour.Close()

	reader := NewZeroReader(Size)
	io.Copy(iour, reader)

	fmt.Printf("writes finished in %d ms.\n", time.Since(start).Milliseconds())
	iour.Wait()
	fmt.Printf("waiting finished in %d ms.\n", time.Since(start).Milliseconds())

	return nil
}
