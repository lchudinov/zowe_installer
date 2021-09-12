package launcher

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

type Launcher struct {
	rootDir     string
	instanceDir string
}

func New() *Launcher {
	var launcher Launcher
	return &launcher
}

func (launcher *Launcher) Run(instanceDir string, haInstanceId string) error {
	launcher.instanceDir = instanceDir
	if _, err := os.Stat(instanceDir); err != nil {
		return errors.Wrapf(err, "failed to find INSTANCE_DIR %s", instanceDir)
	}
	if err := launcher.findRootDir(); err != nil {
		return errors.Wrapf(err, "failed to find ROOT_DIR")
	}
	return nil
}

func (launcher *Launcher) findRootDir() error {
	command := fmt.Sprintf(". %s/bin/internal/read-essential-vars.sh && echo $ROOT_DIR", launcher.instanceDir)
	cmd := exec.Command("/bin/sh -c", command)
	buf, err := cmd.Output()
	if err != nil {
		return errors.Wrapf(err, "failed to run %s", command)
	}
	rootDir := string(buf)
	if _, err := os.Stat(rootDir); err != nil {
		return errors.Wrapf(err, "failed to find ROOT_DIR %s", rootDir)
	}
	launcher.rootDir = rootDir
	log.Printf("ROOT_DIR = %s\n", rootDir)
	return nil
}
