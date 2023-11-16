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
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

func checkError(err error) {
	if err != nil {
		log.Error(err)
	}
}

// func prettyPrint(i interface{}) {
// 	s, err := json.MarshalIndent(i, "", "  ")
// 	checkError(err)
// 	log.Debug(string(s))
// }

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// ParseInput parses CLI input parameters, and returns: list of package DESCRIPTION URLs,
// list of package repository URLs (in descending priority order), a map from package
// repository alias (name) to the package repository URL, and a list of allowed types
// of missing dependencies.
func ParseInput() ([]string, []string, map[string]string, []string) {
	if len(inputPackageList) < 1 && len(inputPackages) == 0 {
		log.Fatal(
			"No packages specified. Please use the --inputPackageList flag ",
			"or supply the list under inputPackages in YAML config.",
		)
	}
	if len(inputRepositoryList) < 1 && len(inputRepositories) == 0 {
		log.Fatal(
			"No package repositories specified. Please use the --inputRepositoryList flag ",
			"or supply the list under inputRepositories in YAML config.",
		)
	}
	var packageList []string
	if len(inputPackageList) > 0 {
		log.Debug(
			"--inputPackageList CLI flag or LOCKSMITH_INPUTPACKAGELIST environment variable ",
			"has been set and is taking precedence over inputPackages key from YAML config.",
		)
		packageList = strings.Split(inputPackageList, ",")
	} else {
		packageList = inputPackages
	}
	var repositoryList []string
	if len(inputRepositoryList) > 0 {
		log.Debug(
			"--inputRepositoryList CLI flag or LOCKSMITH_INPUTREPOSITORYLIST environment variable ",
			"has been set and is taking precedence over inputRepositories key from YAML config.",
		)
		repositoryList = strings.Split(inputRepositoryList, ",")
	} else {
		repositoryList = inputRepositories
	}

	var allowedMissingDependencyTypes []string
	if len(allowIncompleteRenvLock) > 0 {
		allowedMissingDependencyTypes = strings.Split(allowIncompleteRenvLock, ",")
		log.Debug("allowedMissingDependencyTypes = ", allowedMissingDependencyTypes)
	}

	outputRepositoryMap := make(map[string]string)
	var outputRepositoryList []string
	for _, r := range repositoryList {
		if !strings.Contains(r, "=") {
			log.Fatal("Incorrect format of package repositories. Please try: 'Repo1=URL1,Repo2=URL2,...'")
		}
		repository := strings.Split(r, "=")
		outputRepositoryMap[repository[0]] = repository[1]
		outputRepositoryList = append(outputRepositoryList, repository[1])
	}
	log.Debug("inputPackageList = ", packageList)
	log.Debug("inputRepositoryList = ", outputRepositoryList)
	log.Debug("inputRepositoryMap = ", outputRepositoryMap)
	return packageList, outputRepositoryList, outputRepositoryMap, allowedMissingDependencyTypes
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

func writeJSON(filename string, j interface{}) {
	s, err := json.MarshalIndent(j, "", "  ")
	checkError(err)

	err = os.WriteFile(filename, s, 0644) //#nosec
	checkError(err)
}
