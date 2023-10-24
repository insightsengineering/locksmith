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
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

func parseDescriptionFileList(inputDescriptionFiles []DescriptionFile) []PackageDescription {
	var allPackages []PackageDescription
	for _, descriptionFile := range inputDescriptionFiles {
		processDescription(descriptionFile, &allPackages)
	}
	return allPackages
}

func parsePackagesFiles(repositoryPackageFiles map[string]string) map[string]PackagesFile {
	packagesFilesMap := make(map[string]PackagesFile)
	for repository, packagesFile := range repositoryPackageFiles {
		packagesFilesMap[repository] = processPackagesFile(packagesFile)
	}
	return packagesFilesMap
}

// Reads a string containing PACKAGES file, and returns structure with
// fields/properties required for further processing.
func processPackagesFile(content string) PackagesFile {
	var allPackages PackagesFile
	for _, lineGroup := range strings.Split(content, "\n\n") {
		if lineGroup == "" {
			continue
		}
		// Each lineGroup contains information about one package and is separated by an empty line.
		firstLine := strings.Split(lineGroup, "\n")[0]
		packageName := strings.ReplaceAll(firstLine, "Package: ", "")
		cleaned := cleanDescriptionOrPackagesEntry(lineGroup)
		packageMap := make(map[string]string)
		err := yaml.Unmarshal([]byte(cleaned), &packageMap)
		if err != nil {
			log.Error("Error reading ", packageName, " package data from PACKAGES: ", err)
		}
		var packageDependencies []Dependency
		processDependencyFields(packageMap, &packageDependencies)
		allPackages.Packages = append(
			allPackages.Packages,
			PackageDescription{
				packageName, packageMap["Version"], "", "", packageDependencies,
				"", "", "", "", "", "", "",
			},
		)
	}
	return allPackages
}

// Reads a string containing DESCRIPTION file and returns a structure with fields/properties
// required for further processing.
func processDescription(description DescriptionFile, allPackages *[]PackageDescription) {
	cleaned := cleanDescriptionOrPackagesEntry(description.Contents)
	packageMap := make(map[string]string)
	err := yaml.Unmarshal([]byte(cleaned), &packageMap)
	checkError(err)

	var packageDependencies []Dependency
	processDependencyFields(packageMap, &packageDependencies)
	*allPackages = append(
		*allPackages,
		PackageDescription{
			packageMap["Package"], packageMap["Version"], description.PackageSource, "", packageDependencies,
			description.RemoteType, description.RemoteHost, description.RemoteUsername, description.RemoteRepo,
			description.RemoteSubdir, description.RemoteRef, description.RemoteSha,
		},
	)
}

// Processes a multiline string representing information about one package from PACKAGES file, or the whole
// contents of DESCRIPTION file. Removes newlines occurring within filtered fields (which are predominantly
// fields containing lists of package dependencies). Also removes fields which are not required for further processing.
func cleanDescriptionOrPackagesEntry(description string) string {
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

func splitPackageName(r rune) bool {
	return r == ' ' || r == '('
}

// Processes a map containing a YAML-like object representing dependencies of a package.
// Returns a list of Dependency structures corresponding to dependency name, and version constraints.
func processDependencyFields(packageMap map[string]string,
	packageDependencies *[]Dependency) {
	dependencyFields := []string{"Depends", "Imports", "Suggests", "Enhances", "LinkingTo"}
	re := regexp.MustCompile(`\(.*\)`)
	for _, field := range dependencyFields {
		if _, ok := packageMap[field]; ok {
			dependencyList := strings.Split(packageMap[field], ",")
			for _, dependency := range dependencyList {
				if dependency == "" {
					continue
				}
				// There might be a space or '(' right after the package name,
				// so both space and '(' are treated as a delimiter to get the
				// package name from the first field.
				dependencyName := strings.FieldsFunc(strings.TrimSpace(dependency), splitPackageName)[0]
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
					Dependency{
						field, dependencyName, versionConstraintOperator, versionConstraintValue,
					},
				)
			}
		}
	}
}
