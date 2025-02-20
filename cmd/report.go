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

// Package cmd implements the core functionality of locksmith - the renv.lock generator.
package cmd

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"os"
	"strings"
)

type HTMLReport struct {
	Config []HTMLReportConfigItem
	// Errors will be written to the report only if locksmith
	// is running with logLevel = error or lower.
	Errors string
	// Errors will be written to the report only if locksmith
	// is running with logLevel = warning or lower.
	Warnings         string
	Dependencies     []HTMLReportDependency
	RenvLockContents string
}

type HTMLReportConfigItem struct {
	Key   string
	Value string
}

type HTMLReportDependency struct {
	Name       string
	Version    string
	Repository string
	Depends    string
	Imports    string
	LinkingTo  string
	Suggests   string
}

//go:embed template.html
var htmlTemplate string

func GenerateHTMLReport(outputPackageList []PackageDescription,
	inputPackageDescriptions []PackageDescription, packagesFiles map[string]PackagesFile,
	renvLockContents RenvLock, repositoryMap map[string]string) {

	var htmlReport HTMLReport

	htmlReport.Config = append(htmlReport.Config,
		HTMLReportConfigItem{"cfgFile", cfgFile},
		HTMLReportConfigItem{"logLevel", logLevel},
		HTMLReportConfigItem{"inputRenvLock", inputRenvLock},
		HTMLReportConfigItem{"outputRenvLock", outputRenvLock},
		HTMLReportConfigItem{"allowIncompleteRenvLock", allowIncompleteRenvLock},
		HTMLReportConfigItem{"updatePackages", updatePackages},
		HTMLReportConfigItem{"reportFileName", reportFileName},
		HTMLReportConfigItem{"inputPackageList", strings.ReplaceAll(inputPackageList, ",", ", ")},
		HTMLReportConfigItem{"inputRepositoryList", strings.ReplaceAll(inputRepositoryList, ",", ", ")},
		HTMLReportConfigItem{"inputPackages", strings.Join(inputPackages, ", ")},
		HTMLReportConfigItem{"inputRepositories", strings.Join(inputRepositories, ", ")},
	)

	renvLockString, err := json.MarshalIndent(renvLockContents, "", "  ")
	checkError(err)
	htmlReport.RenvLockContents = string(renvLockString)

	htmlReport.Errors = errorBuffer.String()
	htmlReport.Warnings = warnBuffer.String()

	// Find different types of dependencies for the packages added to the output renv.lock.
	for _, p := range outputPackageList {
		var dependsList, importsList, linkingToList, suggestsList, repository string
		// This represents the struct where we should look for the package details
		// including its dependencies.
		var expectedPackageLocation []PackageDescription

		if p.Source == "Repository" {
			// Get package dependencies from the PACKAGES file.
			// Set the repository to repository alias.
			repository = GetRepositoryKeyByValue(p.Repository, repositoryMap)
			expectedPackageLocation = packagesFiles[p.Repository].Packages

		} else {
			// Get package dependencies from DESCRIPTION files of input packages.
			// Set the repository to GitLab or GitHub.
			repository = p.Source
			expectedPackageLocation = inputPackageDescriptions
		}

		for _, pkg := range expectedPackageLocation {
			if pkg.Package == p.Package {
				for _, d := range pkg.Dependencies {
					switch d.DependencyType {
					case depends:
						dependsList += d.DependencyName + ", "
					case imports:
						importsList += d.DependencyName + ", "
					case linkingTo:
						linkingToList += d.DependencyName + ", "
					case suggests:
						suggestsList += d.DependencyName + ", "
					}
				}
				break
			}
		}
		htmlReport.Dependencies = append(htmlReport.Dependencies, HTMLReportDependency{
			p.Package, p.Version, repository,
			strings.TrimSuffix(dependsList, ", "),
			strings.TrimSuffix(importsList, ", "),
			strings.TrimSuffix(linkingToList, ", "),
			strings.TrimSuffix(suggestsList, ", "),
		})
	}
	t, err := template.New("locksmithReport").Parse(htmlTemplate)
	checkError(err)

	reportFile, err := os.Create(reportFileName)
	checkError(err)
	defer reportFile.Close()

	err = t.Execute(reportFile, htmlReport)
	checkError(err)
}
