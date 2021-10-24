package launcher

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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
	output           *Buffer
	router           *mux.Router
	*log.Logger
	*http.Server
}

func New() *Launcher {
	var launcher Launcher
	launcher.components = make(map[string]*Component)
	launcher.wg = new(sync.WaitGroup)
	launcher.router = launcher.makeRouter()
	credentials := handlers.AllowCredentials()
	methods := handlers.AllowedMethods([]string{"POST,GET,DELETE,PUT"})
	origins := handlers.AllowedOrigins([]string{"*"})
	launcher.output = NewBuffer()
	launcher.Logger = log.New(launcher.output, "", 0)
	launcher.Server = &http.Server{
		Addr:    ":8053",
		Handler: handlers.CORS(credentials, methods, origins)(launcher.router),
	}
	return &launcher
}

func (launcher *Launcher) makeRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/components", launcher.handleComponents).Methods("GET")
	router.HandleFunc("/log", launcher.handleLog).Methods("GET")
	router.HandleFunc("/components/{comp}/log", launcher.handleComponentLog).Methods("GET")
	router.HandleFunc("/components/{comp}/stop", launcher.handleComponentStop).Methods("POST")
	router.HandleFunc("/components/{comp}/start", launcher.handleComponentStart).Methods("POST")
	return router
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
	launcher.Printf("components stopped")
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
	launcher.Printf("ROOT_DIR = %s\n", rootDir)
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
	launcher.Printf("LAUNCH COMPONENTS = %s", strings.Join(launcher.launchComponents, ","))
	return nil
}

func (launcher *Launcher) prepareInstance() error {
	launcher.Printf("preparing instance...")
	script := filepath.Join(launcher.rootDir, "bin", "internal", "prepare-instance.sh")
	cmd := exec.Command(script, "-c", launcher.instanceDir, "-r", launcher.rootDir, "-i", launcher.haInstanceId)
	cmd.Env = launcher.env
	cmd.Dir = launcher.instanceDir
	cmd.Stdout = io.MultiWriter(os.Stdout, launcher.output)
	cmd.Stderr = io.MultiWriter(os.Stderr, launcher.output)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to run %s", script)
	}
	launcher.Printf("instance prepared")
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
		if err := launcher.startComponent(comp); err != nil {
			return errors.Wrapf(err, "failed to start component %s", name)
		}
	}
	return nil
}

func (launcher *Launcher) startComponent(comp *Component) error {
	launcher.Printf("starting component %s...", comp.Name)
	script := filepath.Join(launcher.rootDir, "bin", "internal", "start-component.sh")
	cmd := exec.Command(script, "-c", launcher.instanceDir, "-r", launcher.rootDir, "-i", launcher.haInstanceId, "-o", comp.Name)
	cmd.Env = launcher.env
	cmd.Dir = launcher.instanceDir
	cmd.Stdout = io.MultiWriter(os.Stdout, comp.output)
	cmd.Stderr = io.MultiWriter(os.Stderr, comp.output)
	cmd.SysProcAttr = getSysProcAttr()
	if err := cmd.Start(); err != nil {
		return errors.Wrapf(err, "failed to run component %s", comp.Name)
	}
	comp.cmd = cmd
	launcher.Printf("component %s started", comp.Name)
	launcher.wg.Add(1)
	go func() {
		defer launcher.wg.Done()
		if _, err := cmd.Process.Wait(); err != nil {
			launcher.Printf("component %s stopped with error: %v", comp.Name, err)
		} else {
			launcher.Printf("component %s stopped", comp.Name)
		}
		comp.cmd = nil
	}()
	return nil
}

func (launcher *Launcher) stopComponent(comp *Component) error {
	if comp.cmd != nil {
		launcher.Printf("stopping component %s...", comp.Name)
		if err := kill(comp.cmd.Process.Pid); err != nil {
			return errors.Wrapf(err, "failed to kill component %s", comp.Name)
		}
	}
	return nil
}

func (launcher *Launcher) StopComponents() {
	launcher.Printf("stopping components...")
	for name, comp := range launcher.components {
		if err := launcher.stopComponent(comp); err != nil {
			launcher.Printf("failed to stop component %s: %v", name, err)
		}
	}
}

func (launcher *Launcher) Stop() {
	launcher.Shutdown(context.Background())
	launcher.StopComponents()
}

func (launcher *Launcher) handleComponents(w http.ResponseWriter, r *http.Request) {
	var comps []*Component
	for _, comp := range launcher.components {
		comps = append(comps, comp)
	}
	data, _ := json.Marshal(comps)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (launcher *Launcher) handleComponentLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["comp"]
	var level LogLevel = LogLevelAny
	var err error
	if val := r.FormValue("level"); val != "" {
		if level, err = parseLogLevel(val); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			writeError(w, "Unknown log level '%s'", val)
			return
		}
	}
	if comp, ok := launcher.components[name]; ok {
		var lines []string
		scanner := bufio.NewScanner(strings.NewReader(comp.output.String()))
		for scanner.Scan() {
			line := stripEscapeSeqs(scanner.Text())
			lineLevel := getLogLevel(line)
			if lineLevel <= level {
				lines = append(lines, line)
			}
		}
		if err := scanner.Err(); err != nil {
			launcher.Printf("error reading componet output, %v", err)
		}
		data, _ := json.Marshal(lines)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	} else {
		w.WriteHeader(http.StatusNotFound)
		writeError(w, "Component '%s' not found", name)
	}
}

func (launcher *Launcher) handleComponentStop(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["comp"]
	if comp, ok := launcher.components[name]; ok {
		if err := launcher.stopComponent(comp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeError(w, "Couldn't stop component '%s': %v", name, err)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		writeError(w, "Component '%s' not found", name)
	}
}
func (launcher *Launcher) handleComponentStart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["comp"]
	if comp, ok := launcher.components[name]; ok {
		if err := launcher.startComponent(comp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeError(w, "Couldn't start component '%s': %v", name, err)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		writeError(w, "Component '%s' not found", name)
	}
}
func (launcher *Launcher) handleLog(w http.ResponseWriter, r *http.Request) {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(launcher.output.String()))
	for scanner.Scan() {
		line := stripEscapeSeqs(scanner.Text())
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		launcher.Printf("error reading componet output, %v", err)
	}
	data, _ := json.Marshal(lines)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func writeError(w http.ResponseWriter, format string, a ...interface{}) {
	type errorMessage struct {
		Message string `json:"error"`
	}
	var msg errorMessage
	msg.Message = fmt.Sprintf(format, a...)
	data, _ := json.Marshal(msg)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
