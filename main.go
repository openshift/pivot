package main

import (
	"fmt"
	"os"

	// TODO: update to official location
	"github.com/ashcrow/pivot/cmd"
)

// The following are passed in at build time
var commitHash string
var version string

// main is the entry point for the command
func main() {
	fmt.Printf("pivot version %s (%s)\n", version, commitHash)
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
