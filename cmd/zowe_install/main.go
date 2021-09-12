package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/lchudinov/zowe_installer/installer"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <Zowe PAX URL>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	paxURL := os.Args[1]
	installer := installer.New()
	if err := installer.Install(paxURL); err != nil {
		log.Fatalf("failed to install Zowe pax %s: %v", paxURL, err)
	}
}
