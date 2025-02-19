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

type HTMLReport struct {
	Config           []HTMLReportConfigurationItem
	Errors           string
	Warnings         string
	Dependencies     []HTMLReportDependency
	RenvLockContents string
}

type HTMLReportConfigurationItem struct {
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

func GenerateHTMLReport(outputPackageList []PackageDescription,
	inputPackages []PackageDescription, packagesFiles map[string]PackagesFile) {

	var htmlReport HTMLReport
	log.Info(outputPackageList)
	log.Info(inputPackages)
	for _, p := range outputPackageList {
		log.Info(p.Package, " ", p.Version)
		var depends, imports, linkingTo, suggests, repository string
		// This represents the struct where we should look for the package details
		// including its dependencies.
		var expectedPackageLocation []PackageDescription
		if p.Source == "Repository" {
			// Get package dependencies from the PACKAGES file.
			repository = p.Repository
			expectedPackageLocation = packagesFiles[p.Repository].Packages

		} else {
			// Get package dependencies from DESCRIPTION files of input packages.
			// Set the repository to GitLab or GitHub.
			repository = p.Source
			expectedPackageLocation = inputPackages
		}
		for _, pkg := range expectedPackageLocation {
			if pkg.Package == p.Package {
				for _, d := range pkg.Dependencies {
					log.Info(
						pkg.Package, " → ", d.DependencyName,
						" (", d.DependencyType, ")",
					)
					switch d.DependencyType {
					case "Depends":
						depends += d.DependencyName + ", "
					case "Imports":
						imports += d.DependencyName + ", "
					case "LinkingTo":
						linkingTo += d.DependencyName + ", "
					case "Suggests":
						suggests += d.DependencyName + ", "
					}
				}
				break
			}
		}
		htmlReport.Dependencies = append(htmlReport.Dependencies, HTMLReportDependency{
			p.Package, p.Version, repository, depends, imports, linkingTo, suggests,
		})
	}
}
