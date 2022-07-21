package kaffeine

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

// Catalog Manager struct
type CatalogManager struct {
	// The directory of the catalogs to manage. Usually /.kaffine/catalogs/
	Directory string

	// The catalogs themselves. The key is the URI of the FunctionCatalog
	Catalogs map[string]FunctionCatalog

	// Every function contained within each catalog. The key is the
	// FunctionDefinition's GroupName (group + "/" + name). As a result, an effort
	// is made to ensure that every new catalog added has unique functions inside
	Functions map[string]FunctionDefinition
}

// Creates a CatalogManager struct
func MakeCatalogManager(directory string) (cm CatalogManager) {
	cm.Directory = filepath.Clean(filepath.Join(directory, "/catalogs"))
	cm.Catalogs = map[string]FunctionCatalog{}
	cm.Functions = map[string]FunctionDefinition{}

	os.MkdirAll(cm.Directory, os.ModePerm)

	return cm
}

// Moves all current cached catalogs into "../catalogs.bak", then tries to save
// all new catalogs into "catalogs".
func (cm *CatalogManager) Save() (err error) {
	bakDirectory := filepath.Clean(filepath.Join(cm.Directory, "../catalogs.bak"))
	os.RemoveAll(bakDirectory)
	err = os.Rename(cm.Directory, bakDirectory)
	if err != nil {
		return err
	}

	err = os.MkdirAll(cm.Directory, os.ModePerm)
	if err != nil {
		return err
	}

	for uri, cat := range cm.Catalogs {
		b, err := yaml.Marshal(cat)
		if err != nil {
			return err
		}

		hashName := fmt.Sprintf("%x", sha1.Sum([]byte(uri))) + ".yaml"
		err = os.WriteFile(filepath.Join(cm.Directory, hashName), b, os.ModePerm)
		if err != nil {
			return err
		}
	}

	os.RemoveAll(bakDirectory)

	return
}

// Adds the given FunctionCatalog to the catalog manager. Throws errors if the
// catalog:
// 	- has functions with GroupNames already present
// 	- has functions with no versions
func (cm *CatalogManager) AddCatalogFromStruct(uri string, cat FunctionCatalog) (err error) {
	// Check for conflicting names
	for _, fn := range cat.Spec.KrmFunctions {
		if _, ok := cm.Functions[fn.GroupName()]; ok {
			return errors.New("attempted to add catalog that contains conflicting names")
		}
		if len(fn.Versions) == 0 {
			return fmt.Errorf("attempted to add function '%s' with no versions", fn.GroupName())
		}
	}

	for _, fn := range cat.Spec.KrmFunctions {
		cm.Functions[fn.GroupName()] = fn
	}
	cm.Catalogs[uri] = cat

	return nil
}

// Adds the catalog with the given uri to the catalog manager. Throws errors if the
// catalog with the given uri is already present
func (cm *CatalogManager) AddCatalogFromUri(uri string) (err error) {
	// Already added
	if _, ok := cm.Catalogs[uri]; ok {
		return errors.New("catalog already present")
	}

	// Cached on the filesystem
	cat := FunctionCatalog{}

	cat, err = cm.GetCachedCatalog(uri)
	if err != nil {
		cat, err = cm.GetExternalCatalog(uri)

		// Fetch externally
		if err != nil {
			return err
		}
	}

	return cm.AddCatalogFromStruct(uri, cat)
}

// Tries to fetch the catalog with the given uri from the filesystem cache.
func (cm *CatalogManager) GetCachedCatalog(uri string) (fc FunctionCatalog, err error) {
	catalogFileInfo, err := os.ReadDir(cm.Directory)
	if err != nil {
		return
	}

	hashedFilename := fmt.Sprintf("%x", sha1.Sum([]byte(uri))) + ".yaml"
	ok := false
	for _, x := range catalogFileInfo {
		if x.Name() == hashedFilename {
			ok = true
			break
		}
	}

	if !ok {
		return fc, fmt.Errorf("cached catalog '%s' (hash '%s') not present in filesystem", uri, hashedFilename)
	}

	data, err := os.ReadFile(filepath.Join(cm.Directory, hashedFilename))
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &fc)
	if err != nil {
		return
	}

	return
}

// Tries to fetch the catalog from the given uri. External in this case means
// "not from the filesystem cache"
func (cm *CatalogManager) GetExternalCatalog(uri string) (fc FunctionCatalog, err error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return
	}

	var data []byte

	if u.Scheme == "file" {
		data, err = os.ReadFile(u.Path)
	} else {
		var resp *http.Response
		resp, err = http.Get(uri)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
	}

	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &fc)
	if err != nil {
		return
	}

	return
}

// Tries to remove the catalog from the CatalogManager
func (cm *CatalogManager) RemoveCatalog(uri string) (oldFc FunctionCatalog, err error) {
	if _, ok := cm.Catalogs[uri]; !ok {
		return oldFc, errors.New("catalog with uri not present")
	}

	for _, fn := range cm.Catalogs[uri].Spec.KrmFunctions {
		delete(cm.Functions, fn.GroupName())
	}

	oldFc = cm.Catalogs[uri]
	delete(cm.Catalogs, uri)

	return oldFc, nil
}

// Updates the catalog with the given uri. This is accomplished by deleting the
func (cm *CatalogManager) UpdateCatalog(uri string) (oldFc FunctionCatalog, err error) {
	oldFc, err = cm.RemoveCatalog(uri)
	if err != nil {
		return
	}

	var newFc FunctionCatalog
	newFc, err = cm.GetExternalCatalog(uri)
	if err != nil {
		cm.AddCatalogFromStruct(uri, oldFc)
		return
	}

	// Check for conflicting names
	for _, fn := range newFc.Spec.KrmFunctions {
		if _, ok := cm.Functions[fn.GroupName()]; ok {
			cm.AddCatalogFromStruct(uri, oldFc)
			return oldFc, errors.New("attempted to add catalog that contains conflicting names")
		}
		if len(fn.Versions) == 0 {
			cm.AddCatalogFromStruct(uri, oldFc)
			return oldFc, fmt.Errorf("attempted to add function '%s' with no versions", fn.GroupName())
		}
	}

	for _, fn := range newFc.Spec.KrmFunctions {
		cm.Functions[fn.GroupName()] = fn
	}
	cm.Catalogs[uri] = newFc

	return
}

// Executes UpdateCatalog on all catalogs in the struct. Returns an array of old
// catalogs and an array of errors encountered.
func (cm *CatalogManager) UpdateAllCatalogs() (oldFcs []FunctionCatalog, errs []error) {
	for uri, _ := range cm.Catalogs {
		fc, err := cm.UpdateCatalog(uri)
		oldFcs = append(oldFcs, fc)
		errs = append(errs, err)
	}

	return
}

// Searches for the given function
func (cm *CatalogManager) Search(fname string, lowercase bool) (fns []FunctionDefinition, err error) {
	group, name, version := ToGroupNameVersion(fname)
	groupName := name
	if group != "" {
		groupName = group + "/" + groupName
	}

	for _, queryDef := range cm.Functions {
		if lowercase {
			if !strings.Contains(strings.ToLower(queryDef.GroupName()), strings.ToLower(groupName)) {
				continue
			}
		} else if !strings.Contains(queryDef.GroupName(), groupName) {
			continue
		}

		if version != "" {
			var versions []FunctionVersion
			for _, queryVersion := range queryDef.Versions {
				if queryVersion.Name == version {
					versions = append(versions, queryVersion)
				}
			}

			if len(versions) == 0 {
				continue
			}

			queryDef.Versions = versions
		}
		fns = append(fns, queryDef)
	}

	return fns, nil
}
