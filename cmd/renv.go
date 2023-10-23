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

func generateRenvLock(packageList []PackageDescription, repositoryMap map[string]string) {
	var outputRenvLock RenvLock
	outputRenvLock.R.Packages = make(map[string]PackageDescription)
	for _, p := range packageList {
		// Filter out package entries that were intentionally cleared during the process
		// of generating the package list.
		if p.Package == "" || p.Version == "" || p.Source == "" {
			continue
		}
		repositoryKey := getRepositoryKeyByValue(p.Repository, repositoryMap)
		p.Repository = repositoryKey
		outputRenvLock.R.Packages[p.Package] = p
	}
	for k, v := range repositoryMap {
		outputRenvLock.R.Repositories = append(outputRenvLock.R.Repositories, RenvLockRepository{k, v})
	}
	prettyPrint(outputRenvLock)
}

func getRepositoryKeyByValue(repositoryURL string, repositoryMap map[string]string) string {
	for k, v := range repositoryMap {
		if v == repositoryURL {
			return k
		}
	}
	return ""
}
