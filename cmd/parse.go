/*
Copyright 2023 F. Hoffmann-La Roche AG

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"strings"
	"regexp"
	"encoding/json"
	yaml "gopkg.in/yaml.v3"
)

type PackageDescription struct {
	Package      string       `json:"package"`
	Version      string       `json:"version"`
	Dependencies []Dependency `json:"dependencies"`
}

type Dependency struct {
	DependencyType  string `json:"type"`
	DependencyName  string `json:"name"`
	VersionOperator string `json:"operator"`
	VersionValue    string `json:"value"`
}

func parseDescriptionFileList(inputDescriptionFiles []string) {
	var allPackages []PackageDescription
	for _, descriptionFile := range inputDescriptionFiles {
		parseDescription(descriptionFile, &allPackages)
	}
	result, err := json.MarshalIndent(allPackages, "", "  ")
	checkError(err)
	log.Info(string(result))
}

func parseDescription(description string, allPackages *[]PackageDescription) {
	cleaned := cleanDescription(description)
	packageMap := make(map[string]string)
	err := yaml.Unmarshal([]byte(cleaned), &packageMap)
	checkError(err)
	packageName := packageMap["Package"]
	var packageDependencies []Dependency
	processDependencyFields(packageMap, &packageDependencies)
	*allPackages = append(
		*allPackages,
		PackageDescription{packageName, packageMap["Version"], packageDependencies},
	)
}

func cleanDescription(description string) string {
	lines := strings.Split(description, "\n")
	filterFields := []string{"Package:", "Version:", "Depends:", "Imports:", "Suggests:", "LinkingTo:"}
	outputContent := ""
	processingFilteredField := false
	for _, line := range lines {
		filteredFieldFound := false
		// Check if we start processing any of the filtered fields.
		for _, field := range filterFields {
			if strings.HasPrefix(line, field) {
				outputContent += "\n" + line
				processingFilteredField = true
				filteredFieldFound = true
				break
			}
		}
		// Append a line to currently processed filtered field.
		if processingFilteredField && strings.HasPrefix(line, " ") {
			outputContent += " " + strings.TrimSpace(line)
		}
		// We're not processing a filtered field anymore.
		if !filteredFieldFound && !strings.HasPrefix(line, " ") {
			processingFilteredField = false
		}
	}
	return outputContent
}

func processDependencyFields(packageMap map[string]string,
	packageDependencies *[]Dependency) {
	dependencyFields := []string{"Depends", "Imports", "Suggests", "Enhances", "LinkingTo"}
	re := regexp.MustCompile(`\(.*\)`)
	for _, field := range dependencyFields {
		if _, ok := packageMap[field]; ok {
			dependencyList := strings.Split(packageMap[field], ", ")
			for _, dependency := range dependencyList {
				dependencyName := strings.Split(dependency, " ")[0]
				versionConstraintOperator := ""
				versionConstraintValue := ""
				// Check if package is required in some particular version.
				if strings.Contains(dependency, "(") && strings.Contains(dependency, ")") {
					versionConstraint := re.FindString(dependency)
					// Remove brackets surrounding version constraint.
					versionConstraint = versionConstraint[1 : len(versionConstraint)-1]
					versionConstraintOperator = strings.Split(versionConstraint, " ")[0]
					versionConstraintValue = strings.Split(versionConstraint, " ")[1]
				}
				*packageDependencies = append(
					*packageDependencies,
					Dependency{field, dependencyName, versionConstraintOperator, versionConstraintValue},
				)
			}
		}
	}
}
