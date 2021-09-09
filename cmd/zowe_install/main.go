package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/pkg/errors"
)

type ByteCounter struct {
	Total uint64
}

func (bc *ByteCounter) Write(p []byte) (n int, err error) {
	len := len(p)
	bc.Total += uint64(len)
	bc.PrintProgress()
	return n, nil
}

func (bc *ByteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(bc.Total))
}

func DownloadPax(url string, file string) error {
	out, err := os.Create(file)
	if err != nil {
		return errors.Wrapf(err, "failed to create file")
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "failed to download file")
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("failed to download file")
	}
	defer resp.Body.Close()
	var counter ByteCounter
	_, err = io.Copy(out, io.TeeReader(resp.Body, &counter))
	if err != nil {
		return errors.Wrapf(err, "failed to download file")
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <Zowe PAX URL>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	paxURL := os.Args[1]
	url, err := url.Parse(paxURL)
	if err != nil {
		log.Fatalf("failed to parse PAX URL %s: %v\n", paxURL, err)
	}

	paxFile := path.Base(url.Path)
	ext := path.Ext(paxFile)
	dir := strings.TrimSuffix(paxFile, ext)
	fmt.Println(paxFile, dir, ext)

	err = DownloadPax(paxURL, paxFile)
	if err != nil {
		log.Fatal(err)
	}
}
