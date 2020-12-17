package project

import (
	"fmt"
	"path/filepath"
	"regexp"
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
	Name               string
	Path               string
	Repository         string
	Revision           string
	FetchRevision      string
	Assemblies         []string
	AssemblyAttachment string
	Server             bool
	Submodules         []Module
	apply              Apply
	ApplyExecute       bool
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

// Returns a list of pom.xml files for this modules. If the relative parameter
// is set to "true", the path to the pom.xml files will be relative to the
// module root.
func (module *Module) PomFiles(relative bool) []string {
	var list []string
	if relative {
		utils.InDirectory(module.Path, func() {
			list = pomparse.FindPomFiles("")
		})
	} else {
		list = pomparse.FindPomFiles(module.Path)
	}
	return list
}

func (module *Module) RelativePath() string {
	// The path in the "<module>" tag needs to be relative to make maven happy!
	// Using the GetRelativePathEvalSymlinks function here to make sure the relative path is as short as possible.
	// Otherwise we might get a very long relative path if the current working dir is inside a symlink.
	//
	// Example:
	//     cwd = /home/bernd/workspace/graylog-project (where workspace is a symlink to /tmp/workspace)
	//     module.Path = /home/bernd/workspace/graylog-project-repos/graylog2-server
	//
	// Calling utils.GetRelativePath(module.Path) (which does not evaluate the symlink in the given path, but uses an
	// evaluated current working dir) would result in:
	//     ../../../home/bernd/workspace/graylog-project-repos/graylog2-server
	//
	// Calling utils.GetRelativePathEvalSymlinks(module.Path) returns the following instead:
	//     ../graylog-project-repos/graylog2-server
	//
	// TODO: This behavior is an implementation detail of utils.GetCwd which calls filepath.EvalSymlinks on the result
	// TODO: of os.Getwd to help filepath.Walk... This is a bit of a mess and should be fixed eventually.
	//
	// IntelliJ cannot handle the first variant so we have to make sure we get a short relative path when putting it
	// in the "<module/>" attribute in pom.xml.
	return utils.GetRelativePathEvalSymlinks(module.Path)
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

type projectOptions struct {
	moduleOverride bool
	pullRequests   bool
}

type projectOption func(*projectOptions)

func WithModuleOverride() projectOption {
	return func(o *projectOptions) {
		o.moduleOverride = true
	}
}

func WithPullRequests() projectOption {
	return func(o *projectOptions) {
		o.pullRequests = true
	}
}

func New(config config.Config, manifestFiles []string, options ...projectOption) Project {
	// Create a new project options object and process all given options
	opt := projectOptions{}
	for _, o := range options {
		o(&opt)
	}

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

		// Create apply data for this module and fill the blanks from the default apply data
		moduleApply := newApplyFromTemplate(module, defaultApply)

		if config.ForceHttpsRepos {
			moduleRepository = utils.ConvertGithubGitToHTTPS(module.Repository)
		}

		if module.HasSubmodules() {
			for _, submodule := range module.SubModules {
				path := getModulePath(repositoryRoot, moduleName, submodule)
				name := getMavenCoordinates(path).ArtifactId

				// Create apply data for this submodule and fill the blanks from the parent module apply data
				submoduleApply := newApplyFromTemplate(submodule, moduleApply)

				if name == "" {
					name = moduleName
				}

				submodules = append(submodules, Module{
					Name:               name,
					Path:               path,
					Repository:         moduleRepository,
					Revision:           module.Revision,
					Assemblies:         submodule.Assemblies,
					AssemblyAttachment: submodule.AssemblyAttachment,
					apply:              submoduleApply,
				})
			}
		}

		path := getModulePath(repositoryRoot, moduleName, module)
		name := getMavenCoordinates(path).ArtifactId

		if name == "" {
			name = moduleName
		}

		newModule := Module{
			Name:               name,
			Path:               path,
			Repository:         moduleRepository,
			Revision:           module.Revision,
			Assemblies:         module.Assemblies,
			AssemblyAttachment: module.AssemblyAttachment,
			Server:             module.Server,
			Submodules:         submodules,
			apply:              moduleApply,
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

	if len(config.Checkout.ModuleOverride) > 0 && opt.moduleOverride {
		projectModules = applyModuleOverride(config, projectModules)
	}

	if len(config.Checkout.PullRequests) > 0 && opt.pullRequests {
		projectModules = applyPullRequestsOverride(config, projectModules)
	}

	project := Project{
		config:  config,
		Server:  server,
		Modules: projectModules,
	}

	return project
}

func newApplyFromTemplate(module manifest.ManifestModule, template Apply) Apply {
	newApply := Apply{
		FromRevision: module.Apply.FromRevision,
		NewBranch:    module.Apply.NewBranch,
		NewVersion:   module.Apply.NewVersion,
	}

	// Fill in missing Apply attributes from template
	if err := mergo.Merge(&newApply, template); err != nil {
		logger.Fatal("Couldn't merge apply state: src=%#v dst=%#v", template, newApply)
	}

	return newApply
}

func applyModuleOverride(c config.Config, modules []Module) []Module {
	newModules := make([]Module, 0)

	// We check if there is an override for any module in our manifests
	for _, module := range modules {
		for _, override := range c.Checkout.ModuleOverride {
			// The override option is "<repo-match-substring>=<repo-replacement-name>@<revision>
			parts := strings.SplitN(override, "=", 2)
			if len(parts) != 2 {
				logger.Error("invalid override <%s> - skipping", override)
				continue
			}
			matchString, replacement := parts[0], parts[1]

			// Now split the replacement
			replacementParts := strings.SplitN(replacement, "@", 2)
			if len(parts) != 2 {
				logger.Error("invalid override <%s> - skipping", override)
				continue
			}
			replacementRepo, replacementRev := replacementParts[0], replacementParts[1]

			// The repo-match-substring of the override is used to select if the current module has
			// an override
			if strings.Contains(module.Repository, matchString) {
				// Build a new repo URL depending on the original type. (SSH, HTTPS, ...)
				newRepo, err := utils.ReplaceGitHubURL(module.Repository, replacementRepo)
				if err != nil {
					logger.Error("couldn't replace repository for module <%S>: %v", module.Name, err)
					continue
				}

				logger.Info("Overriding <%s@%s> with <%s@%s> for module <%s>",
					module.Repository, module.Revision, newRepo, replacementRev, module.Name)

				module.Repository = newRepo
				module.Revision = replacementRev
			}
		}

		newModules = append(newModules, module)
	}

	return newModules
}

func applyPullRequestsOverride(c config.Config, modules []Module) []Module {
	newModules := make([]Module, 0)
	regexp.MustCompile("^\\d+$")

	// We check if there is an override for any module in our manifests
	for _, module := range modules {
		for _, pullRequest := range c.Checkout.PullRequests {
			prRepo, prNumber, err := utils.ParseGitHubPRString(pullRequest)
			if err != nil {
				logger.Fatal("Error parsing pull request: %v", err)
			}

			repoUrl, err := utils.ParseGitHubURL(module.Repository)
			if err != nil {
				logger.Error("couldn't parse repository URL: %v", err)
			}

			if repoUrl.Matches(prRepo) {
				logger.Debug("Checking out pull-request %s for module %s", pullRequest, module.Name)

				module.Revision = fmt.Sprintf("pull-request/%d", prNumber)
				// PRs can be fetched from GitHub. See "Checking out pull requests locally":
				// https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/checking-out-pull-requests-locally
				module.FetchRevision = fmt.Sprintf("+refs/pull/%d/head:refs/remotes/origin/%s", prNumber, module.Revision)
			}
		}

		newModules = append(newModules, module)
	}

	return newModules
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
