package kaffeine

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/exp/maps"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type FunctionManager struct {
	Directory string

	CatMan *CatalogManager
	Cfg    *Config

	Installed map[string]FunctionDefinition
}

// Traverses the file tree upward, until it finds either a folder named
// ".kaffeine" or tries to go past "/". If no directory is found, it returns
// the result of:
// 	wd, _ := os.Getwd()
// 	filepath.Join(wd, "/.kaffeine/")`
func GetDirectory() (dir string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return
	}

	checkDir := wd

	for ok := true; ok; ok = (checkDir != "/") {
		info, err := os.Stat(filepath.Join(checkDir, "/.kaffeine/"))

		if !os.IsNotExist(err) && info.IsDir() {
			dir = filepath.Join(checkDir, "/.kaffeine/")
			break
		}

		checkDir = filepath.Clean(filepath.Join(checkDir, ".."))
	}

	if dir == "" {
		dir = filepath.Clean(filepath.Join(wd, "/.kaffeine/"))
	}

	return dir, err
}

// Returns a new KRM Function Manager struct.
// If directory == "", it will use GetDirectory() to find where to store its
// files.
func NewFunctionManager(directory string) *FunctionManager {
	if directory == "" {
		directory, _ = GetDirectory()
	}

	os.MkdirAll(directory, os.ModePerm)

	fm := FunctionManager{}

	fm.Directory = directory
	catman := MakeCatalogManager(directory)
	fm.CatMan = &catman
	cfg := MakeConfig(directory)
	fm.Cfg = &cfg

	for _, uri := range fm.Cfg.Catalogs {
		err := fm.CatMan.AddCatalogFromUri(uri)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}
	fm.Cfg.Catalogs = maps.Keys(fm.CatMan.Catalogs)

	// LIST CACHE
	// n    n     - Do nothing
	// n    Y     - Remove from cache
	// Y    n     - Attempt to fetch catalog and load into memory
	// Y    Y     - Load into memory
	os.MkdirAll(filepath.Join(fm.Directory, "functions"), os.ModePerm)
	fm.Installed = map[string]FunctionDefinition{}
	for _, fname := range fm.Cfg.Dependencies.KrmFunctions {
		_, err := fm.AddFunctionDefinition(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			continue
		}
	}

	return &fm
}

func (fm *FunctionManager) Save() error {
	fm.UpdateConfig()

	fnBakDirectory := filepath.Join(fm.Directory, "functions.bak")
	os.RemoveAll(fnBakDirectory)
	err := os.Rename(filepath.Join(fm.Directory, "functions"), fnBakDirectory)
	if err != nil {
		return err
	}

	for _, groupName := range maps.Keys(fm.Installed) {
		fm.SaveFunctionDefinition(groupName)
	}

	os.RemoveAll(fnBakDirectory)

	if installedCatalog, err := fm.GenerateInstalledCatalog(); err != nil {
		return err
	} else {
		os.WriteFile(filepath.Join(fm.Directory, "installed.yaml"), installedCatalog, os.ModePerm)
	}

	if err := fm.Cfg.Save(); err != nil {
		return err
	}
	if err := fm.CatMan.Save(); err != nil {
		return err
	}

	return nil
}

func (fm *FunctionManager) SaveFunctionDefinition(fname string) (fd FunctionDefinition, error error) {
	group, name, _ := ToGroupNameVersion(fname)
	groupName := group + "/" + name
	fd, ok := fm.Installed[groupName]
	if !ok {
		return fd, fmt.Errorf("function '%s' not installed (check spelling?)", groupName)
	}

	fnDir := filepath.Join(fm.Directory, "functions", fd.Group)
	// fmt.Println(fnDir)
	fnFile := fd.Names.Kind + ".yaml"

	os.MkdirAll(fnDir, os.ModePerm)

	// FIXME: Better binary management
	if len(fd.Versions[0].Runtime.Exec.Platforms) > 0 {
		// fmt.Println("HEEERE")
		oldUri := fd.Versions[0].Runtime.Exec.Platforms[0].Uri
		resp, err := http.Get(oldUri)
		if err != nil {
			// fmt.Println(err)
			return fd, err
		}
		defer resp.Body.Close()

		binFile := filepath.Join(fnDir, fd.Names.Kind+filepath.Ext(oldUri))
		// fmt.Println(binFile)
		out, err := os.Create(binFile)
		if err != nil {
			return fd, err
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fd, err
		}

		fd.Metadata.Annotations[OriginalBinaryLocation] = oldUri
		cpy := make([]FunctionRuntimePlatform, len(fd.Versions[0].Runtime.Exec.Platforms))
		copy(cpy, fd.Versions[0].Runtime.Exec.Platforms)
		cpy[0].Uri = "file://" + binFile
		fd.Metadata.Annotations[LocalBinaryLocation] = "file://" + binFile
		fd.Versions[0].Runtime.Exec.Platforms = cpy
	}

	b, err := yaml.Marshal(fd)
	if err != nil {
		return fd, err
	}
	os.WriteFile(filepath.Join(fnDir, fnFile), b, os.ModePerm)

	return fd, nil
}

func (fm *FunctionManager) AddFunctionDefinition(fname string) (fn FunctionDefinition, err error) {
	group, name, _ := ToGroupNameVersion(fname)
	groupName := group + "/" + name
	if _, ok := fm.Installed[groupName]; ok {
		return fn, fmt.Errorf("function '%s' already installed", fname)
	}

	fn, err = fm.GetCachedFunctionDefinition(fname)
	if err != nil {
		fn, err = fm.GetExternalFunctionDefinition(fname)

		if err != nil {
			return fn, err
		}
	}

	if _, ok := fm.Installed[fn.GroupName()]; ok {
		return fn, fmt.Errorf("function '%s' already installed", fn.GroupName())
	}

	// FIXME: Better binary management
	if val, ok := fn.Metadata.Annotations[OriginalBinaryLocation]; ok && val != "" {
		cpy := make([]FunctionRuntimePlatform, len(fn.Versions[0].Runtime.Exec.Platforms))
		copy(cpy, fn.Versions[0].Runtime.Exec.Platforms)
		fn.Metadata.Annotations[LocalBinaryLocation] = cpy[0].Uri
		cpy[0].Uri = val
		fn.Versions[0].Runtime.Exec.Platforms = cpy
	}

	fm.Installed[fn.GroupName()] = fn

	return fn, nil
}

func (fm *FunctionManager) RemoveFunctionDefinition(fname string) (oldFd FunctionDefinition, err error) {
	group, name, _ := ToGroupNameVersion(fname)
	groupName := group + "/" + name

	if _, ok := fm.Installed[fname]; !ok {
		return oldFd, fmt.Errorf("function with name '%s' not installed", groupName)
	}

	oldFd = fm.Installed[groupName]
	delete(fm.Installed, groupName)

	return oldFd, nil
}

// returns a function with a single version
func (fm *FunctionManager) GetCachedFunctionDefinition(fname string) (fn FunctionDefinition, err error) {
	group, name, version := ToGroupNameVersion(fname)
	fnDir := filepath.Join(fm.Directory, "functions", group)
	fnFile := name + ".yaml"

	if _, err := os.Stat(filepath.Join(fnDir, fnFile)); errors.Is(err, os.ErrNotExist) {
		return fn, fmt.Errorf("function definition '%s' not found in cache", fname)
	}

	// Load from file
	b, err := os.ReadFile(filepath.Join(fnDir, fnFile))
	if err != nil {
		return
	}

	err = yaml.Unmarshal(b, &fn)
	if err != nil {
		return
	}

	if len(fn.Versions) != 1 {
		return fn, fmt.Errorf("cached function definition for '%s' has does not have exactly 1 version", fname)
	}

	if fn.Metadata == nil {
		fn.Metadata = &v1.ObjectMeta{Annotations: map[string]string{}}
	}

	if version != "" {
		if fn.Versions[0].Name != version {
			return fn, fmt.Errorf("cached function definition for '%s' does not have version", version)
		}
		fn.Metadata.Annotations[IgnoreAutoUpdates] = "true"
	} else {
		fn.Metadata.Annotations[IgnoreAutoUpdates] = "false"
	}

	return
}

// returns a function with a single version
func (fm *FunctionManager) GetExternalFunctionDefinition(fname string) (fn FunctionDefinition, err error) {
	_, _, version := ToGroupNameVersion(fname)
	result, err := fm.CatMan.Search(fname, false)
	if err != nil {
		return
	}
	if len(result) == 0 {
		return fn, fmt.Errorf("no functions with name '%s'", fname)
	}
	if len(result) > 1 {
		return fn, fmt.Errorf("more than one function found with search term '%s'", fname)
	}

	fn = result[0]

	if fn.Metadata == nil {
		fn.Metadata = &v1.ObjectMeta{Annotations: map[string]string{}}
	}

	var v FunctionVersion

	if version == "" {
		v = fn.GetHighestVersion()
		fn.Metadata.Annotations[IgnoreAutoUpdates] = "false"
	} else {
		v, err = result[0].GetVersion(version)
		if err != nil {
			return
		}

		fn.Metadata.Annotations[IgnoreAutoUpdates] = "true"
	}

	fn.Versions = []FunctionVersion{v}

	return
}

func (fm *FunctionManager) UpdateFunctionDefinition(fname string) (oldFn FunctionDefinition, err error) {
	oldFn, err = fm.RemoveFunctionDefinition(fname)
	if err != nil {
		return
	}
	if oldFn.Metadata.Annotations[IgnoreAutoUpdates] == "true" {
		fm.Installed[oldFn.GroupName()] = oldFn
		// return FunctionDefinition{}, fmt.Errorf("attempted to update function with pegged version. please remove the function in question first.")
		return FunctionDefinition{}, nil
	}

	var newFn FunctionDefinition
	newFn, err = fm.GetExternalFunctionDefinition(fname)
	if err != nil {
		fm.Installed[oldFn.GroupName()] = oldFn
		return
	}

	fm.Installed[newFn.GroupName()] = newFn

	return
}

func (fm *FunctionManager) UpdateAllFunctionDefinitions() (oldFns []FunctionDefinition, errs []error) {
	for fname, _ := range fm.Installed {
		fd, err := fm.UpdateFunctionDefinition(fname)
		oldFns = append(oldFns, fd)
		errs = append(errs, err)
	}

	return
}

func (fm *FunctionManager) SearchFunctionDefintions(fname string) (result []byte, err error) {
	fds, err := fm.CatMan.Search(fname, true)
	if err != nil {
		return nil, err
	}
	fc := MakeFunctionCatalog("Search results for '" + fname + "'")
	fc.Spec.KrmFunctions = append(fc.Spec.KrmFunctions, fds...)
	return yaml.Marshal(fc)
}

func (fm *FunctionManager) GenerateInstalledCatalog() (result []byte, err error) {
	fc := MakeFunctionCatalog("kaffeine Managed Functions")
	for _, fn := range fm.Installed {
		// FIXME: Better binary management
		if val, ok := fn.Metadata.Annotations[LocalBinaryLocation]; ok && val != "" {
			fn.Versions[0].Runtime.Exec.Platforms[0].Uri = val
		}
		fc.Spec.KrmFunctions = append(fc.Spec.KrmFunctions, fn)
	}
	result, err = yaml.Marshal(fc)
	if err != nil {
		return result, err
	}

	// FIXME: Better binary management
	for _, fn := range fm.Installed {
		if val, ok := fn.Metadata.Annotations[OriginalBinaryLocation]; ok && val != "" {
			fn.Versions[0].Runtime.Exec.Platforms[0].Uri = val
		}
	}

	return result, err
}

func (fm *FunctionManager) UpdateConfig() (err error) {
	fm.Cfg.Catalogs = maps.Keys(fm.CatMan.Catalogs)

	fm.Cfg.Dependencies.KrmFunctions = make([]string, 0)
	for groupName, fd := range fm.Installed {
		fname := groupName
		if fd.Metadata != nil {
			if val, ok := fd.Metadata.Annotations[IgnoreAutoUpdates]; ok && val == "true" {
				fname = fname + "@" + fd.Versions[0].Name
			}
		}
		fm.Cfg.Dependencies.KrmFunctions = append(fm.Cfg.Dependencies.KrmFunctions, fname)
	}

	return nil
}
