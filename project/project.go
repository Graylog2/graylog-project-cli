package project

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	"github.com/Graylog2/graylog-project-cli/pomparse"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/imdario/mergo"
)

type Project struct {
	config  config.Config
	Server  Module
	Modules []Module
}

type Apply struct {
	FromRevision string
	NewBranch    string
	NewVersion   string
}

type Module struct {
	Name         string
	Path         string
	Repository   string
	Revision     string
	Assembly     bool
	Server       bool
	Submodules   []Module
	apply        Apply
	ApplyExecute bool
}

func (module *Module) IsMavenModule() bool {
	return utils.FileExists(filepath.Join(module.Path, "pom.xml"))
}

func (module *Module) IsNpmModule() bool {
	return utils.FileExists(filepath.Join(module.Path, "package.json"))
}

func (module *Module) HasSubmodules() bool {
	return len(module.Submodules) > 0
}

func (module *Module) RelativePath() string {
	// The path in the "<module>" tag needs to be relative to make maven happy!
	return utils.GetRelativePath(module.Path)
}

func (module *Module) GroupId() string {
	return getMavenCoordinates(module.Path).GroupId
}

func (module *Module) ArtifactId() string {
	return getMavenCoordinates(module.Path).ArtifactId
}

func (module *Module) Version() string {
	return getMavenCoordinates(module.Path).Version
}

func (module *Module) ParentGroupId() string {
	return getMavenCoordinates(module.Path).ParentGroupId
}

func (module *Module) ParentArtifactId() string {
	return getMavenCoordinates(module.Path).ParentArtifactId
}

func (module *Module) ParentVersion() string {
	return getMavenCoordinates(module.Path).ParentVersion
}

func (module *Module) ParentRelativePath() string {
	return getMavenCoordinates(module.Path).ParentRelativePath
}

func (module *Module) HasParent() bool {
	coordinates := getMavenCoordinates(module.Path)
	return coordinates.ParentGroupId != "" && coordinates.ParentArtifactId != ""
}

func (module *Module) ApplyFromRevision() string {
	return module.apply.FromRevision
}

func (module *Module) ApplyNewBranch() string {
	return module.apply.NewBranch
}

func (module *Module) ApplyNewVersion() string {
	return module.apply.NewVersion
}

func getMavenCoordinates(path string) pomparse.MavenCoordinates {
	return pomparse.GetMavenCoordinates(filepath.Join(path, "pom.xml"))
}

func New(config config.Config, manifestFiles []string) Project {
	readManifest := manifest.ReadManifest(manifestFiles)

	var server Module

	// Make sure we use an absolute path!
	repositoryRoot := utils.GetAbsolutePath(config.RepositoryRoot)
	projectModules := make([]Module, 0)

	defaultApply := Apply{
		FromRevision: readManifest.DefaultApply.FromRevision,
		NewBranch:    readManifest.DefaultApply.NewBranch,
		NewVersion:   readManifest.DefaultApply.NewVersion,
	}

	for _, module := range readManifest.Modules {
		moduleName := utils.NameFromRepository(module.Repository)
		moduleRepository := module.Repository
		submodules := make([]Module, 0)

		if config.ForceHttpsRepos {
			moduleRepository = utils.ConvertGithubGitToHTTPS(module.Repository)
		}

		if module.HasSubmodules() {
			for _, submodule := range module.SubModules {
				path := getModulePath(repositoryRoot, moduleName, submodule)
				name := getMavenCoordinates(path).ArtifactId

				if name == "" {
					name = moduleName
				}

				submodules = append(submodules, Module{
					Name:       name,
					Path:       path,
					Repository: moduleRepository,
					Revision:   module.Revision,
					Assembly:   submodule.Assembly,
				})
			}
		}

		path := getModulePath(repositoryRoot, moduleName, module)
		name := getMavenCoordinates(path).ArtifactId

		if name == "" {
			name = moduleName
		}

		moduleApply := Apply{
			FromRevision: module.Apply.FromRevision,
			NewBranch:    module.Apply.NewBranch,
			NewVersion:   module.Apply.NewVersion,
		}

		// Merge the module `apply` field with the default apply values.
		mergo.Merge(&moduleApply, defaultApply)

		newModule := Module{
			Name:       name,
			Path:       path,
			Repository: moduleRepository,
			Revision:   module.Revision,
			Assembly:   module.Assembly,
			Server:     module.Server,
			Submodules: submodules,
			apply:      moduleApply,
		}

		// Set execute flag if the manifest should be applied if it contains apply config
		newModule.ApplyExecute = config.ApplyManifest.Execute

		projectModules = append(projectModules, newModule)

		// Decide if this module is the server module based on the config option
		if newModule.Server {
			// Only set if server is not already set
			if !server.Server {
				server = newModule
			} else {
				logger.Error("Server module already set to %v, not setting it to %v", server.Name, newModule.Name)
				logger.Error("Check your manifests %v, only one module should have 'server: true'", manifestFiles)
			}
		}
	}

	if server.Name == "" {
		logger.Fatal("No server module in manifests: %v", manifestFiles)
	}

	project := Project{
		config:  config,
		Server:  server,
		Modules: projectModules,
	}

	return project
}

func getModulePath(repositoryPath string, name string, module manifest.ManifestModule) string {
	if module.Path == "" {
		return filepath.Join(repositoryPath, name)
	} else {
		return filepath.Join(repositoryPath, name, module.Path)
	}
}

func MavenDependencies(project Project) []Module {
	dependencies := make([]Module, 0)

	ForEachModuleOrSubmodules(project, func(module Module) {
		if module.IsMavenModule() {
			dependencies = append(dependencies, module)
		}
	})

	return dependencies
}

func MavenAssemblies(project Project) []string {
	dependencies := make([]string, 0)

	ForEachModuleOrSubmodules(project, func(module Module) {
		if module.IsMavenModule() && module.Assembly {
			dependencies = append(dependencies, fmt.Sprintf("%s:%s", module.GroupId(), module.ArtifactId()))
		}
	})

	return dependencies
}

func forEachModule(modules []Module, callback func(Module)) {
	for _, module := range modules {
		callback(module)
	}
}

func forEachModuleOrSubmodules(modules []Module, callback func(Module)) {
	for _, module := range modules {
		if module.HasSubmodules() {
			forEachModule(module.Submodules, callback)
		} else {
			callback(module)
		}
	}
}

func forEachModuleAndSubmodules(modules []Module, callback func(Module)) {
	for _, module := range modules {
		callback(module)
		if module.HasSubmodules() {
			forEachModule(module.Submodules, callback)
		}
	}
}

func ForEachModule(project Project, callback func(Module)) {
	forEachModule(project.Modules, callback)
}

func ForEachSelectedModule(project Project, callback func(Module)) {
	forEachModule(SelectedModules(project), callback)
}

func ForEachModuleOrSubmodules(project Project, callback func(Module)) {
	forEachModuleOrSubmodules(project.Modules, callback)
}

func ForEachSelectedModuleOrSubmodules(project Project, callback func(Module)) {
	forEachModuleOrSubmodules(SelectedModules(project), callback)
}

func ForEachModuleAndSubmodules(project Project, callback func(Module)) {
	forEachModuleAndSubmodules(project.Modules, callback)
}

func ForEachSelectedModuleAndSubmodules(project Project, callback func(Module)) {
	forEachModuleAndSubmodules(SelectedModules(project), callback)
}

func SelectedModules(project Project) []Module {
	var selectedModules []Module

	if project.config.SelectedModules == "" {
		return project.Modules
	}

	substrings := strings.Split(project.config.SelectedModules, ",")

	for _, module := range project.Modules {
		for _, substring := range substrings {
			if strings.Contains(module.Name, substring) {
				selectedModules = append(selectedModules, module)
			}
		}
	}

	return selectedModules
}

func MaxModuleNameLength(project Project) int {
	maxNameLength := 0

	ForEachSelectedModule(project, func(module Module) {
		if len(module.Name) > maxNameLength {
			maxNameLength = len(module.Name)
		}
	})

	return maxNameLength
}

func HasModule(project Project, groupId string, artifactId string) (bool, Module) {
	var matchingModule Module
	var matched bool

	forEachModuleOrSubmodules(project.Modules, func(module Module) {
		c := getMavenCoordinates(module.Path)

		if (c.GroupId == groupId || c.ParentGroupId == groupId) && (c.ArtifactId == artifactId || c.ParentArtifactId == artifactId) {
			matchingModule = module
			matched = true
		}
	})

	return matched, matchingModule
}
