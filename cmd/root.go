package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/ashcrow/pivot/types"
	"github.com/ashcrow/pivot/utils"
	"github.com/spf13/cobra"
)

// flag storage
var touchIfChanged bool
var keep bool
var reboot bool
var container string

// RootCmd houses the cobra config for the main command
var RootCmd = &cobra.Command{
	Use:   "pivot",
	Short: "",
	//	Long: ``,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("An image name must be provided")
		}
		return nil
	},
	Run: Execute,
}

// Executes upon import
func init() {
	RootCmd.PersistentFlags().BoolVarP(&touchIfChanged, "touch-if-changed", "t", false, "if changed, touch a file")
	RootCmd.PersistentFlags().BoolVarP(&keep, "keep", "k", false, "Do not remove container image")
	RootCmd.PersistentFlags().BoolVarP(&reboot, "reboot", "r", false, "reboot if changed")
}

// podmanRemove kills and removes a container
func podmanRemove(cid string) {
	exec.Command("podman", "kill", cid).Run()
	exec.Command("podman", "rm", "-f", cid).Run()
}

// Execute runs the command
func Execute(cmd *cobra.Command, args []string) {
	container := args[0]
	previousPivot := ""
	if _, err := os.Stat(types.PivotDonePath); err == nil {
		content, err := ioutil.ReadFile(types.PivotDonePath)
		if err != nil {
			utils.Fatal(fmt.Sprintf("Unable to read %s: %s", types.PivotDonePath, err))
		}
		previousPivot = strings.TrimSpace(string(content))
		fmt.Printf("Previous pivot: %s\n", previousPivot)
	}

	// Use skopeo to find the sha256, so we can refer to it reliably
	output := utils.RunGetOut("skopeo", "inspect", fmt.Sprintf("docker://%s", container))

	var imagedata types.ImageInspection
	json.Unmarshal([]byte(output), &imagedata)
	imgid := fmt.Sprintf("%s@%s", container, imagedata.Digest)

	if previousPivot == imgid {
		fmt.Printf("Already pivoted to: %s\n", imgid)
		os.Exit(0)
	}

	// Pull the image
	utils.Run("podman", "pull", imgid)
	fmt.Printf("Pivoting to: %s\n", imgid)

	//Clean up a previous container
	podmanRemove(types.PivotName)
	// `podman mount` wants a running container, so let's make a dummy one
	cid := utils.RunGetOutln("podman", "run", "-d", "--name",
		types.PivotName, "--entrypoint", "sleep", imgid, "infinity")
	// Use the container ID to find its mount point
	mnt := utils.RunGetOutln("podman", "mount", cid)
	os.Chdir(mnt)

	// List all refs from the OSTree repository embedded in the container
	refsCombined := utils.RunGetOut("ostree", "--repo=srv/tree/repo", "refs")
	refs := strings.Split(strings.TrimSpace(refsCombined), "\n")
	rlen := len(refs)
	// Today, we only support one ref.  Down the line we may do multiple.
	if rlen != 1 {
		utils.Fatal(fmt.Sprintf("Found %d refs, expected exactly 1", rlen))
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
	err := ioutil.WriteFile(types.PivotDonePath, []byte(fmt.Sprintf("%s\n", imgid)), 0644)
	if err != nil {
		utils.Fatal(fmt.Sprintf("Unable to write the new imgid of %s to %s", imgid, types.PivotDonePath))
	}

	// Kill our dummy container
	podmanRemove(types.PivotName)

	// By default, delete the image.
	if !keep {
		utils.Run("podman", "rmi", imgid)
	}
}
