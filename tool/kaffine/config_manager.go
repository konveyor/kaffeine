package kaffine

import (
	"errors"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

type Config struct {
	FilePath string `json:"-"`

	Catalogs     []string `json:"catalogs"`
	Dependencies struct {
		KrmFunctions []string `json:"krmFunctions"`
	} `json:"dependencies"`
}

func MakeConfig(directory string) (c Config) {
	var data []byte
	filePath := filepath.Join(directory, "config.yaml")

	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		os.WriteFile(filePath, DefaultConfig, 0644)
		data = DefaultConfig
	} else {
		data, _ = os.ReadFile(filePath)
	}

	yaml.Unmarshal(data, &c)

	c.FilePath = filePath
	return
}

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
