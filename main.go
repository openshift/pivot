package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	// TODO: update to official location
	"github.com/ashcrow/pivot/types"
	"github.com/ashcrow/pivot/utils"
)

const (
	// PivotDonePath is the path to the file used to denote pivot work
	PivotDonePath = "/etc/os-container-pivot.stamp"
	// PivotName is literally the name of the new pivot
	PivotName = "ostree-container-pivot"
)

// The following are passed in at build time
var commitHash string
var version string

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

// podmanRemove kills and removes a container
func podmanRemove(cid string) {
	exec.Command("podman", "kill", cid)
	exec.Command("podman", "rm", "-f", cid)
}

// TODO: much of the main function should be broken up into functions and
// moved to other packages
// main is the entry point for the command
func main() {
	// flags
	var touchIfChanged bool
	var keep bool
	var reboot bool
	var container string

	flag.BoolVar(&touchIfChanged, "touch-if-changed", false, "if changed, touch a file")
	flag.BoolVar(&keep, "keep", false, "Do not remove container image")
	flag.BoolVar(&reboot, "reboot", false, "reboot if changed")
	flag.Usage = func() {
		fmt.Printf("pivot: version %s (%s)\n", version, commitHash)
		fmt.Printf("Usage:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Parse the flags into variables

	arguments := flag.Args()
	if len(arguments) != 1 {
		fatal("Only one argument name may be provided")
	}
	container = arguments[0]

	// Logic

	// This file holds the imageid (container name + sha256) that
	// we successfully used for a previous pivot.
	previousPivot := ""
	if _, err := os.Stat(PivotDonePath); err == nil {
		content, err := ioutil.ReadFile(PivotDonePath)
		if err != nil {
			fatal(fmt.Sprintf("Unable to read %s: %s", PivotDonePath, err))
		}
		previousPivot = strings.TrimSpace(string(content))
	}
	fmt.Printf("Previous pivot: %s\n", previousPivot)

	// Use skopeo to find the sha256, so we can refer to it reliably
	output, err := exec.Command(
		"skopeo", "inspect", fmt.Sprintf("docker://%s", container)).Output()
	if err != nil {
		fatal(fmt.Sprintf("Unable to run skopeo: %s", err))
	}

	var imagedata types.ImageInspection
	json.Unmarshal(output, &imagedata)
	imgid := fmt.Sprintf("%s@%s", container, imagedata.Digest)

	if previousPivot == imgid {
		fmt.Printf("Already pivoted to: %s", imgid)
		os.Exit(0)
	}

	// Pull the image
	exec.Command("podman", "pull", imgid)
	fmt.Printf("Pivoting to: %s", imgid)

	//Clean up a previous container
	podmanRemove(PivotName)
	// `podman mount` wants a running container, so let's make a dummy one
	cid := utils.RunGetOutln("podman", "run", "-d", "--name",
		PivotName, "--entrypoint", "sleep", imgid, "infinity")
	// Use the container ID to find its mount point
	mnt := utils.RunGetOutln("podman", "mount", cid)
	os.Chdir(mnt)

	// List all refs from the OSTree repository embedded in the container
	refsCombined := utils.Run("ostree", "--repo=srv/tree/repo", "refs")
	refs := strings.Split(refsCombined, "\n")
	rlen := len(refs)
	// Today, we only support one ref.  Down the line we may do multiple.
	if rlen != 1 {
		fatal(fmt.Sprintf("Found %d refs, expected exactly 1", rlen))
	}
	targetRef := refs[0]
	// Find the concrete OSTree commit
	rev := utils.RunGetOutln("ostree", "--repo=srv/tree/repo", "rev-parse", targetRef)

	// Use pull-local to extract the data into the system repo; this is *significantly*
	// faster than talking to the container over HTTP.
	utils.Run("ostree", "pull-local", "srv/tree/repo", rev)

	// The leading ':' here means "no remote".  See also
	// https://github.com/projectatomic/rpm-ostree/pull/1396
	utils.Run("rpm-ostree", "rebase", fmt.Sprintf(":%s", rev))

	// Done!  Write our stamp file.  TODO: Teach rpm-ostree how to encode
	// this data in the origin.
	err = ioutil.WriteFile(PivotDonePath, []byte(fmt.Sprintf("%s\n", imgid)), 0644)
	if err != nil {
		fatal(fmt.Sprintf("Unable to write the new imgid of %s to %s", imgid, PivotDonePath))
	}

	// Kill our dummy container
	podmanRemove(PivotName)

	// By default, delete the image.
	if keep {
		utils.Run("podman", "rmi", imgid)
	}
}
