package repo

import (
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/git"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"os"
	"path/filepath"
	"strings"
)

type RepoManager struct {
	Config config.Config
}

func NewRepoManager(config config.Config) *RepoManager {
	return &RepoManager{Config: config}
}

func (manager *RepoManager) SetupProjectRepositories(project p.Project) {
	manager.SetupProjectRepositoriesWithApply(project, false)
}

func (manager *RepoManager) SetupProjectRepositoriesWithApply(project p.Project, withApply bool) {
	if utils.FileExists(manifest.ManifestStateFile) && !manager.Config.Checkout.Force {
		prevManifests := manifest.ReadState().Files()

		for _, file := range prevManifests {
			if !utils.FileExists(file) {
				logger.Error("Manifest %v from state file does not exist anymore", file)
				goto modules
			}
		}

		prevProject := p.New(manager.Config, prevManifests)
		dirty := make([]p.Module, 0)
		revisions := make(map[string][]string, 0)

		for _, module := range prevProject.Modules {
			if !utils.FileExists(module.Path) {
				logger.Info("Skipping module %v because it does not exist yet", module.Name)
				continue
			}
			utils.InDirectory(module.Path, func() {
				revision := git.GitValue("rev-parse", "--abbrev-ref", "HEAD")

				if revision != module.Revision {
					dirty = append(dirty, module)
					revisions[module.Name] = []string{module.Revision, revision}
				}
			})
		}

		if len(dirty) != 0 {
			logger.Error("Not changing revisions, some repositories are on an unexpected revision:")
			for _, module := range dirty {
				logger.Error("  %v  (expected: %v, current: %v)", module.Name, revisions[module.Name][0], revisions[module.Name][1])
				utils.InDirectory(module.Path, func() {
					git.Git("status", "-s", "-b")
				})
			}
			os.Exit(1)
		}
	}

modules:
	for _, module := range project.Modules {
		logger.Info("Repository: %v", module.Repository)

		manager.EnsureRepository(module, module.Path)

		if module.Revision == "" {
			logger.Info("Missing revision for %v in manifest", module.Repository)
		}

		if withApply {
			manager.CheckoutRevision(module.Path, module.ApplyFromRevision())
		} else {
			manager.CheckoutRevision(module.Path, module.Revision)
		}
	}
}

func (manager *RepoManager) EnsureRepository(module p.Module, path string) {
	defer utils.Chdir(utils.GetCwd())

	if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
		if os.IsNotExist(err) {
			if manager.Config.Checkout.ShallowClone {
				logger.Info("Cloning %v into %v (shallow clone)", module.Repository, path)
				git.Git("clone", "--depth=1", "--no-single-branch", module.Repository, path)
			} else {
				logger.Info("Cloning %v into %v", module.Repository, path)
				git.Git("clone", module.Repository, path)
			}
		}
	} else {
		if manager.Config.Checkout.UpdateRepos {
			logger.Info("Updating %v", module.Repository)
			utils.Chdir(path)
			git.Git("fetch", "--all", "--tags")
		}
	}
}

func (manager *RepoManager) CheckoutRevision(repoPath string, revision string) {
	trimmedRevision := strings.TrimSpace(revision)

	if trimmedRevision == "" {
		logger.Fatal("Revision is empty, abort!")
	}

	defer utils.Chdir(utils.GetCwd())
	utils.Chdir(repoPath)

	logger.Info("Checkout revision: %v", trimmedRevision)

	// Create local branch first. Ignore error if branch already exists.
	git.GitErrOk("branch", trimmedRevision, "origin/"+trimmedRevision)
	// Checkout the <revision> branch
	git.Git("checkout", trimmedRevision)

	if manager.Config.Checkout.UpdateRepos {
		git.Git("merge", "--ff-only", "origin/"+trimmedRevision)
	}
}
