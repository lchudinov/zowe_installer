package launcher

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type Launcher struct {
	rootDir          string
	instanceDir      string
	haInstanceId     string
	launchComponents []string
}

func New() *Launcher {
	var launcher Launcher
	return &launcher
}

func (launcher *Launcher) Run(instanceDir string, haInstanceId string) error {
	launcher.instanceDir = instanceDir
	launcher.haInstanceId = haInstanceId
	if _, err := os.Stat(instanceDir); err != nil {
		return errors.Wrapf(err, "failed to find INSTANCE_DIR %s", instanceDir)
	}
	if err := launcher.findRootDir(); err != nil {
		return errors.Wrapf(err, "failed to find ROOT_DIR")
	}
	if err := launcher.getLaunchComponents(); err != nil {
		return errors.Wrapf(err, "failed to find launch components")
	}
	return nil
}

func (launcher *Launcher) findRootDir() error {
	command := fmt.Sprintf(". %s/bin/internal/read-essential-vars.sh && echo $ROOT_DIR", launcher.instanceDir)
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("INSTANCE_DIR=%s", launcher.instanceDir))
	buf, err := cmd.Output()
	if err != nil {
		return errors.Wrapf(err, "failed to run %s", command)
	}
	rootDir := strings.TrimSuffix(string(buf), "\n")
	if _, err := os.Stat(rootDir); err != nil {
		return errors.Wrapf(err, "failed to find ROOT_DIR %s", rootDir)
	}
	launcher.rootDir = rootDir
	log.Printf("ROOT_DIR = %s\n", rootDir)
	return nil
}

func (launcher *Launcher) getLaunchComponents() error {
	script := filepath.Join(launcher.rootDir, "bin", "internal", "get-launch-components.sh")
	cmd := exec.Command(script, "-c", launcher.instanceDir, "-r", launcher.rootDir, "-i", launcher.haInstanceId)
	output, err := cmd.Output()
	if err != nil {
		return errors.Wrapf(err, "failed to run %s", script)
	}
	launcher.launchComponents = strings.Split(strings.TrimSuffix(string(output), "\n"), ",")
	if len(launcher.launchComponents) == 0 {
		return errors.New("no launch components")
	}
	log.Printf("LAUNCH COMPONENTS: %s", strings.Join(launcher.launchComponents, ","))
	return nil
}
