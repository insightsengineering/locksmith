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
const github = "github"
const gitlab = "gitlab"
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
	// Create temporary directory to clone the packages to be updated.
	return strings.Join(allUpdateExpressions, "|")
}

func GetPackageVersionFromDescription(descriptionFilePath string) string {
	byteValue, err := os.ReadFile(descriptionFilePath)
	checkError(err)
	cleanedDescription := CleanDescriptionOrPackagesEntry(string(byteValue), true)
	descriptionContents := make(map[string]string)
	err = yaml.Unmarshal([]byte(cleanedDescription), &descriptionContents)
	checkError(err)
	return descriptionContents["Version"]
}

func GetGitRepositoryURL(p PackageDescription) string {
	var repoURL string
	switch p.Source {
	case GitHub:
		repoURL = "https://github.com/" + p.RemoteUsername + "/" + p.RemoteRepo
	case GitLab:
		// The behavior of renv.lock is not standardized in terms of whether GitLab host address
		// starts with 'https://' or not.
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

func GetDefaultBranchSha(gitDirectory string, repoURL string,
	environmentCredentialsType string) string {
	err := os.MkdirAll(gitDirectory, os.ModePerm)
	checkError(err)
	var gitCloneOptions *git.CloneOptions
	switch {
	case environmentCredentialsType == gitlab:
		gitCloneOptions = &git.CloneOptions{
			URL: repoURL,
			Auth: &githttp.BasicAuth{
				Username: "This can be any string.",
				Password: os.Getenv("GITLAB_TOKEN"),
			},
			Depth: 1,
		}
	case environmentCredentialsType == github:
		gitCloneOptions = &git.CloneOptions{
			URL: repoURL,
			Auth: &githttp.BasicAuth{
				Username: "This can be any string.",
				Password: os.Getenv("GITHUB_TOKEN"),
			},
			Depth: 1,
		}
	default:
		gitCloneOptions = &git.CloneOptions{
			URL:   repoURL,
			Depth: 1,
		}
	}
	repository, err := git.PlainClone(gitDirectory, false, gitCloneOptions)
	if err != nil {
		return ""
	}
	// Get SHA of repository HEAD.
	ref, err := repository.Head()
	checkError(err)
	return ref.Hash().String()
}

func UpdateGitPackages(renvLock *RenvLock, updatePackageRegexp string) {
	gitUpdatesDirectory := localTempDirectory + "/git_updates"
	err := os.RemoveAll(gitUpdatesDirectory)
	checkError(err)
	err = os.MkdirAll(gitUpdatesDirectory, os.ModePerm)
	checkError(err)
	for k, v := range renvLock.Packages {
		match, err := regexp.MatchString(updatePackageRegexp, k)
		checkError(err)
		if match && (v.Source == GitLab || v.Source == GitHub) {
			log.Debug("Package ", k, " matches updated packages regexp ",
				updatePackageRegexp)
			var credentialsType string
			if v.Source == GitLab {
				credentialsType = gitlab
			} else if v.Source == GitHub {
				credentialsType = github
			}
			// Clone package's default branch.
			newPackageSha := GetDefaultBranchSha(
				gitUpdatesDirectory+k,
				GetGitRepositoryURL(v),
				credentialsType,
			)
			// Read newest package version from DESCRIPTION.
			var remoteSubdir string
			if v.RemoteSubdir != "" {
				remoteSubdir = "/" + v.RemoteSubdir
			}
			newPackageVersion := GetPackageVersionFromDescription(
				gitUpdatesDirectory + k + remoteSubdir + "/DESCRIPTION")
			// Update renv structure with new package version and SHA.
			if entry, ok := renvLock.Packages[k]; ok {
				log.Info("Updating package ", k, " version: ",
					entry.Version, " -> ", newPackageVersion,
					", SHA: ", entry.RemoteSha, " -> ", newPackageSha,
				)
				entry.Version = newPackageVersion
				entry.RemoteSha = newPackageSha
				renvLock.Packages[k] = entry
			}
		} else {
			log.Debug(
				"Package ", k, " doesn't match updated packages regexp ",
				updatePackageRegexp, " or is not a git repository.",
			)
		}
	}
}

func UpdateRenvLock(inputFileName, updatePackages string) RenvLock {
	var renvLock RenvLock
	byteValue, err := os.ReadFile(inputFileName)
	checkError(err)

	err = json.Unmarshal(byteValue, &renvLock)
	checkError(err)

	updatePackageRegex := GetPackageRegex(updatePackages)
	UpdateGitPackages(&renvLock, updatePackageRegex)
	return renvLock
}
