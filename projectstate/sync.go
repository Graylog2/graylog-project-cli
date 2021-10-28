package projectstate

import (
	"encoding/json"
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/pom"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
)

const webModulesFile = "web-modules.json"

func Sync(project p.Project, config config.Config) {
	pom.WriteTemplates(config, project)

	if err := writeWebModules(project); err != nil {
		logger.Fatal("%s", err)
	}
}

type WebModules struct {
	Modules []WebModule `json:"modules"`
}

type WebModule struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func writeWebModules(project p.Project) error {
	var webModules []WebModule
	var serverModule p.Module

	p.ForEachModule(project, func(module p.Module) {
		if module.Server {
			serverModule = module
		}
	})

	if serverModule.Name == "" {
		return errors.New("Couldn't find any server module in project")
	}

	serverWebPath := ""

	// Find the server web path first
	p.ForEachModuleOrSubmodules(project, func(module p.Module) {
		if module.IsNpmModule() {
			// We need to find the web module for the server to get the correct output path
			if module.Repository == serverModule.Repository {
				serverWebPath = module.Path
			}
		}
	})

	if serverWebPath == "" {
		return errors.New("Couldn't find web output path for server module")
	}

	p.ForEachModuleOrSubmodules(project, func(module p.Module) {
		if module.IsNpmModule() {
			// Use a relative path to the module to make this work in other environments where the absolute
			// path might be different. (e.g. Docker container)
			modulePath, err := filepath.Rel(serverWebPath, module.Path)
			if err != nil {
				logger.Error("Couldn't get relative path for <%s>, using absolute path.", module.Path)
				modulePath = module.Path
			}

			webModules = append(webModules, WebModule{
				Name: module.Name,
				Path: modulePath,
			})
		}
	})

	return writeWebModulesFile(filepath.Join(serverWebPath, webModulesFile), webModules)
}

func writeWebModulesFile(path string, modules []WebModule) error {
	buf, err := json.MarshalIndent(WebModules{modules}, "", "  ")

	if err != nil {
		return errors.Wrap(err, "Couldn't serialize the web modules")
	}

	logger.Info("Generating %s", path)
	if err := ioutil.WriteFile(path, buf, 0644); err != nil {
		return errors.Wrapf(err, "Unable to write file %v", path)
	}

	return nil
}
