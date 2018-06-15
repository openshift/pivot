package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	// TODO: update to official location
	"github.com/ashcrow/pivot/cmd"
)

// The following are passed in at build time
var commitHash string
var version string

// main is the entry point for the command
func main() {
	fmt.Printf("pivot version %s (%s)\n", version, commitHash)

	flag.CommandLine.Parse([]string{})
	pflag.Set("logtostderr", "true")

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
