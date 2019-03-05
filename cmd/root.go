package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/openshift/pivot/types"
	"github.com/openshift/pivot/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// flag storage
var keep bool
var reboot bool
var container string
var exit_77 bool

const (
	// the number of times to retry commands that pull data from the network
	numRetriesNetCommands = 5
	etcPivotFile          = "/etc/pivot/image-pullspec"
	runPivotRebootFile    = "/run/pivot/reboot-needed"
	// Pull secret.  Written by the machine-config-operator
	kubeletAuthFile       = "/var/lib/kubelet/config.json"
)

// RootCmd houses the cobra config for the main command
var RootCmd = &cobra.Command{
	Use: "pivot [FLAGS] [IMAGE_PULLSPEC]",
	DisableFlagsInUseLine: true,
	Short: "Allows moving from one OSTree deployment to another",
	Args:  cobra.MaximumNArgs(1),
	Run:   Execute,
}

// init executes upon import
func init() {
	RootCmd.PersistentFlags().BoolVarP(&keep, "keep", "k", false, "Do not remove container image")
	RootCmd.PersistentFlags().BoolVarP(&reboot, "reboot", "r", false, "Reboot if changed")
	RootCmd.PersistentFlags().BoolVar(&exit_77, "unchanged-exit-77", false, "If unchanged, exit 77")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

// podmanRemove kills and removes a container
func podmanRemove(cid string) {
	utils.RunIgnoreErr("podman", "kill", cid)
	utils.RunIgnoreErr("podman", "rm", "-f", cid)
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

// pullAndRebase potentially rebases system if not already rebased.
func pullAndRebase(container string) (imgid string, changed bool) {
	defaultDeployment := getDefaultDeployment()

	previousPivot := ""
	if len(defaultDeployment.CustomOrigin) > 0 {
		if strings.HasPrefix(defaultDeployment.CustomOrigin[0], "pivot://") {
			previousPivot = defaultDeployment.CustomOrigin[0][len("pivot://"):]
			glog.Infof("Previous pivot: %s", previousPivot)
		}
	}

	var authArgs []string
	if utils.FileExists(kubeletAuthFile) {
		authArgs = append(authArgs, "--authfile", kubeletAuthFile)
	}

	// Use skopeo to canonicalize to $name@$digest, so we can refer to it reliably
	skopeoArgs := []string{"inspect"}
	skopeoArgs = append(skopeoArgs, authArgs...)
	skopeoArgs = append(skopeoArgs, fmt.Sprintf("docker://%s", container))
	output := utils.RunExt(true, numRetriesNetCommands, "skopeo", skopeoArgs...)

	var imagedata types.ImageInspection
	json.Unmarshal([]byte(output), &imagedata)
	imgid = fmt.Sprintf("%s@%s", imagedata.Name, imagedata.Digest)
	glog.Infof("Resolved to: %s", imgid)

	if previousPivot == imgid {
		changed = false
		return
	}

	// Pull the image
	podmanArgs := []string{"pull"}
	podmanArgs = append(podmanArgs, authArgs...)
	podmanArgs = append(podmanArgs, imgid)
	utils.RunExt(false, numRetriesNetCommands, "podman", podmanArgs...)

	// Clean up a previous container
	podmanRemove(types.PivotName)

	// `podman mount` wants a container, so let's make create a dummy one, but not run it
	cid := utils.RunGetOut("podman", "create", "--net=none", "--name", types.PivotName, imgid)
	// Use the container ID to find its mount point
	mnt := utils.RunGetOut("podman", "mount", cid)
	repo := fmt.Sprintf("%s/srv/repo", mnt)

	// Now we need to figure out the commit to rebase to

	// Commit label takes priority
	ostree_csum, ok := imagedata.Labels["com.coreos.ostree-commit"]
	if ok {
		if ostree_version, ok := imagedata.Labels["version"]; ok {
			glog.Infof("Pivoting to: %s (%s)", ostree_version, ostree_csum)
		} else {
			glog.Infof("Pivoting to: %s", ostree_csum)
		}
	} else {
		glog.Infof("No com.coreos.ostree-commit label found in metadata! Inspecting...")
		refs := strings.Split(utils.RunGetOut("ostree", "refs", "--repo", repo), "\n")
		if len(refs) == 1 {
			glog.Infof("Using ref %s", refs[0])
			ostree_csum = utils.RunGetOut("ostree", "rev-parse", "--repo", repo, refs[0])
		} else if len(refs) > 1 {
			glog.Fatalf("Multiple refs found in repo!")
		} else {
			// XXX: in the future, possibly scan the repo to find a unique .commit object
			glog.Fatalf("No refs found in repo!")
		}
	}

	// This will be what will be displayed in `rpm-ostree status` as the "origin spec"
	customURL := fmt.Sprintf("pivot://%s", imgid)

	// RPM-OSTree can now directly slurp from the mounted container!
	// https://github.com/projectatomic/rpm-ostree/pull/1732
	utils.Run("rpm-ostree", "rebase", "--experimental",
		fmt.Sprintf("%s:%s", repo, ostree_csum),
		"--custom-origin-url", customURL,
		"--custom-origin-description", "Managed by pivot tool")

	// Kill our dummy container
	podmanRemove(types.PivotName)

	changed = true
	return
}

// Execute runs the command
func Execute(cmd *cobra.Command, args []string) {
	var fromFile bool
	var container string
	if len(args) > 0 {
		container = args[0]
		fromFile = false
	} else {
		glog.Infof("Using image pullspec from %s", etcPivotFile)
		data, err := ioutil.ReadFile(etcPivotFile)
		if err != nil {
			glog.Fatalf("Failed to read from %s: %v", etcPivotFile, err)
		}
		container = strings.TrimSpace(string(data))
		fromFile = true
	}

	imgid, changed := pullAndRebase(container)

	// Delete the file now that we successfully rebased
	if fromFile {
		if err := os.Remove(etcPivotFile); err != nil {
			if !os.IsNotExist(err) {
				glog.Fatal("Failed to delete %s: %v", etcPivotFile, err)
			}
		}
	}

	// By default, delete the image.
	if !keep {
		// Related: https://github.com/containers/libpod/issues/2234
		utils.RunIgnoreErr("podman", "rmi", imgid)
	}

	if !changed {
		glog.Info("Already at target pivot; exiting...")
		if exit_77 {
			os.Exit(77)
		}
	} else if reboot || utils.FileExists(runPivotRebootFile) {
		// Reboot the machine if asked to do so
		utils.Run("systemctl", "reboot")
	}
}
