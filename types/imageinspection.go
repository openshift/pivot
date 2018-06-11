package types

import (
	"time"

	"github.com/opencontainers/go-digest"
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
