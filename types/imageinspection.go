package types

import (
	"time"

	"github.com/opencontainers/go-digest"
)

const (
	// PivotDonePath is the path to the file used to denote pivot work
	PivotDonePath = "/etc/os-container-pivot.stamp"
	// PivotName is literally the name of the new pivot
	PivotName = "ostree-container-pivot"
)

// ImageInspection is a public implementation of
// https://github.com/projectatomic/skopeo/blob/master/cmd/skopeo/inspect.go#L20-L31
type ImageInspection struct {
	Name          string `json:",omitempty"`
	Tag           string `json:",omitempty"`
	Digest        digest.Digest
	RepoTags      []string
	Created       *time.Time
	DockerVersion string
	Labels        map[string]string
	Architecture  string
	Os            string
	Layers        []string
}
