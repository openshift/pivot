package cmd

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	// Enable sha256 in container image references
	_ "crypto/sha256"

	"github.com/golang/glog"
	"github.com/openshift/pivot/types"
	"github.com/openshift/pivot/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	imgref "github.com/containers/image/docker/reference"
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
	kubeletAuthFile = "/var/lib/kubelet/config.json"
	// File containing kernel arg changes for tuning
	kernelTuningFile = "/etc/pivot/kernel-args"
	cmdLineFile      = "/proc/cmdline"
)

// TODO: fill out the whitelist
// tuneableArgsWhitelist contains allowed keys for tunable arguments
var tuneableArgsWhitelist = map[string]bool{
	"nosmt": true,
}

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

// isArgTuneable returns if the argument provided is allowed to be modified
func isArgTunable(arg string) bool {
	return tuneableArgsWhitelist[arg]
}

// isArgInUse checks to see if the argument is already in use by the system currently
func isArgInUse(arg, cmdLinePath string) (bool, error) {
	if cmdLinePath == "" {
		cmdLinePath = cmdLineFile
	}
	content, err := ioutil.ReadFile(cmdLinePath)
	if err != nil {
		return false, err
	}

	checkable := string(content)
	if strings.Contains(checkable, arg) {
		return true, nil
	}
	return false, nil
}

// parseTuningFile parses the kernel argument tuning file
func parseTuningFile(tuningFilePath, cmdLinePath string) ([]types.TuneArgument, []types.TuneArgument, error) {
	addArguments := []types.TuneArgument{}
	deleteArguments := []types.TuneArgument{}
	if tuningFilePath == "" {
		tuningFilePath = kernelTuningFile
	}
	if cmdLinePath == "" {
		cmdLinePath = cmdLineFile
	}
	// Return fast if the file does not exist
	if _, err := os.Stat(tuningFilePath); os.IsNotExist(err) {
		glog.V(2).Infof("no kernel tuning needed as %s does not exist", tuningFilePath)
		// This isn't an error. Return out.
		return addArguments, deleteArguments, err
	}
	// Read and parse the file
	file, err := os.Open(tuningFilePath)
	// Clean up
	defer file.Close()

	if err != nil {
		// If we have an issue reading return an error
		glog.Infof("Unable to open %s for reading: %v", tuningFilePath, err)
		return addArguments, deleteArguments, err
	}

	// Parse the tuning lines
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ADD ") {
			// NOTE: Today only specific bare kernel arguments are allowed so
			// there is not a need to split on =.
			key := strings.TrimSpace(line[len("ADD "):])
			if isArgTunable(key) {
				// Find out if the argument is in use
				inUse, err := isArgInUse(key, cmdLinePath)
				if err != nil {
					return addArguments, deleteArguments, err
				}
				if !inUse {
					addArguments = append(addArguments, types.TuneArgument{Key: key, Bare: true})
				} else {
					glog.Infof(`skipping "%s" as it is already in use`, key)
				}
			} else {
				glog.Infof("%s not a whitelisted kernel argument", key)
			}
		} else if strings.HasPrefix(line, "DELETE ") {
			// NOTE: Today only specific bare kernel arguments are allowed so
			// there is not a need to split on =.
			key := strings.TrimSpace(line[len("DELETE "):])
			if isArgTunable(key) {
				inUse, err := isArgInUse(key, cmdLinePath)
				if err != nil {
					return addArguments, deleteArguments, err
				}
				if inUse {
					deleteArguments = append(deleteArguments, types.TuneArgument{Key: key, Bare: true})
				} else {
					glog.Infof(`skipping "%s" as it is not present in the current argument list`, key)
				}
			} else {
				glog.Infof("%s not a whitelisted kernel argument", key)
			}
		} else {
			glog.V(2).Infof(`skipping malformed line in %s: "%s"`, tuningFilePath, line)
		}
	}
	return addArguments, deleteArguments, nil
}

// updateTuningArgs executes additions and removals of kernel tuning arguments
func updateTuningArgs(tuningFilePath, cmdLinePath string) (bool, error) {
	if cmdLinePath == "" {
		cmdLinePath = cmdLineFile
	}
	changed := false
	additions, deletions, err := parseTuningFile(tuningFilePath, cmdLinePath)
	if err != nil {
		return changed, err
	}

	// Execute additions
	for _, toAdd := range additions {
		if toAdd.Bare {
			changed = true
			utils.Run("rpm-ostree", "kargs", fmt.Sprintf("--append=%s", toAdd.Key))
		} else {
			// TODO: currently not supported
		}
	}
	// Execute deletions
	for _, toDelete := range deletions {
		if toDelete.Bare {
			changed = true
			utils.Run("rpm-ostree", "kargs", fmt.Sprintf("--delete=%s", toDelete.Key))
		} else {
			// TODO: currently not supported
		}
	}
	return changed, nil
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

// getRefDigest parses a Docker/OCI image reference and returns
// its digest, or an error if the string fails to parse as
// a "canonical" image reference with a digest.
func getRefDigest(ref string) (string, error) {
	refParsed, err := imgref.ParseNamed(ref)
	if err != nil {
		return "", fmt.Errorf("parsing reference: %q: %v", ref, err)
	}
	canon, ok := refParsed.(imgref.Canonical)
	if !ok {
		return "", fmt.Errorf("not canonical form: %q: %v", ref, err)
	}

	return canon.Digest().String(), nil
}

// compareOSImageURL determines whether two images are the same, or have
// matching digests.
func compareOSImageURL(current, desired string) (bool, error) {
	if current == desired {
		return true, nil
	}

	currentDigest, err := getRefDigest(current)
	if err != nil {
		return false, fmt.Errorf("parsing current osImageURL: %v", err)
	}
	desiredDigest, err := getRefDigest(desired)
	if err != nil {
		return false, fmt.Errorf("parsing desired osImageURL: %v", err)
	}

	if currentDigest == desiredDigest {
		glog.Infof("Current and target osImageURL have matching digest %q", currentDigest)
		return true, nil
	}

	return false, nil
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

	targetMatched, err := compareOSImageURL(previousPivot, imgid)
	if err != nil {
		glog.Fatalf("%v", err)
	}
	if targetMatched {
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
				glog.Fatalf("Failed to delete %s: %v", etcPivotFile, err)
			}
		}
	}

	// By default, delete the image.
	if !keep {
		// Related: https://github.com/containers/libpod/issues/2234
		utils.RunIgnoreErr("podman", "rmi", imgid)
	}

	// Check to see if we need to tune kernel arguments
	tuningChanged, err := updateTuningArgs(kernelTuningFile, cmdLineFile)
	if err != nil {
		glog.Infof("unable to parse tuning file %s: %s", kernelTuningFile, err)
	}
	// If tuning changes but the oscontainer didn't we still denote we changed
	// for the reboot
	if tuningChanged {
		changed = true
		if err != nil {
			glog.Infof(`Unable to remove kernel tuning file %s: "%s"`, kernelTuningFile, err)
		}

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
