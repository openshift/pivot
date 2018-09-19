package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/openshift/pivot/cmd"
	"github.com/spf13/pflag"
)

// The following are passed in at build time
var commitHash string
var version string

// showHeader generates and prints the program header line
func showHeader() {
	header := fmt.Sprintf("pivot version %s", version)
	// If we have a commit hash then add it to the program header
	if commitHash != "" {
		header = fmt.Sprintf("%s (%s)", header, commitHash)
	}
	fmt.Println(header)
}

// main is the entry point for the command
func main() {
	showHeader()
	flag.CommandLine.Parse([]string{})
	pflag.Set("logtostderr", "true")

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
