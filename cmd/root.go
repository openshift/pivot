package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ashcrow/pivot/types"
	"github.com/ashcrow/pivot/utils"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// flag storage
var keep bool
var reboot bool
var container string

// RootCmd houses the cobra config for the main command
var RootCmd = &cobra.Command{
	Use:   "pivot",
	Short: "Allows moving from one OSTree deployment to another",
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
	RootCmd.PersistentFlags().BoolVarP(&keep, "keep", "k", false, "Do not remove container image")
	RootCmd.PersistentFlags().BoolVarP(&reboot, "reboot", "r", false, "reboot if changed")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

// podmanRemove kills and removes a container
func podmanRemove(cid string) {
	exec.Command("podman", "kill", cid).Run()
	exec.Command("podman", "rm", "-f", cid).Run()
}

// getDefaultDeployment uses rpm-ostree status --json to get the current deployment
func getDefaultDeployment() types.RpmOstreeDeployment {
	// use --status for now, we can switch to D-Bus if we need more info
	var rosState types.RpmOstreeState
	output := utils.RunGetOut("rpm-ostree", "status", "--json")
	if err := json.Unmarshal([]byte(output), &rosState); err != nil {
		glog.Fatalf("Failed to parse `rpm-ostree status --json` output: %v", err)
	}

	// just make it a hard error if we somehow don't have any deployments
	if len(rosState.Deployments) == 0 {
		glog.Fatalf("Not currently booted in a deployment")
	}

	return rosState.Deployments[0]
}

// Execute runs the command
func Execute(cmd *cobra.Command, args []string) {
	container := args[0]
	defaultDeployment := getDefaultDeployment()

	previousPivot := ""
	if len(defaultDeployment.CustomOrigin) > 0 {
		if strings.HasPrefix(defaultDeployment.CustomOrigin[0], "pivot://") {
			previousPivot = defaultDeployment.CustomOrigin[0][len("pivot://"):]
			glog.Infof("Previous pivot: %s\n", previousPivot)
		}
	}

	// Use skopeo to canonicalize to $name@$digest, so we can refer to it reliably
	output := utils.RunGetOut("skopeo", "inspect", fmt.Sprintf("docker://%s", container))

	var imagedata types.ImageInspection
	json.Unmarshal([]byte(output), &imagedata)
	imgid := fmt.Sprintf("%s@%s", imagedata.Name, imagedata.Digest)

	if previousPivot == imgid {
		glog.Info("Already at target pivot; exiting...\n")
		return
	}

	// Pull the image
	utils.Run("podman", "pull", imgid)
	glog.Infof("Pivoting to: %s\n", imgid)

	// Clean up a previous container
	podmanRemove(types.PivotName)

	// `podman mount` wants a running container, so let's make a dummy one
	cid := utils.RunGetOutln("podman", "run", "-d", "--name",
		types.PivotName, "--entrypoint", "sleep", imgid, "infinity")
	// Use the container ID to find its mount point
	mnt := utils.RunGetOutln("podman", "mount", cid)
	os.Chdir(mnt)

	// List all refs from the OSTree repository embedded in the container
	refsCombined := utils.RunGetOut("ostree", "--repo=srv/repo", "refs")
	refs := strings.Split(strings.TrimSpace(refsCombined), "\n")
	rlen := len(refs)
	// Today, we only support one ref.  Down the line we may do multiple.
	if rlen != 1 {
		glog.Fatalf("Found %d refs, expected exactly 1", rlen)
	}
	targetRef := refs[0]
	// Find the concrete OSTree commit
	rev := utils.RunGetOutln("ostree", "--repo=srv/repo", "rev-parse", targetRef)

	// Use pull-local to extract the data into the system repo; this is *significantly*
	// faster than talking to the container over HTTP.
	utils.Run("ostree", "pull-local", "srv/repo", rev)

	// This will be what will be displayed in `rpm-ostree status` as the "origin spec"
	customURL := fmt.Sprintf("pivot://%s", imgid)

	// The leading ':' here means "no remote".  See also
	// https://github.com/projectatomic/rpm-ostree/pull/1396
	utils.Run("rpm-ostree", "rebase", fmt.Sprintf(":%s", rev),
		"--custom-origin-url", customURL,
		"--custom-origin-description", "Managed by pivot tool")

	// Kill our dummy container
	podmanRemove(types.PivotName)

	// By default, delete the image.
	if !keep {
		utils.Run("podman", "rmi", imgid)
	}

	// Reboot the machine if asked to do so
	if reboot {
		utils.Run("systemctl", "reboot")
	}
}
