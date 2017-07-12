package manifest

import (
	"encoding/json"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/fatih/color"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

const DownloadedManifestPrefix = "downloaded-manifest"

type Manifest struct {
	Modules      []ManifestModule `json:"modules"`
	Includes     []string         `json:"includes,omitempty"`
	DefaultApply ManifestApply    `json:"default_apply,omitempty"`
}

type ManifestModule struct {
	Repository         string           `json:"repository,omitempty"`
	Revision           string           `json:"revision,omitempty"`
	Maven              string           `json:"maven,omitempty"`
	Path               string           `json:"path,omitempty"`
	Assembly           bool             `json:"assembly,omitempty"`
	AssemblyDescriptor string           `json:"assembly_descriptor,omitempty"`
	Server             bool             `json:"server,omitempty"`
	SubModules         []ManifestModule `json:"submodules,omitempty"`
	Apply              ManifestApply    `json:"apply,omitempty"`
}

func (mod ManifestModule) HasSubmodules() bool {
	return len(mod.SubModules) > 0
}

type ManifestApply struct {
	FromRevision string `json:"from_revision,omitempty"`
	NewBranch    string `json:"new_branch,omitempty"`
	NewVersion   string `json:"new_version,omitempty"`
}

func readManifestFile(filename string) Manifest {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Fatal("Unable to read manifest: %v", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(bytes, &manifest); err != nil {
		logger.Fatal("Unable to decode manifest JSON: %v", err)
	}

	return manifest
}

func hasModule(modules []ManifestModule, module ManifestModule) bool {
	for _, m := range modules {
		if m.Repository == module.Repository {
			return true
		}
	}

	return false
}

func ReadManifest(filenames []string) Manifest {
	done := make(map[string]bool)
	return readManifestWithDoneState(filenames, &done)
}

func readManifestWithDoneState(paths []string, done *map[string]bool) Manifest {
	// Use the last manifest file as base
	lastManifest := paths[len(paths)-1]

	var filename string
	if filepath.IsAbs(lastManifest) {
		filename = lastManifest
	} else {
		filename = utils.GetAbsolutePath(lastManifest)
	}

	logger.Debug("Reading manifest file: %s", filename)
	selectedManifest := readManifestFile(filename)

	if len(paths) > 1 {
		// Add the other manifest files as includes to the last one
		for _, path := range paths[0 : len(paths)-1] {
			selectedManifest.Includes = append(selectedManifest.Includes, utils.GetAbsolutePath(path))
		}
	}

	var manifest *Manifest

	for _, file := range selectedManifest.Includes {
		includedFile := file
		if !filepath.IsAbs(includedFile) {
			// Includes which are not absolute paths should be resolved relative to the manifest file it's included in.
			includedFile = filepath.Join(filepath.Dir(filename), includedFile)
		}
		// Check if we already read the manifest file to avoid recursion
		if v := (*done)[includedFile]; v == true {
			logger.ColorInfo(color.FgYellow, "Recursion detected! Already read manifest file: %v (check manifest %v)", includedFile, filename)
			continue
		} else {
			(*done)[includedFile] = true
		}

		logger.Debug("Read included manifest: %v", includedFile)
		includedManifest := readManifestWithDoneState([]string{includedFile}, done)

		if manifest == nil {
			manifest = &includedManifest
			continue
		}

		for _, mod := range includedManifest.Modules {
			if hasModule(manifest.Modules, mod) {
				logger.Debug("Skipping duplicate module: %v", mod.Repository)
			} else {
				manifest.Modules = append(manifest.Modules, mod)
			}
		}
	}

	if manifest == nil {
		return selectedManifest
	}

	for _, mod := range selectedManifest.Modules {
		if hasModule(manifest.Modules, mod) {
			logger.Debug("Skipping duplicate module: %v", mod.Repository)
		} else {
			manifest.Modules = append(manifest.Modules, mod)
		}
	}

	return *manifest
}

// Downloads the given manifest URL into a local temporary file.
// It returns the path to the temporary file. (The caller is responsible to remove the temporary file!)
func DownloadManifestFromGitHub(manifestUrl string, authToken string) string {
	f, err := ioutil.TempFile("", DownloadedManifestPrefix)
	if err != nil {
		logger.Fatal("Unable to create temp file: %v", err)
	}

	if _, err := f.Write(fetchManifestFromGitHub(manifestUrl, authToken)); err != nil {
		logger.Fatal("Unable to write manifest to temp file: %v", err)
	}
	f.Close()

	return f.Name()
}

func fetchManifestFromGitHub(url string, authToken string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Fatal("Unable to create request for <%s>: %v", url, err)
	}

	if authToken != "" {
		req.Header.Add("Authorization", "token "+authToken)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Fatal("Unable to fetch manifest from <%s>: %v", req.URL, err)
	}

	if res.StatusCode > 200 {
		logger.Fatal("Requesting manifest <%s> failed: %s\nUse GPC_AUTH_TOKEN or --auth-token to access private repositories!", req.URL, res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Fatal("Unable to read body from <%s>: %v", req.URL, err)
	}
	defer res.Body.Close()

	return bytes
}

const ManifestStateFile = ".graylog-project-manifest-state"

type ManifestStateJSON struct {
	// DEPRECATED use the `Files' field!
	File  string   `json:"file,omitempty"`
	Files []string `json:"files"`
}

type ManifestState struct {
	files []string
}

func (m ManifestState) Files() []string {
	return m.files
}

func WriteState(filenames []string) {
	var files []string

	for _, file := range filenames {
		files = append(files, filepath.Clean(file))
	}

	buf, err := json.Marshal(ManifestStateJSON{Files: files})
	if err != nil {
		logger.Fatal("Unable to serialize manifest state: %v", err)
	}

	logger.Info("Writing manifest state to %v", ManifestStateFile)
	if err := ioutil.WriteFile(ManifestStateFile, buf, 0644); err != nil {
		logger.Fatal("Unable to write manifest state to %v: %v", ManifestStateFile, err)
	}
}

func ReadState() ManifestState {
	logger.Debug("Reading manifest state from %v", ManifestStateFile)

	var state ManifestStateJSON

	buf, err := ioutil.ReadFile(ManifestStateFile)
	if err != nil {
		logger.Fatal("Unable to read manifest state from %v: %v", ManifestStateFile, err)
	}

	if err := json.Unmarshal(buf, &state); err != nil {
		logger.Fatal("Unable to parse manifest state: %v", err)
	}

	var files []string

	// Handle deprecated File field
	if state.File != "" {
		files = append(files, state.File)
	}
	files = append(files, state.Files...)

	return ManifestState{files: files}
}
