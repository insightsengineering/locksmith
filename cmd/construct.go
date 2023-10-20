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
	"strconv"
	"strings"
)

type OutputPackage struct {
	Package    string `json:"package"`
	Version    string `json:"version"`
	Repository string `json:"repository"`
}

func constructOutputPackageList(packages []PackageDescription, packagesFiles map[string]PackagesFile,
	repositoryList []string) []OutputPackage {
	var outputPackageList []OutputPackage
	// Add all input packages to output list, as the packages should be downloaded from git repositories.
	for _, p := range packages {
		outputPackageList = append(outputPackageList, OutputPackage{
			p.Package, p.Version, p.Repository,
		})
	}
	for _, p := range packages {
		for _, d := range p.Dependencies {
			skipDependency := false
			if d.DependencyType == "Depends" || d.DependencyType == "Imports" || d.DependencyType == "Suggests" {
				if checkIfBasePackage(d.DependencyName) {
					log.Debug("Skipping package ", d.DependencyName, " as it is a base R package.")
					skipDependency = true
				}
				if checkIfPackageOnOutputList(d.DependencyName, outputPackageList) {
					log.Debug("Package ", d.DependencyName, " is already present on the output list.")
					skipDependency = true
				}
				if !skipDependency {
					log.Info(p.Package, " → ", d.DependencyName)
					resolveDependenciesRecursively(
						&outputPackageList, d.DependencyName, d.VersionOperator,
						d.VersionValue, repositoryList, packagesFiles, 1,
					)
				}
			}
		}
	}
	return outputPackageList
}

func resolveDependenciesRecursively(outputList *[]OutputPackage, name string, versionOperator string,
	versionValue string, repositoryList []string, packagesFiles map[string]PackagesFile, recursionLevel int) {
	var indentation string
	for i := 0; i < recursionLevel; i++ {
		indentation += "  "
	}
	if checkIfBasePackage(name) {
		log.Debug(indentation, "Skipping package ", name, " as it is a base R package.")
		return
	}
	if checkIfPackageOnOutputList(name, *outputList) {
		log.Debug(indentation, "Package ", name, " is already present on the output list.")
		return
	}
	for _, r := range repositoryList {
		// Check if the package is present in the PACKAGES file for the repository.
		for _, p := range packagesFiles[r].Packages {
			if p.Package == name {
				if r != repositoryList[0] {
					log.Warn(indentation, name, " not found in top repository.")
				}
				// Check if package in the repository is available in sufficient version.
				if !checkIfVersionSufficient(p.Version, versionOperator, versionValue) {
					// Try to retrieve the package from the next repository.
					log.Warn(
						indentation, p.Package, " in repository ", r,
						" is available in version ", p.Version,
						" which is insufficient according to requirement ",
						versionOperator, " ", versionValue,
					)
					continue
				}
				// Add package to the output list.
				// Repository is saved as an URL, and will be changed into an alias
				// during the processing of output package list into renv.lock file.
				*outputList = append(*outputList, OutputPackage{
					p.Package, p.Version, r,
				})
				for _, d := range p.Dependencies {
					if d.DependencyType == "Depends" || d.DependencyType == "Imports" {
						if !checkIfSkipDependency(indentation, p.Package, d.DependencyName,
							d.VersionOperator, d.VersionValue, outputList) {
							log.Info(indentation, p.Package, " → ", d.DependencyName)
							resolveDependenciesRecursively(
								outputList, d.DependencyName, d.VersionOperator,
								d.VersionValue, repositoryList, packagesFiles, recursionLevel+1,
							)
						}
					}
				}
				// Package found in repository and all dependencies processed.
				return
			}
		}
	}
	var versionConstraint string
	if versionOperator != "" && versionValue != "" {
		versionConstraint = " in version " + versionOperator + " " + versionValue
	}
	log.Warn(indentation, "Could not find package ", name, versionConstraint, " in any of the repositories.")
}

func checkIfBasePackage(name string) bool {
	var basePackages = []string{
		"base", "compiler", "datasets", "graphics", "grDevices", "grid",
		"methods", "parallel", "splines", "stats", "stats4", "tcltk", "tools",
		"translations", "utils", "R",
	}
	return stringInSlice(name, basePackages)
}

func checkIfPackageOnOutputList(name string, outputList []OutputPackage) bool {
	for _, o := range outputList {
		if name == o.Package {
			return true
		}
	}
	return false
}

func checkIfSkipDependency(indentation string, packageName string, dependencyName string,
	versionOperator string, versionValue string, outputList *[]OutputPackage) bool {
	if checkIfBasePackage(dependencyName) {
		log.Debug(indentation, "Skipping package ", dependencyName, " as it is a base R package.")
		return true
	}
	// Go through the list of dependencies added to the output list previously, to check
	// if it contains a dependency required by the currently processed package but in a version
	// that is too low.
	for i := 0; i < len(*outputList); i++ {
		if dependencyName == (*outputList)[i].Package {
			// Dependency found on the output list.
			if checkIfVersionSufficient((*outputList)[i].Version, versionOperator, versionValue) {
				return true
			}
			log.Warn(
				indentation,
				"Output list already contains dependency ", dependencyName, " version ",
				(*outputList)[i].Version, " but it is insufficient as ", packageName,
				" requires ", dependencyName, " ", versionOperator, " ", versionValue,
			)
			// Overwrite the information about the previous version of the dependency on the output list.
			// The new version of the dependency will be subsequently added by deeper recursion levels,
			// according to the higher requirements by currently processed package.
			// When generating the output renv.lock, these empty entries will be filtered out.
			(*outputList)[i].Package = ""
			(*outputList)[i].Version = ""
			(*outputList)[i].Repository = ""
			return false
		}
	}
	// Dependency not yet added to the output list.
	return false
}

func splitVersion(r rune) bool {
	return r == '.' || r == '-'
}

func stringsToInts(input []string) []int {
	var output []int
	for _, i := range input {
		j, err := strconv.Atoi(i)
		checkError(err)
		output = append(output, j)
	}
	return output
}

func checkIfVersionSufficient(availableVersionValue string, versionOperator string,
	requiredVersionValue string) bool {
	// Check if there are any version requirements at all.
	if versionOperator == "" && requiredVersionValue == "" {
		return true
	}

	// Version strings are split by "." or "-", which are treated in an equivalent way.
	availableVersionStrings := strings.FieldsFunc(availableVersionValue, splitVersion)
	requiredVersionStrings := strings.FieldsFunc(requiredVersionValue, splitVersion)

	// Make sure the length of available and required versions is the same.
	// In case trailing version component(s) are missing, add -1 in their place.
	if len(availableVersionStrings) > len(requiredVersionStrings) {
		for i := 0; i < len(availableVersionStrings)-len(requiredVersionStrings); i++ {
			requiredVersionStrings = append(requiredVersionStrings, "-1")
		}
	}
	if len(requiredVersionStrings) > len(availableVersionStrings) {
		for i := 0; i < len(requiredVersionStrings)-len(availableVersionStrings); i++ {
			availableVersionStrings = append(availableVersionStrings, "-1")
		}
	}

	availableVersion := stringsToInts(availableVersionStrings)
	requiredVersion := stringsToInts(requiredVersionStrings)

	available := "="
	// Compare up to 4 dot- or dash-separated version components.
	for i := 0; i < 4; i++ {
		if availableVersion[i] > requiredVersion[i] {
			available = ">"
			break
		} else if availableVersion[i] < requiredVersion[i] {
			available = "<"
			break
		} else if availableVersion[i] == requiredVersion[i] && len(requiredVersion) <= i+1 {
			available = "="
			break
		}
	}

	if versionOperator != ">" && versionOperator != ">=" {
		log.Error("Unknown version constraint operator: ", versionOperator)
	}
	return strings.Contains(versionOperator, available)
}
