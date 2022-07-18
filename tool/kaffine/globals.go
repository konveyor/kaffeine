package kaffine

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"

	_ "embed"
)

var Directory string = ""

//go:embed default_config.yaml
var DefaultConfig []byte

var Fm *FunctionManager

// Helper functions
func SHA1(s string) string {
	a := sha1.New()
	a.Write([]byte(s))
	return hex.EncodeToString(a.Sum(nil))
}

func InitializeGlobals() (err error) {
	// Directory
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	checkDir := wd

	for ok := true; ok; ok = (checkDir != "/") {
		info, err := os.Stat(filepath.Join(checkDir, "/.kaffine/"))

		if !os.IsNotExist(err) && info.IsDir() {
			Directory = filepath.Join(checkDir, "/.kaffine/")
			break
		}

		checkDir = filepath.Clean(filepath.Join(checkDir, ".."))
	}

	if Directory == "" {
		Directory = filepath.Clean(filepath.Join(wd, "/.kaffine/"))
	}

	err = os.MkdirAll(Directory, os.ModePerm)
	if err != nil {
		return err
	}

	Fm = NewFunctionManager(Directory)

	return
}

func DestroyGlobals() (err error) {
	err = Fm.Save()
	if err != nil {
		return
	}

	return
}
