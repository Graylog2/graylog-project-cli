package xmltemplate

import (
	"bytes"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/hashicorp/go-version"
	"os"
	"text/template"
)

type TemplateInventory struct {
	Server            p.Module
	Modules           []p.Module
	Dependencies      []p.Module
	Assemblies        map[string][]Assembly
	AssemblyPlatforms []string
}

type Assembly struct {
	GroupId    string
	ArtifactId string
	Attachment string
}

func (a Assembly) String() string {
	return fmt.Sprintf("%s:%s", a.GroupId, a.ArtifactId)
}

func mavenAssemblies(project p.Project) map[string][]Assembly {
	assemblies := make(map[string][]Assembly)

	p.ForEachModuleOrSubmodules(project, func(module p.Module) {
		if module.IsMavenModule() && len(module.Assemblies) > 0 {
			// Each module can be in one or more assemblies
			for _, assemblyId := range module.Assemblies {
				assemblies[assemblyId] = append(assemblies[assemblyId], Assembly{
					GroupId:    module.GroupId(),
					ArtifactId: module.ArtifactId(),
					Attachment: module.AssemblyAttachment,
				})
			}
		}
	})

	return assemblies
}

func WriteXmlFile(config config.Config, project p.Project, templateFile string, outputFile string) {
	logger.Info("Generating %v", outputFile)
	bts, err := os.ReadFile(templateFile)
	if err != nil {
		logger.Fatal("Error reading %v: %v", templateFile, err)
	}

	serverVersion, err := version.NewVersion(project.Server.Version())
	if err != nil {
		logger.Fatal("Error parsing server version %q: %v", project.Server.Version(), err)
	}

	tmpl, err := template.New(templateFile).Funcs(versionTemplateFuncs(serverVersion)).Parse(string(bts))
	if err != nil {
		logger.Fatal("Error parsing template: %v", err)
	}

	inventory := TemplateInventory{
		Server:            project.Server,
		Modules:           project.Modules,
		Dependencies:      p.MavenDependencies(project),
		Assemblies:        mavenAssemblies(project),
		AssemblyPlatforms: project.AssemblyPlatforms,
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, &inventory); err != nil {
		logger.Fatal("Unable to execute template: %v", err)
	}

	if err := os.WriteFile(outputFile, buf.Bytes(), 0644); err != nil {
		logger.Fatal("Unable to write file %v: %v", outputFile, err)
	}
}

func versionTemplateFuncs(serverVersion *version.Version) template.FuncMap {
	compare := func(compareFunc func(*version.Version) bool) func(any) (bool, error) {
		return func(versionValue any) (bool, error) {
			givenVersion, err := version.NewVersion(fmt.Sprintf("%s", versionValue))
			if err != nil {
				return false, fmt.Errorf("couldn't parse version %q: %w", versionValue, err)
			}
			return compareFunc(givenVersion), nil
		}
	}

	return template.FuncMap{
		"versionGt":  compare(serverVersion.GreaterThan),
		"versionGte": compare(serverVersion.GreaterThanOrEqual),
		"versionLt":  compare(serverVersion.LessThan),
		"versionLte": compare(serverVersion.LessThanOrEqual),
	}
}
