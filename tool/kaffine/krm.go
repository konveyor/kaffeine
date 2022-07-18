package kaffine

import (
	"fmt"
	"sort"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FunctionCatalog struct {
	// required
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Spec       struct {
		KrmFunctions []FunctionDefinition `json:"krmFunctions"`
	} `json:"spec"`
	// optional
	Metadata *v1.ObjectMeta `json:"metadata,omitempty"`
}

func MakeFunctionCatalog(name string) (fc FunctionCatalog) {
	fc.APIVersion = "config.kubernetes.io/v1alpha1"
	fc.Kind = "KRMFunctionCatalog"
	fc.Metadata = &v1.ObjectMeta{}
	fc.Metadata.Name = name
	fc.Metadata.SetCreationTimestamp(v1.Now())

	return
}

var IgnoreAutoUpdates string = "kaffine.config/ignore-auto-updates"

type FunctionDefinition struct {
	// required
	Group       string `json:"group"`
	Description string `json:"description"`
	Publisher   string `json:"publisher"`
	Names       struct {
		Kind string `json:"kind"`
	} `json:"names"`
	Versions []FunctionVersion `json:"versions"`
	// optional
	Home        string         `json:"home,omitempty"`
	Maintainers []string       `json:"maintainers,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	Metadata    *v1.ObjectMeta `json:"metadata,omitempty"`
}

// Right now just lexicographically compares the version names
func (m FunctionDefinition) GetHighestVersion() FunctionVersion {
	sort.Slice(m.Versions, func(i, j int) bool {
		return m.Versions[i].Name < m.Versions[j].Name
	})

	return m.Versions[len(m.Versions)-1]
}

func (m FunctionDefinition) GetVersion(v string) (fv FunctionVersion, err error) {
	for _, fv = range m.Versions {
		if fv.Name == v {
			return fv, nil
		}
	}

	return fv, fmt.Errorf("no version '%s' in function '%s'", m.GroupName(), v)
}

// Get rightmost @ and get rightmost /
func ToGroupNameVersion(nameString string) (group string, name string, version string) {
	for i := len(nameString) - 1; i >= 0; i-- {
		if nameString[i:i+1] == "@" {
			version = nameString[i+1:]
			nameString = nameString[:i]
			break
		}
	}

	for i := len(nameString) - 1; i >= 0; i-- {
		if nameString[i:i+1] == "/" {
			group = nameString[:i]
			nameString = nameString[i+1:]
			break
		}
	}

	name = nameString
	return
}

func (m FunctionDefinition) GroupName() string {
	return m.Group + "/" + m.Names.Kind
}

type FunctionVersion struct {
	// required
	// Schema     struct{ OpenAPIV3Schema v1beta1.JSONSchemaProps `json:"openAPIV3Schema"` }  `json:"schema"`
	Name       string   `json:"name"`
	Idempotent bool     `json:"idempotent"`
	Usage      string   `json:"usage"`
	Examples   []string `json:"examples"`
	License    string   `json:"license"`
	Runtime    struct {
		Container FunctionRuntimeContainer `json:"container,omitempty"`
		Exec      FunctionRuntimeExec      `json:"exec,omitempty"`
	} `json:"runtime"`
	// optional
	Maintainers []string `json:"maintainers,omitempty"`
}

type FunctionRuntimeContainer struct {
	// required
	Image string `json:"image"`
	// optional
	Sha256              string `json:"sha256,omitempty"`
	RequireNetwork      bool   `json:"requireNetwork,omitempty"`
	RequireStorageMount bool   `json:"requireStorageMount,omitempty"`
}

type FunctionRuntimeExec struct {
	// required
	Platforms []FunctionRuntimePlatform `json:"platforms"`
}

type FunctionRuntimePlatform struct {
	// required
	Bin    string `json:"bin"`
	Os     string `json:"os"`
	Arch   string `json:"arch"`
	Uri    string `json:"uri"`
	Sha256 string `json:"sha256"`
}
