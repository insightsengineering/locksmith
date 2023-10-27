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
	"sort"
)

// GenerateRenvLock generates renv.lock file structure which can be then saved as a JSON file.
// It uses a list of package data created by ConstructOutputPackageList, and the map of
// package repositories containing the packages.
func GenerateRenvLock(packageList []PackageDescription, repositoryMap map[string]string) RenvLock {
	var outputRenvLock RenvLock
	outputRenvLock.Packages = make(map[string]PackageDescription)
	for _, p := range packageList {
		// Filter out package entries that were intentionally cleared during the process
		// of generating the package list.
		if p.Package == "" || p.Version == "" || p.Source == "" {
			continue
		}
		// Replace package repository URL with package repository alias/name.
		repositoryKey := GetRepositoryKeyByValue(p.Repository, repositoryMap)
		p.Repository = repositoryKey
		outputRenvLock.Packages[p.Package] = p
	}
	// As the repository map is not sorted, in order to generate predictable output
	// we have to process the repository names in sorted order.
	var repositoryKeys []string
	for k := range repositoryMap {
		repositoryKeys = append(repositoryKeys, k)
	}
	sort.Strings(repositoryKeys)
	for _, k := range repositoryKeys {
		outputRenvLock.R.Repositories = append(outputRenvLock.R.Repositories, RenvLockRepository{k, repositoryMap[k]})
	}
	return outputRenvLock
}

// GetRepositoryKeyByValue searches for repository URL in repositoryMap and returns
// the name (alias) of that repository which will then be used in output renv.lock file.
func GetRepositoryKeyByValue(repositoryURL string, repositoryMap map[string]string) string {
	for k, v := range repositoryMap {
		if v == repositoryURL {
			return k
		}
	}
	return ""
}
