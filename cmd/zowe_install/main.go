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

type ZoweInstaller struct {
	paxURL      string
	paxFileName string
	dir         string
	rootDir     string
	instanceDir string
}

func New() *ZoweInstaller {
	installer := ZoweInstaller{}
	return &installer
}

func (installer *ZoweInstaller) Install(paxURL string) error {
	if err := installer.PrepareInstallation(paxURL); err != nil {
		return err
	}
	if err := installer.DownloadPax(); err != nil {
		return err
	}
	if err := installer.ExtractPax(); err != nil {
		return nil
	}
	if err := installer.InstallPax(); err != nil {
		return err
	}
	if err := installer.InitInstance(); err != nil {
		return nil
	}
	return nil
}

func (installer *ZoweInstaller) PrepareInstallation(paxURL string) error {
	url, err := url.Parse(paxURL)
	if err != nil {
		return errors.Wrapf(err, "failed to parse PAX URL %s\n", paxURL)
	}
	paxFile := path.Base(url.Path)
	ext := path.Ext(paxFile)
	dir := strings.TrimSuffix(paxFile, ext)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrapf(err, "failed to get user home dir")
	}
	installer.dir = filepath.Join(homeDir, dir)
	if err := os.RemoveAll(installer.dir); err != nil {
		return errors.Wrapf(err, "failed to cleanup installation dir")
	}
	if err := os.MkdirAll(installer.dir, 0755); err != nil {
		return errors.Wrapf(err, "failed to create directory for installation")
	}
	installer.paxFileName = filepath.Join(installer.dir, paxFile)
	installer.paxURL = paxURL
	return nil
}

func (installer *ZoweInstaller) DownloadPax() (err error) {
	out, err := os.Create(installer.paxFileName)
	if err != nil {
		err = errors.Wrapf(err, "failed to create file for pax")
		return
	}
	defer out.Close()
	resp, err := http.Get(installer.paxURL)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.Errorf("bad status code - %d", resp.StatusCode)
		return
	}
	defer resp.Body.Close()
	var counter ByteCounter
	_, err = io.Copy(out, io.TeeReader(resp.Body, &counter))
	if err != nil {
		errors.Wrapf(err, "failed to read response body")
		return
	}
	return
}

func (installer *ZoweInstaller) ExtractPax() error {
	pax := installer.paxFileName
	workDir := filepath.Dir(pax)
	cmd := exec.Command("pax", "-rvf", pax)
	cmd.Dir = workDir
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "error unpacking %s", pax)
	}
	return nil
}

func (installer *ZoweInstaller) InstallPax() error {
	dir := installer.dir
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
		return errors.Wrapf(err, "installation failed")
	}
	return nil
}

func (installer *ZoweInstaller) InitInstance() error {
	dir := installer.dir
	instanceDir := filepath.Join(dir, "instance")
	if err := os.Mkdir(instanceDir, 0755); err != nil {
		return errors.Wrapf(err, "failed to create instance dir %s", instanceDir)
	}
	log.Printf("Configuring instance..")
	rootDir := filepath.Join(dir, "root")
	userInfo, err := user.Current()
	if err != nil {
		return errors.Wrapf(err, "failed to get current user")
	}
	rootBinDir := filepath.Join(rootDir, "bin")
	groupInfo, err := user.LookupGroupId(userInfo.Gid)
	if err != nil {
		return errors.Wrapf(err, "failed to get group name for user %s", userInfo.Name)
	}
	cmd := exec.Command("./zowe-configure-instance.sh", "-c", instanceDir, "-g", groupInfo.Name)
	cmd.Dir = rootBinDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to configure instance")
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <Zowe PAX URL>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	paxURL := os.Args[1]
	installer := New()
	if err := installer.Install(paxURL); err != nil {
		log.Fatalf("failed to install Zowe pax %s: %v", paxURL, err)
	}
}
