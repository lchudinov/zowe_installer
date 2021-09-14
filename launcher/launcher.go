package launcher

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

type Launcher struct {
	rootDir          string
	instanceDir      string
	haInstanceId     string
	launchComponents []string
	components       map[string]*Component
}

func New() *Launcher {
	var launcher Launcher
	launcher.components = make(map[string]*Component)
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
	if err := launcher.prepareInstance(); err != nil {
		return errors.Wrapf(err, "failed to prepare instance")
	}
	if err := launcher.getLaunchComponents(); err != nil {
		return errors.Wrapf(err, "failed to find launch components")
	}
	if err := launcher.initComponents(); err != nil {
		return errors.Wrapf(err, "failed to init components")
	}
	if err := launcher.startComponents(); err != nil {
		return errors.Wrap(err, "failed to start components")
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
	list := strings.TrimSuffix(string(output), "\n")
	list = strings.TrimSuffix(list, ",")
	launcher.launchComponents = strings.Split(list, ",")
	if len(launcher.launchComponents) == 0 {
		return errors.New("no launch components")
	}
	log.Printf("LAUNCH COMPONENTS = %s", strings.Join(launcher.launchComponents, ","))
	return nil
}

func (launcher *Launcher) prepareInstance() error {
	log.Printf("preparing instance...")
	script := filepath.Join(launcher.rootDir, "bin", "internal", "prepare-instance.sh")
	cmd := exec.Command(script, "-c", launcher.instanceDir, "-r", launcher.rootDir, "-i", launcher.haInstanceId)
	cmd.Dir = launcher.instanceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to run %s", script)
	}
	log.Printf("instance prepared")
	return nil
}

func (launcher *Launcher) initComponents() error {
	for _, name := range launcher.launchComponents {
		launcher.components[name] = NewComponent(name)
	}
	return nil
}

func (launcher *Launcher) startComponents() error {
	for name, comp := range launcher.components {
		if name != "zss" && name != "app-server" {
			continue
		}
		if err := launcher.startComponent(comp); err != nil {
			return errors.Wrapf(err, "failed to start component %s", name)
		}
	}
	return nil
}

func (launcher *Launcher) startComponent(comp *Component) error {
	log.Printf("starting component %s...", comp.Name)
	script := filepath.Join(launcher.rootDir, "bin", "internal", "start-component.sh")
	cmd := exec.Command(script, "-c", launcher.instanceDir, "-r", launcher.rootDir, "-i", launcher.haInstanceId, "-o", comp.Name)
	cmd.Dir = launcher.instanceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	var attr syscall.SysProcAttr
	attr.Setpgid = true
	cmd.SysProcAttr = &attr
	if err := cmd.Start(); err != nil {
		return errors.Wrapf(err, "failed to run component %s", comp.Name)
	}
	comp.cmd = cmd
	log.Printf("component %s started", comp.Name)
	return nil
}
