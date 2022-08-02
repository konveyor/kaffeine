package kaffeine

import (
	"errors"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

// Config struct. Unmarshalled from ".kaffeine/config.yaml"
type Config struct {
	// The filepath, usually ".kaffeine/config.yaml"
	FilePath string `json:"-"`

	// The list of catalog URIs that kaffeine should manage
	Catalogs []string `json:"catalogs"`

	// The list of dependencies that kaffeine should manage. In the future, this
	// could be extended to pipelines and indirect dependencies.
	Dependencies struct {
		// The list of KRM Functions to manage. They should be in the format
		// "Group/Name" and can be followed by "@Version" to peg it to a specific
		// version
		KrmFunctions []string `json:"krmFunctions"`
	} `json:"dependencies"`
}

// Creates config struct
func MakeConfig(directory string) (c Config) {
	var data []byte
	filePath := filepath.Join(directory, "config.yaml")

	if _, err := os.Stat(filePath); !errors.Is(err, os.ErrNotExist) {
		data, _ = os.ReadFile(filePath)
	}

	yaml.Unmarshal(data, &c)

	c.FilePath = filePath
	return
}

// Saves config struct
func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	file, err := os.Create(c.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	file.Write(data)
	return nil
}
