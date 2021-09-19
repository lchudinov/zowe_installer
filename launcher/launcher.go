package launcher

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/pkg/errors"
)

type Launcher struct {
	rootDir          string
	instanceDir      string
	haInstanceId     string
	launchComponents []string
	components       map[string]*Component
	wg               *sync.WaitGroup
	env              []string
	output           bytes.Buffer
	*http.Server
}

func New() *Launcher {
	var launcher Launcher
	launcher.components = make(map[string]*Component)
	launcher.wg = new(sync.WaitGroup)
	launcher.Server = &http.Server{Addr: ":8053", Handler: &launcher}
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
	launcher.env = launcher.makeEnvironment()
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

func (launcher *Launcher) Wait() {
	launcher.wg.Wait()
	log.Printf("components stopped")
}

func (launcher *Launcher) findRootDir() error {
	command := fmt.Sprintf(". %s/bin/internal/read-essential-vars.sh && echo $ROOT_DIR", launcher.instanceDir)
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("INSTANCE_DIR=%s", launcher.instanceDir))
	cmd.Dir = launcher.instanceDir
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

func (launcher *Launcher) makeEnvironment() []string {
	env := os.Environ()
	env = append(env, fmt.Sprintf("INSTANCE_DIR=%s", launcher.instanceDir))
	env = append(env, fmt.Sprintf("ROOT_DIR=%s", launcher.rootDir))
	return env
}

func (launcher *Launcher) getLaunchComponents() error {
	script := filepath.Join(launcher.rootDir, "bin", "internal", "get-launch-components.sh")
	cmd := exec.Command(script, "-c", launcher.instanceDir, "-r", launcher.rootDir, "-i", launcher.haInstanceId)
	cmd.Env = launcher.env
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
	cmd.Env = launcher.env
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
	cmd.Env = launcher.env
	cmd.Dir = launcher.instanceDir
	cmd.Stdout = io.MultiWriter(os.Stdout, &comp.output)
	cmd.Stderr = io.MultiWriter(os.Stderr, &comp.output)
	var attr syscall.SysProcAttr
	attr.Setpgid = true
	cmd.SysProcAttr = &attr
	if err := cmd.Start(); err != nil {
		return errors.Wrapf(err, "failed to run component %s", comp.Name)
	}
	comp.cmd = cmd
	log.Printf("component %s started", comp.Name)
	launcher.wg.Add(1)
	go func() {
		defer launcher.wg.Done()
		if _, err := cmd.Process.Wait(); err != nil {
			log.Printf("component stopped with error: %v", err)
		}
	}()
	return nil
}

func (launcher *Launcher) stopComponent(comp *Component) error {
	if comp.cmd != nil {
		log.Printf("stopping component %s...", comp.Name)
		if err := syscall.Kill(-comp.cmd.Process.Pid, syscall.SIGTERM); err != nil {
			return errors.Wrapf(err, "failed to kill component %s", comp.Name)
		}
	}
	return nil
}

func (launcher *Launcher) StopComponents() {
	log.Printf("stopping components...")
	for name, comp := range launcher.components {
		if err := launcher.stopComponent(comp); err != nil {
			log.Printf("failed to stop component %s: %v", name, err)
		}
	}
}

func (launcher *Launcher) Stop() {
	launcher.Shutdown(context.Background())
	launcher.StopComponents()
}

func (launcher *Launcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, comp := range launcher.components {
		fmt.Fprintf(w, "---\n%s\n----%s\n", comp.Name, string(comp.output.Bytes()))
	}
}
