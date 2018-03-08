package xmltemplate

import (
	"bytes"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	p "github.com/Graylog2/graylog-project-cli/project"
	"io/ioutil"
	"text/template"
)

type TemplateInventory struct {
	Server       p.Module
	Modules      []p.Module
	Dependencies []p.Module
	Assemblies   map[string][]Assembly
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
	logger.Info("Generating %v file from template %v", outputFile, templateFile)
	bts, err := ioutil.ReadFile(templateFile)
	if err != nil {
		logger.Fatal("Error reading %v: %v", templateFile, err)
	}

	tmpl, err := template.New(templateFile).Parse(string(bts))
	if err != nil {
		logger.Fatal("Error parsing template: %v", err)
	}

	inventory := TemplateInventory{
		Server:       project.Server,
		Modules:      project.Modules,
		Dependencies: p.MavenDependencies(project),
		Assemblies:   mavenAssemblies(project),
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, &inventory); err != nil {
		logger.Fatal("Unable to execute template: %v", err)
	}

	if err := ioutil.WriteFile(outputFile, buf.Bytes(), 0644); err != nil {
		logger.Fatal("Unable to write file %v: %v", outputFile, err)
	}
}
