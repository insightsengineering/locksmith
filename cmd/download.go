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
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type GitLabAPIResponse struct {
	PathWithNamespace string `json:"path_with_namespace"`
}

type GitLabTagOrBranchResponse struct {
	Commit GitLabCommit `json:"commit"`
}

type GitLabCommit struct {
	ID string `json:"id"`
}

type GitHubTagOrBranchResponse struct {
	Object GitHubObject `json:"object"`
}

type GitHubObject struct {
	Sha string `json:"sha"`
}

// Returns HTTP status code for downloaded file, number of bytes in downloaded content,
// and the downloaded content itself.
func downloadTextFile(url string, parameters map[string]string) (int, int64, string) { // #nosec G402
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	checkError(err)
	for k, v := range parameters {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	checkError(err)

	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			body, err2 := io.ReadAll(resp.Body)
			checkError(err2)
			return resp.StatusCode, resp.ContentLength, string(body)
		}
	}
	return -1, 0, ""
}

// Retrieve information about GitLab repository (project path, repository name and commit SHA) from
// projectURL repository GitLab API endpoint.
func getGitLabProjectAndSha(projectURL string, remoteRef string, token map[string]string,
	downloadFileFunction func(string, map[string]string) (int, int64, string)) (string, string, string) {
	var remoteUsername, remoteRepo, remoteSha string
	log.Trace("Downloading project information from ", projectURL)
	statusCode, _, projectDataResponse := downloadFileFunction(projectURL, token)
	if statusCode == 200 {
		var projectData GitLabAPIResponse
		err := json.Unmarshal([]byte(projectDataResponse), &projectData)
		checkError(err)
		projectPath := strings.Split(projectData.PathWithNamespace, "/")
		projectPathLength := len(projectPath)
		remoteUsername = strings.Join(projectPath[:projectPathLength-1], "/")
		remoteRepo = projectPath[projectPathLength-1]
	} else {
		log.Warn("An error occurred while retrieving project information from ", projectURL)
	}
	match, errMatch := regexp.MatchString(`v\d+(\.\d+)*`, remoteRef)
	var urlPath string
	if match {
		log.Trace("remoteRef = ", remoteRef, " matches tag name regexp.")
		urlPath = "tags"
	} else {
		log.Trace("remoteRef = ", remoteRef, " doesn't match tag name regexp.")
		urlPath = "branches"
	}
	tagOrBranchURL := projectURL + "/repository/" + urlPath + "/" + remoteRef
	statusCode, _, tagOrBranchDataResponse := downloadFileFunction(tagOrBranchURL, token)
	if statusCode == 200 {
		var tagOrBranchData GitLabTagOrBranchResponse
		err := json.Unmarshal([]byte(tagOrBranchDataResponse), &tagOrBranchData)
		checkError(err)
		remoteSha = tagOrBranchData.Commit.ID
	} else {
		log.Warn("An error occurred while retrieving tag information from ", tagOrBranchURL)
	}
	checkError(errMatch)
	return remoteUsername, remoteRepo, remoteSha
}

// Retrieve SHA of the remoteRef from the 'remoteUsername/remoteRepo' GitHub repository.
func getGitHubSha(remoteUsername string, remoteRepo string, remoteRef string, token map[string]string,
	downloadFileFunction func(string, map[string]string) (int, int64, string)) string {
	var remoteSha string
	log.Trace("Downloading data for GitHub project ", remoteUsername, "/", remoteRepo)
	match, errMatch := regexp.MatchString(`v\d+(\.\d+)*`, remoteRef)
	var urlPath string
	if match {
		log.Trace("remoteRef = ", remoteRef, " matches tag name regexp.")
		urlPath = "tags"

	} else {
		log.Trace("remoteRef = ", remoteRef, " doesn't match tag name regexp.")
		urlPath = "heads"
	}
	tagOrBranchURL := "https://api.github.com/repos/" + remoteUsername + "/" + remoteRepo + "/git/ref/" + urlPath + "/" + remoteRef
	statusCode, _, tagDataResponse := downloadFileFunction(tagOrBranchURL, token)
	if statusCode == 200 {
		var tagOrBranchData GitHubTagOrBranchResponse
		err := json.Unmarshal([]byte(tagDataResponse), &tagOrBranchData)
		checkError(err)
		remoteSha = tagOrBranchData.Object.Sha
	} else {
		log.Warn("An error occurred while retrieving data from ", tagOrBranchURL)
	}
	checkError(errMatch)
	return remoteSha
}

// Get information about packages stored in git repositories.
func processDescriptionURL(descriptionURL string,
	downloadFileFunction func(string, map[string]string) (int, int64, string),
) (map[string]string, string, string, string, string, string, string, string, string) {
	token := make(map[string]string)
	var remoteType, remoteRef, remoteHost, remoteUsername, remoteRepo string
	var remoteSubdir, remoteSha, packageSource string
	if strings.HasPrefix(descriptionURL, "https://raw.githubusercontent.com") {
		// Expecting URL in form:
		// https://raw.githubusercontent.com/<organization>/<repo-name>/<ref-name>/<optional-subdirectories>/DESCRIPTION
		token["Authorization"] = "token " + gitHubToken
		remoteType = "github"
		packageSource = "GitHub"
		shorterURL := strings.TrimPrefix(descriptionURL, "https://raw.githubusercontent.com/")
		remoteHost = "api.github.com"
		remoteUsername = strings.Split(shorterURL, "/")[0]
		remoteRepo = strings.Split(shorterURL, "/")[1]
		remoteRef = strings.Split(shorterURL, "/")[2]
		remoteSha = getGitHubSha(remoteUsername, remoteRepo, remoteRef, token, downloadFileFunction)
		// Check whether package is stored in a subdirectory of the git repository.
		for i, j := range strings.Split(shorterURL, "/") {
			if j == "DESCRIPTION" {
				remoteSubdir = strings.Join(strings.Split(shorterURL, "/")[3:i], "/")
			}
		}
	} else {
		// Expecting URL in form:
		// https://example.gitlab.com/api/v4/projects/<project-id>/repository/files/<optional-subdirectories>/DESCRIPTION/raw?ref=<ref-name>
		// <optional-subdirectories> contains '/' encoded as '%2F'
		re := regexp.MustCompile(`ref=.*$`)
		token["Private-Token"] = gitLabToken
		remoteType = "gitlab"
		packageSource = "GitLab"
		shorterURL := strings.TrimPrefix(descriptionURL, "https://")
		remoteHost = "https://" + strings.Split(shorterURL, "/")[0]
		remoteRef = strings.TrimPrefix(re.FindString(descriptionURL), "ref=")
		projectURL := "https://" + strings.Join(strings.Split(shorterURL, "/")[0:5], "/")
		descriptionPath := strings.Split(shorterURL, "/")[7]
		// Check whether package is stored in a subdirectory of the git repository.
		if strings.Contains(descriptionPath, "%2F") {
			descriptionPath := strings.Split(strings.ReplaceAll(descriptionPath, "%2F", "/"), "/")
			remoteSubdir = strings.Join(descriptionPath[:len(descriptionPath)-1], "/")
		}
		remoteUsername, remoteRepo, remoteSha = getGitLabProjectAndSha(projectURL, remoteRef, token, downloadFileFunction)
	}
	return token, remoteType, packageSource, remoteHost, remoteUsername, remoteRepo, remoteSubdir, remoteRef, remoteSha
}

// Downloads DESCRIPTION files from the list of supplied URLs.
// Returns a list of structures representing the contents of DESCRIPTION file for the packages,
// and the git repositories storing the packages.
func downloadDescriptionFiles(packageDescriptionList []string,
	downloadFileFunction func(string, map[string]string) (int, int64, string)) []DescriptionFile {
	var inputDescriptionFiles []DescriptionFile
	for _, packageDescriptionURL := range packageDescriptionList {
		token, remoteType, packageSource, remoteHost, remoteUsername, remoteRepo, remoteSubdir, remoteRef, remoteSha :=
			processDescriptionURL(packageDescriptionURL, downloadFileFunction)
		log.Info(
			"Downloading ", packageDescriptionURL, "\nremoteType = ", remoteType,
			", remoteUsername = ", remoteUsername, ", remoteRepo = ", remoteRepo,
			", remoteSubdir = ", remoteSubdir, ", remoteRef = ", remoteRef, ", remoteSha = ", remoteSha,
		)
		statusCode, _, descriptionContent := downloadFileFunction(packageDescriptionURL, token)
		if statusCode == 200 {
			inputDescriptionFiles = append(
				inputDescriptionFiles,
				DescriptionFile{
					descriptionContent, packageSource, remoteType, remoteHost,
					remoteUsername, remoteRepo, remoteSubdir, remoteRef, remoteSha,
				},
			)
		} else {
			log.Warn(
				"An error occurred while downloading ", packageDescriptionURL,
				" Please make sure you provided an access token (in LOCKSMITH_GITHUBTOKEN ",
				"or LOCKSMITH_GITLABTOKEN environment variable).",
			)
		}
	}
	return inputDescriptionFiles
}

// Downloads PACKAGES files from repository URLs specified in the repositoryList.
// Returns a map from repository URL to the string with the contents of PACKAGES file
// for that repository.
func downloadPackagesFiles(repositoryList []string,
	downloadFileFunction func(string, map[string]string) (int, int64, string)) map[string]string {
	inputPackagesFiles := make(map[string]string)
	for _, repository := range repositoryList {
		statusCode, _, packagesFileContent := downloadFileFunction(repository+"/src/contrib/PACKAGES", map[string]string{})
		if statusCode == 200 {
			inputPackagesFiles[repository] = packagesFileContent
		} else {
			log.Warn("An error occurred while downloading ", repository+"/src/contrib/PACKAGES")
		}
	}
	return inputPackagesFiles
}
