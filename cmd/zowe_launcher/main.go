package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/lchudinov/zowe_installer/launcher"
)

func setupInterrutHandler(launcher *launcher.Launcher) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			launcher.Stop()
			break
		}
	}()
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s INSTANCE_DIR HA_INSTANCE_ID\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	instanceDir := os.Args[1]
	haInstanceId := os.Args[2]
	launcher := launcher.New()
	if err := launcher.Run(instanceDir, haInstanceId); err != nil {
		log.Fatalf("failed to run Zowe: %v", err)
	}
	setupInterrutHandler(launcher)
	go func() {
		if err := launcher.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	launcher.Wait()
}
