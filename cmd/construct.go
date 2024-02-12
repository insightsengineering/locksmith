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
)

const lowestPossiblePackageVersion = "0.0.0.0.0"

// ConstructOutputPackageList generates a list of all packages and their dependencies
// which should be included in the output renv.lock file,
// based on the list of package descriptions, and information contained in the PACKAGES files.
func ConstructOutputPackageList(packages []PackageDescription, packagesFiles map[string]PackagesFile,
	repositoryList []string, allowedMissingDependencyTypes []string) []PackageDescription {
	var outputPackageList []PackageDescription
	fatalMissingPackageVersions := make(map[string]DependencyVersion)
	nonFatalMissingPackageVersions := make(map[string]DependencyVersion)
	// Add all input packages to output list, as the packages should be downloaded from git repositories.
	for _, p := range packages {
		outputPackageList = append(outputPackageList, PackageDescription{
			p.Package, p.Version, p.Source, "", []Dependency{},
			p.RemoteType, p.RemoteHost, p.RemoteUsername, p.RemoteRepo, p.RemoteSubdir,
			p.RemoteRef, p.RemoteSha, []string{},
		})
	}
	for _, p := range packages {
		for _, d := range p.Dependencies {
			if d.DependencyType == "Depends" || d.DependencyType == "Imports" ||
				d.DependencyType == "Suggests" || d.DependencyType == "LinkingTo" {
				if !CheckIfSkipDependency("", p.Package, d.DependencyName,
					d.VersionOperator, d.VersionValue, &outputPackageList) {
					log.Info(p.Package, " → ", d.DependencyName, " (", d.DependencyType, ")")
					ResolveDependenciesRecursively(
						&outputPackageList, d.DependencyName, d.VersionOperator,
						d.VersionValue, d.DependencyType, allowedMissingDependencyTypes,
						repositoryList, packagesFiles, 1, fatalMissingPackageVersions,
						nonFatalMissingPackageVersions,
					)
				}
			}
		}
	}
	errorsString := "Packages not found in any repository:\n"
	for packageName, versionConstraint := range nonFatalMissingPackageVersions {
		var versionConstraintString string
		if versionConstraint.VersionValue != lowestPossiblePackageVersion {
			versionConstraintString = (" " +
				versionConstraint.VersionOperator + " " +
				versionConstraint.VersionValue)
		}
		errorsString += packageName + versionConstraintString + "\n"
	}
	if len(errorsString) > len("Packages not found in any repository:\n") {
		log.Error(errorsString)
	}
	errorsString = "Packages not found in any repository:\n"
	for packageName, versionConstraint := range fatalMissingPackageVersions {
		var versionConstraintString string
		if versionConstraint.VersionValue != lowestPossiblePackageVersion {
			versionConstraintString = (" " +
				versionConstraint.VersionOperator + " " +
				versionConstraint.VersionValue)
		}
		errorsString += packageName + versionConstraintString + "\n"
	}
	if len(errorsString) > len("Packages not found in any repository:\n") {
		log.Fatal(errorsString)
	}
	return outputPackageList
}

// ResolveDependenciesRecursively checks dependencies of the package, and their required versions.
// Checks if the required version is already included in the output package list
// (later used to generate the renv.lock), or if the dependency should be downloaded from a package repository.
// Repeats the process recursively for all dependencies not yet processed.
func ResolveDependenciesRecursively(outputList *[]PackageDescription, name string, versionOperator string,
	versionValue string, dependencyType string, allowedMissingDependencyTypes []string,
	repositoryList []string, packagesFiles map[string]PackagesFile, recursionLevel int,
	fatalMissingPackageVersions map[string]DependencyVersion,
	nonFatalMissingPackageVersions map[string]DependencyVersion) {
	var indentation string
	for i := 0; i < recursionLevel; i++ {
		indentation += "  "
	}
	for _, r := range repositoryList {
		// Check if the package is present in the PACKAGES file for the repository.
		for _, p := range packagesFiles[r].Packages {
			if p.Package == name {
				if r != repositoryList[0] {
					log.Warn(indentation, name, " not found in top repository.")
				}
				// Check if package in the repository is available in sufficient version.
				if !CheckIfVersionSufficient(p.Version, versionOperator, versionValue) {
					log.Warn(
						indentation, p.Package, " in repository ", r,
						" is available in version ", p.Version,
						" which is insufficient according to requirement ",
						versionOperator, " ", versionValue,
					)
					// Try to retrieve the package from the next repository.
					continue
				}
				// Add package to the output list.
				// Repository is saved as an URL, and will be changed into an alias
				// during the processing of output package list into renv.lock file.
				*outputList = append(*outputList, PackageDescription{
					p.Package, p.Version, "Repository", r, []Dependency{},
					"", "", "", "", "", "", "", []string{},
				})
				for _, d := range p.Dependencies {
					if d.DependencyType == "Depends" || d.DependencyType == "Imports" ||
						d.DependencyType == "LinkingTo" {
						if !CheckIfSkipDependency(indentation, p.Package, d.DependencyName,
							d.VersionOperator, d.VersionValue, outputList) {
							log.Info(
								indentation, p.Package, " → ", d.DependencyName,
								" (", d.DependencyType, ")",
							)
							ResolveDependenciesRecursively(
								outputList, d.DependencyName, d.VersionOperator, d.VersionValue,
								d.DependencyType, allowedMissingDependencyTypes, repositoryList,
								packagesFiles, recursionLevel+1, fatalMissingPackageVersions,
								nonFatalMissingPackageVersions,
							)
						}
					}
				}
				// Package found in repository and all dependencies processed.
				return
			}
		}
	}
	ProcessMissingPackage(
		indentation, name, versionOperator, versionValue, dependencyType,
		allowedMissingDependencyTypes, fatalMissingPackageVersions,
		nonFatalMissingPackageVersions,
	)
}

// ProcessMissingPackage saves information about missing packages (dependencies) and their versions.
// This information is later reported to the user, together with optionally exiting the application
// with failed status, depending on the types of missing dependencies and the configuration
// provided by --allowIncompleteRenvLock flag.
func ProcessMissingPackage(indentation string, packageName string, versionOperator string,
	versionValue string, dependencyType string, allowedMissingDependencyTypes []string,
	fatalMissingPackageVersions map[string]DependencyVersion,
	nonFatalMissingPackageVersions map[string]DependencyVersion) {
	var versionConstraint string
	if versionOperator != "" && versionValue != "" {
		versionConstraint = " (version " + versionOperator + " " + versionValue + ")"
	} else {
		versionOperator = ">="
		versionValue = lowestPossiblePackageVersion
	}
	message := "Could not find package " + packageName + versionConstraint + " in any of the repositories.\n"
	if stringInSlice(dependencyType, allowedMissingDependencyTypes) {
		log.Warn(indentation + message)
		val, ok := nonFatalMissingPackageVersions[packageName]
		if !ok || (ok && !CheckIfVersionSufficient(val.VersionValue, versionOperator, versionValue)) {
			// This is the first time we see this package as a missing dependency, or
			// some other package already requires this missing dependency in a lower version
			// and the currently processed package requires it in a higher version,
			// so the current requirement is more important.
			nonFatalMissingPackageVersions[packageName] = DependencyVersion{versionOperator, versionValue}
			log.Trace("Adding package ", packageName, " ", versionOperator, " ", versionValue, " to missing packages list.")
		}
	} else {
		log.Error(indentation + message)
		val, ok := fatalMissingPackageVersions[packageName]
		if !ok || (ok && !CheckIfVersionSufficient(val.VersionValue, versionOperator, versionValue)) {
			// See a comment above for explanation of this condition.
			fatalMissingPackageVersions[packageName] = DependencyVersion{versionOperator, versionValue}
			log.Trace("Adding package ", packageName, " ", versionOperator, " ", versionValue, " to missing packages list.")
		}
	}
}

// CheckIfBasePackage checks whether the package should be treated as a base R package
// (included in every R installation) or if it should be treated as a dependency
// to be downloaded from a package repository.
func CheckIfBasePackage(name string) bool {
	var basePackages = []string{
		"base", "compiler", "datasets", "graphics", "grDevices", "grid",
		"methods", "parallel", "splines", "stats", "stats4", "tcltk", "tools",
		"translations", "utils", "R",
	}
	return stringInSlice(name, basePackages)
}

// CheckIfSkipDependency checks if processing of the package (dependency) should be skipped.
// Dependency should be skipped if it is a base R package, or has already been added to output
// package list (later used to generate the renv.lock).
func CheckIfSkipDependency(indentation string, packageName string, dependencyName string,
	versionOperator string, versionValue string, outputList *[]PackageDescription) bool {
	if CheckIfBasePackage(dependencyName) {
		log.Trace(indentation, "Skipping package ", dependencyName, " as it is a base R package.")
		return true
	}
	// Go through the list of dependencies added to the output list previously, to check
	// if it contains a dependency required by the currently processed package but in a version
	// that is too low.
	for i := 0; i < len(*outputList); i++ {
		if dependencyName == (*outputList)[i].Package {
			// Dependency found on the output list.
			if CheckIfVersionSufficient((*outputList)[i].Version, versionOperator, versionValue) {
				var requirementMessage string
				if versionOperator != "" && versionValue != "" {
					requirementMessage = " according to the requirement " + versionOperator + " " + versionValue
				} else {
					requirementMessage = " since no required version has been specified."
				}
				log.Debug(
					indentation, "Output list already contains ", dependencyName, " version ",
					(*outputList)[i].Version, " which is sufficient for ", packageName,
					requirementMessage,
				)
				return true
			}
			log.Warn(
				indentation,
				"Output list already contains ", dependencyName, " but the version ",
				(*outputList)[i].Version, " is insufficient as ", packageName,
				" requires ", dependencyName, " ", versionOperator, " ", versionValue,
			)
			// Overwrite the information about the previous version of the dependency on the output list.
			// The new version of the dependency will be subsequently added by deeper recursion levels,
			// according to the higher requirements by currently processed package.
			// When generating the output renv.lock, these empty entries will be filtered out.
			(*outputList)[i].Package = ""
			(*outputList)[i].Version = ""
			(*outputList)[i].Source = ""
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

// CheckIfVersionSufficient checks if availableVersionValue fulfills the requirement
// expressed by versionOperator ('>=' or '>') and requiredVersionValue.
func CheckIfVersionSufficient(availableVersionValue string, versionOperator string,
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
	// Compare up to 5 dot- or dash-separated version components.
	// Examples of packages with 5 version components: RcppEigen, RcppArmadillo.
	for i := 0; i < 5; i++ {
		breakLoop := false
		switch {
		case availableVersion[i] > requiredVersion[i]:
			available = ">"
			breakLoop = true
		case availableVersion[i] < requiredVersion[i]:
			available = "<"
			breakLoop = true
		case availableVersion[i] == requiredVersion[i] && len(requiredVersion) <= i+1:
			available = "="
			breakLoop = true
		}
		if breakLoop {
			break
		}
	}

	if versionOperator != ">" && versionOperator != ">=" {
		log.Error("Unknown version constraint operator: ", versionOperator)
	}
	return strings.Contains(versionOperator, available)
}
