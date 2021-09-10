package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
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

func ExtractPax(pax string) error {
	workDir := filepath.Dir(pax)
	cmd := exec.Command("pax", "-rvf", pax)
	cmd.Dir = workDir
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "failed to extract pax")
	}
	return nil
}

/*
folder="${dir:0:11}"

if [ ! -d "${folder}" ]; then
  echo "${folder} not found"
  exit 1
fi

cd "${folder}/install" || exit 1

export ROOT_DIR="${HOME}/${dir}/root"
./zowe-install.sh -i "${ROOT_DIR}" -h "${USER}" || exit 1

*/

func InstallPax(pax string) error {
	dir := filepath.Dir(pax)
	folder := filepath.Base(dir)[0:11]
	installDir := filepath.Join(dir, folder, "install")
	if _, err := os.Stat(installDir); err != nil {
		return errors.Wrapf(err, "failed to find install dir %s: %v", installDir)
	}
	rootDir := filepath.Join(dir, "root")
	user, err := user.Current()
	if err != nil {
		return errors.Wrapf(err, "failed to get current user")
	}
	cmd := exec.Command("./zowe-install.sh", "-i", rootDir, "-h", user.Username)
	cmd.Dir = installDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to install zowe pax")
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to get user home dir: %v\n", err)
	}

	targetDir := filepath.Join(homeDir, dir)
	if err := os.RemoveAll(targetDir); err != nil {
		log.Fatalf("failed to cleanup installation dir: %v", err)
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Fatalf("failed to create directory for installation: %v", err)
	}
	targetPax := filepath.Join(targetDir, paxFile)

	if err := DownloadPax(paxURL, targetPax); err != nil {
		log.Fatal(err)
	}
	if err := ExtractPax(targetPax); err != nil {
		log.Fatal(err)
	}
	if err := InstallPax(targetPax); err != nil {
		log.Fatal(err)
	}
}
