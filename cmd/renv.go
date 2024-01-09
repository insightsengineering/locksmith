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
	"regexp"
	"sort"
	"strings"

	git "github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	yaml "gopkg.in/yaml.v3"
)

const GitHub = "GitHub"
const GitLab = "GitLab"
const https = "https://"

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

// GetPackageRegex processes the comma-separated expression with wildcards indicating which packages
// should be updated and returns a real regex.
func GetPackageRegex(updatedPackages string) string {
	splitUpdatePackages := strings.Split(updatedPackages, ",")
	var allUpdateExpressions []string
	// For each comma-separated wildcard expression convert "." and "*"
	// characters to regexp equivalent.
	for _, singleRegexp := range splitUpdatePackages {
		singleRegexp = strings.ReplaceAll(singleRegexp, `.`, `\.`)
		singleRegexp = strings.ReplaceAll(singleRegexp, "*", ".*")
		allUpdateExpressions = append(allUpdateExpressions, "^"+singleRegexp+"$")
	}
	return strings.Join(allUpdateExpressions, "|")
}

// GetPackageVersionFromDescription reads the DESCRIPTION file located in descriptionFilePath
// and returns the package version.
func GetPackageVersionFromDescription(descriptionFilePath string) string {
	byteValue, err := os.ReadFile(descriptionFilePath)
	checkError(err)
	cleanedDescription := CleanDescriptionOrPackagesEntry(string(byteValue), true)
	descriptionContents := make(map[string]string)
	err = yaml.Unmarshal([]byte(cleanedDescription), &descriptionContents)
	checkError(err)
	version := descriptionContents["Version"]
	return version
}

// GetGitRepositoryURL reads the PackageDescription struct corresponding to a single package
// in the renv.lock and returns the git repository URL from which the package should be cloned.
func GetGitRepositoryURL(p PackageDescription) string {
	var repoURL string
	switch p.Source {
	case GitHub:
		repoURL = "https://github.com/" + p.RemoteUsername + "/" + p.RemoteRepo
	case GitLab:
		// The behavior of renv.lock is not standardized in terms of whether GitLab
		// host address starts with 'https://' or not.
		var remoteHost string
		if strings.HasPrefix(p.RemoteHost, https) {
			remoteHost = p.RemoteHost
		} else {
			remoteHost = https + p.RemoteHost
		}
		repoURL = remoteHost + "/" + p.RemoteUsername + "/" + p.RemoteRepo
	}
	return repoURL
}

// GetDefaultBranchSha clones the git repository located at repoURL to gitDirectory, using Personal
// Access Tokens from LOCKSMITH_GITLABTOKEN or LOCKSMITH_GITHUBTOKEN environment variables.
// It returns the commit SHA of the HEAD of default branch and the name of the default branch,
// or empty string in case of error.
func GetDefaultBranchSha(gitDirectory string, repoURL string,
	environmentCredentialsType string) (string, string) {
	err := os.MkdirAll(gitDirectory, os.ModePerm)
	checkError(err)
	var gitCloneOptions *git.CloneOptions
	switch {
	case environmentCredentialsType == GitLab:
		gitCloneOptions = &git.CloneOptions{
			URL: repoURL,
			Auth: &githttp.BasicAuth{
				Username: "This can be any string.",
				Password: os.Getenv("LOCKSMITH_GITLABTOKEN")},
			Depth: 1,
		}
	case environmentCredentialsType == GitHub:
		gitCloneOptions = &git.CloneOptions{
			URL: repoURL,
			Auth: &githttp.BasicAuth{
				Username: "This can be any string.",
				Password: os.Getenv("LOCKSMITH_GITHUBTOKEN")},
			Depth: 1,
		}
	}
	repository, err := git.PlainClone(gitDirectory, false, gitCloneOptions)
	if err != nil {
		log.Error("Error while cloning ", repoURL, ": ", err)
		return "", ""
	}
	// Get SHA of repository HEAD.
	ref, err := repository.Head()
	checkError(err)
	return ref.Hash().String(), ref.Name().String()
}

// UpdateGitPackages iterates through the packages in renv.lock and updates the entries
// corresponding to packages stored in git repositories. Package version and latest commit SHA
// are updated in the renvLock struct. Only packages matching the updatePackageRegexp are updated.
func UpdateGitPackages(renvLock *RenvLock, updatePackageRegexp string,
	getDefaultBranchShaFunction func(string, string, string) (string, string),
	gitUpdatesDirectory string) {
	for k, v := range renvLock.Packages {
		match, err := regexp.MatchString(updatePackageRegexp, k)
		checkError(err)
		if !match || (v.Source != GitLab && v.Source != GitHub) {
			log.Trace("Package ", k, " doesn't match updated packages regexp ",
				updatePackageRegexp, " or is not a git repository.")
			continue
		}
		log.Trace("Package ", k, " matches updated packages regexp ",
			updatePackageRegexp)
		// Read default branch HEAD SHA from git.
		newPackageSha, defaultBranchName := getDefaultBranchShaFunction(
			gitUpdatesDirectory+k, GetGitRepositoryURL(v), v.Source,
		)
		var remoteSubdir string
		if v.RemoteSubdir != "" {
			remoteSubdir = "/" + v.RemoteSubdir
		}
		// Read newest package version from DESCRIPTION.
		newPackageVersion := GetPackageVersionFromDescription(
			gitUpdatesDirectory + k + remoteSubdir + "/DESCRIPTION")
		if entry, ok := renvLock.Packages[k]; ok && newPackageSha != "" && newPackageVersion != "" {
			if newPackageVersion != entry.Version && newPackageSha != entry.RemoteSha {
				// Update the renv structure with new version only if the current
				// default branch SHA and current package version could be retrieved
				// and they are different than old ones.
				log.Info("Updating package ", k, " version: ",
					entry.Version, " → ", newPackageVersion,
					", SHA: ", entry.RemoteSha, " → ", newPackageSha)
				entry.Version = newPackageVersion
				entry.RemoteSha = newPackageSha
				// Set the default branch name to ensure consistency,
				// in case previously renv.lock pointed to e.g. a tag.
				entry.RemoteRef = defaultBranchName
				renvLock.Packages[k] = entry
			}
		}
	}
}

// UpdateRepositoryPackages iterates through the packages in renv.lock and updates the entries
// corresponding to packages downloaded from CRAN-like repositories. Package version is updated
// in the renvLock struct. Only packages matching the updatePackageRegexp are updated.
func UpdateRepositoryPackages(renvLock *RenvLock, updatePackageRegexp string,
	packagesFiles map[string]PackagesFile) {
	for k, v := range renvLock.Packages {
		match, err := regexp.MatchString(updatePackageRegexp, k)
		checkError(err)
		if !match || v.Source == GitLab || v.Source == GitHub {
			log.Trace("Package ", k, " doesn't match updated packages regexp ",
				updatePackageRegexp, " or is not a git repository.")
			continue
		}
		log.Trace("Package ", k, " matches updated packages regexp ",
			updatePackageRegexp)
		var repositoryPackagesFile PackagesFile
		repositoryName := v.Repository
		repositoryPackagesFile, ok := packagesFiles[repositoryName]
		if !ok {
			log.Error(`Could not retrieve PACKAGES for "`, repositoryName, `" repository `,
				`(referenced by `, k, `). Attempting to use CRAN's PACKAGES as a fallback.`)
			repositoryPackagesFile = packagesFiles["CRAN"]
			repositoryName = "CRAN"
		}
		var newPackageVersion string
		for _, singlePackage := range repositoryPackagesFile.Packages {
			if singlePackage.Package == k {
				newPackageVersion = singlePackage.Version
				continue
			}
		}
		if newPackageVersion == "" {
			log.Error(`Could not find package `, k, ` in PACKAGES file for "`, repositoryName, `" repository.`)
			continue
		}
		if entry, ok := renvLock.Packages[k]; ok {
			if newPackageVersion != entry.Version {
				log.Info("Updating package ", k, " version: ",
					entry.Version, " → ", newPackageVersion)
				entry.Version = newPackageVersion
				renvLock.Packages[k] = entry
			}
		}
	}
}

// GetPackagesFiles downloads PACKAGES files from repositories defined in the renv.lock header.
// It returns a map from the repository name (as defined in the renv.lock header) to the PackagesFile
// struct representing repository's PACKAGES file.
func GetPackagesFiles(renvLock RenvLock) map[string]PackagesFile {
	repositoryPackagesFiles := make(map[string]PackagesFile)
	for _, repository := range renvLock.R.Repositories {
		packagesFileContent := GetPackagesFileContent(repository.URL, DownloadTextFile)
		packagesFile := ProcessPackagesFile(packagesFileContent)
		repositoryPackagesFiles[repository.Name] = packagesFile
	}

	// Check if the PACKAGES file from a repository named CRAN has been downloaded.
	_, ok := repositoryPackagesFiles["CRAN"]
	if !ok {
		// If not, save CRAN's PACKAGES file to be used as a fallback, for packages which
		// (according to renv.lock) should be downloaded from a repository not defined in
		// the renv.lock header.
		_, _, cranPackagesContent := DownloadTextFile(
			"https://cloud.r-project.org/src/contrib/PACKAGES", make(map[string]string),
		)
		cranPackagesFile := ProcessPackagesFile(cranPackagesContent)
		repositoryPackagesFiles["CRAN"] = cranPackagesFile
	}
	return repositoryPackagesFiles
}

// UpdateRenvLock reads the renv.lock from inputFileName. It then retrieves the information
// about the newest package versions from respective repositories (CRAN-like or git repositories)
// from which the packages should be downloaded according to the renv.lock.
// It returns the RenvLock struct represeting the renv.lock with updated package versions.
func UpdateRenvLock(inputFileName, updatePackages string) RenvLock {
	// Read renv.lock into RenvLock struct.
	var renvLock RenvLock
	byteValue, err := os.ReadFile(inputFileName)
	checkError(err)
	err = json.Unmarshal(byteValue, &renvLock)
	checkError(err)

	updatePackageRegex := GetPackageRegex(updatePackages)

	// Remove and recreate directories where temporary clones of git repositories
	// used to get the newest default branch SHA will be stored.
	gitUpdatesDirectory := localTempDirectory + "/git_updates/"
	err = os.RemoveAll(gitUpdatesDirectory)
	checkError(err)
	err = os.MkdirAll(gitUpdatesDirectory, os.ModePerm)
	checkError(err)

	UpdateGitPackages(&renvLock, updatePackageRegex, GetDefaultBranchSha, gitUpdatesDirectory)
	repositoryPackagesFiles := GetPackagesFiles(renvLock)
	UpdateRepositoryPackages(&renvLock, updatePackageRegex, repositoryPackagesFiles)
	return renvLock
}
