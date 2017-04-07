package pomparse

import (
	"encoding/xml"
	"github.com/Graylog2/graylog-project-cli/exec"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/utils"
	"io/ioutil"
	"os"
	"strings"
)

type MavenDependency struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope"`
}

type MavenPom struct {
	XMLName              xml.Name          `xml:"project"`
	GroupId              string            `xml:"groupId"`
	ArtifactId           string            `xml:"artifactId"`
	Version              string            `xml:"version"`
	ParentGroupId        string            `xml:"parent>groupId"`
	ParentArtifactId     string            `xml:"parent>artifactId"`
	ParentVersion        string            `xml:"parent>version"`
	ParentRelativePath   string            `xml:"parent>relativePath"`
	Modules              []string          `xml:"modules>module"`
	Properties           Properties        `xml:"properties"`
	Dependencies         []MavenDependency `xml:"dependencies>dependency"`
	DependencyManagement []MavenDependency `xml:"dependencyManagement>dependencies>dependency"`
}

func (pom MavenPom) PropertiesMap() map[string]string {
	properties := make(map[string]string)
	decoder := xml.NewDecoder(strings.NewReader(pom.Properties.XmlString))

	var curKey *string
	for {
		tok, _ := decoder.Token()

		if tok == nil {
			break
		}
		switch se := tok.(type) {
		case xml.StartElement:
			curKey = &se.Name.Local
		case xml.EndElement:
			curKey = nil
		case xml.CharData:
			if curKey != nil {
				properties[*curKey] = strings.TrimSpace(string(se.Copy()))
			}
		}
	}

	return properties
}

type Properties struct {
	XmlString string `xml:",innerxml"`
}

type MavenCoordinates struct {
	GroupId            string
	ArtifactId         string
	Version            string
	ParentGroupId      string
	ParentArtifactId   string
	ParentVersion      string
	ParentRelativePath string
}

func GetMavenCoordinates(path string) MavenCoordinates {
	if !utils.FileExists(path) {
		return MavenCoordinates{}
	}

	pom := ParsePom(path)

	groupId, err := utils.FirstNonEmpty(pom.GroupId, pom.ParentGroupId)
	if err != nil {
		logger.Fatal("Unable to get groupId from pom file %v (%#v): %v", path, pom, err)
	}
	artifactId, err := utils.FirstNonEmpty(pom.ArtifactId, pom.ParentArtifactId)
	if err != nil {
		logger.Fatal("Unable to get artifactId from pom file %v (%#v): %v", path, pom, err)
	}
	version, err := utils.FirstNonEmpty(pom.Version, pom.ParentVersion)
	if err != nil {
		logger.Fatal("Unable to get version from pom file %v (%#v): %v", path, pom, err)
	}

	return MavenCoordinates{
		GroupId:            groupId,
		ArtifactId:         artifactId,
		Version:            version,
		ParentGroupId:      pom.ParentGroupId,
		ParentArtifactId:   pom.ParentArtifactId,
		ParentVersion:      pom.ParentVersion,
		ParentRelativePath: pom.ParentRelativePath,
	}

}

func ParsePom(filename string) MavenPom {
	pomBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Fatal("Error reading pom file: %v", filename)
	}

	var mavenPom MavenPom

	if err := xml.Unmarshal(pomBytes, &mavenPom); err != nil {
		logger.Fatal("Unable to parse pom file: %v", err)
	}

	return mavenPom
}

func ParseEffectivePom(moduleName string, modulePath string) MavenPom {
	file, err := ioutil.TempFile("", moduleName)
	defer os.Remove(file.Name())
	if err != nil {
		logger.Fatal("Unable to create temp file for module %s: %v", moduleName, err)
	}

	output, err := exec.ExecCommandInPath(modulePath, "mvn", "help:effective-pom", "-Doutput="+file.Name())
	if err != nil {
		logger.Error("Unable to build effective pom for %s: %v", moduleName, err)
		logger.Error("%s", output.Stdout.String())
		logger.Fatal("%s", output.Stderr.String())
	}

	return ParsePom(file.Name())
}
