package pom

import (
	c "github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/pomparse"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/Graylog2/graylog-project-cli/xmltemplate"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func SetProperty(module p.Module, name string, value string) {
	pomFile := filepath.Join(module.Path, "pom.xml")
	properties := pomparse.ParsePom(pomFile).PropertiesMap()

	prevValue, hasName := properties[name]

	if hasName && prevValue == value {
		logger.Debug("Not updating property %v in %v, value does not change", name, module.Name)
		return
	}

	if strings.HasPrefix(prevValue, "${") {
		logger.Debug("Not updatig property %v in %v, existing value is a variable: %v", name, module.Name, prevValue)
		return
	}

	buf, err := ioutil.ReadFile(pomFile)
	if err != nil {
		logger.Fatal("Unable to read %v: %v", pomFile, err)
	}

	if hasName {
		logger.Info("Updating %s from %v to %v in %v", name, prevValue, value, module.Name)

		// Ensure that newlines are matched (via "(?s)")
		re := regexp.MustCompile("(?s)<" + regexp.QuoteMeta(name) + ">.+</" + regexp.QuoteMeta(name) + ">")

		newContent := re.ReplaceAllString(string(buf), "<"+name+">"+value+"</"+name+">")

		if err := ioutil.WriteFile(pomFile, []byte(newContent), 0); err != nil {
			logger.Fatal("Unable to set version in %v: %v", pomFile, err)
		}
	} else {
		logger.Debug("There is no \"%v\" property in %v that can be set and adding new properties is currently not supported :-(", name, pomFile)
	}

}

func SetParent(module p.Module, groupId string, artifactId string, version string, relativePath string) {
	SetParentIfMatches(module, groupId, artifactId, version, relativePath, func(module p.Module, pom pomparse.MavenPom) bool {
		return true
	})
}

func SetParentIfMatches(module p.Module, groupId string, artifactId string, version string, relativePath string, ifMatches func(module p.Module, pom pomparse.MavenPom) bool) {
	if groupId == "" || artifactId == "" || version == "" {
		logger.Fatal("One of groupId, artifactId or version is empty: groupId=%s artifactId=%s version=%s", groupId, artifactId, version)
	}

	pomFile := filepath.Join(module.Path, "pom.xml")
	pom := pomparse.ParsePom(pomFile)

	if !ifMatches(module, pom) {
		logger.Debug("Skip setting parent in %s because condition function was false", pomFile)
		return
	}

	buf, err := ioutil.ReadFile(pomFile)
	if err != nil {
		logger.Fatal("Unable to read %v: %v", pomFile, err)
	}

	logger.Debug("Setting parent to %s:%s:%s:%s in %v", groupId, artifactId, version, relativePath, module.Name)

	// Ensure that newlines are matched (via "(?s)")
	re := regexp.MustCompile("(?s)<parent>.+</parent>")

	newContent := re.ReplaceAllString(string(buf), "<parent>\n        <groupId>"+groupId+"</groupId>\n        <artifactId>"+artifactId+"</artifactId>\n        <version>"+version+"</version>\n        <relativePath>"+relativePath+"</relativePath>\n    </parent>")

	if err := ioutil.WriteFile(pomFile, []byte(newContent), 0); err != nil {
		logger.Fatal("Unable to set version in %v: %v", pomFile, err)
	}
}

var templateFileSuffixes = map[string]string{
	".xml.tmpl": ".xml",
	".xml-tmpl": ".xml",
}

func WriteTemplates(config c.Config, project p.Project) {
	// Scan project directory for all supported templates and generate the actual files
	err := filepath.Walk(utils.GetCwd(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		for templateSuffix, outputSuffix := range templateFileSuffixes {
			if strings.HasSuffix(path, templateSuffix) {
				templateFile := utils.GetRelativePath(path)
				outputFile := strings.TrimSuffix(templateFile, templateSuffix) + outputSuffix
				xmltemplate.WriteXmlFile(config, project, templateFile, outputFile)
			}
		}

		return nil
	})

	if err != nil {
		logger.Fatal("Unable to generate template files: %v", err)
	}
}
