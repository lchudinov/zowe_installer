package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/lchudinov/zowe_installer/launcher"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Usage: %s INSTANCE_DIR HA_INSTANCE_ID\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	instanceDir := os.Args[1]
	haInstanceId := os.Args[2]
	launcher := launcher.New()
	if err := launcher.Run(instanceDir, haInstanceId); err != nil {
		log.Fatalf("failed to run Zowe: %v", err)
	}
}
