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
	"io"
	"net/http"
	"strings"
	"os"
	"regexp"
	"encoding/json"
)

type GitLabAPIResponse struct {
	PathWithNamespace string `json:"path_with_namespace"`
}

type GitLabTagOrBranchResponse struct {
	Commit GitLabCommit `json:"commit"`
}

type GitLabCommit struct {
	Id string `json:"id"`
}

type GitHubTagOrBranchResponse struct {
	Object GitHubObject `json:"object"`
}

type GitHubObject struct {
	Sha string `json:"sha"`
}

// Returns HTTP status code for downloaded file, number of bytes in downloaded content,
// and the downloaded content itself.
func downloadTextFile(url string, parameters map[string]string) (int, int64, string) {
	tr := &http.Transport{ // #nosec
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec
	} // #nosec
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

func getGitLabProjectAndSha(projectURL string, remoteRef string, token map[string]string) (string, string, string) {
	var remoteUsername string
	var remoteRepo string
	var remoteSha string
	log.Debug("Downloading project information from ", projectURL)
	statusCode, _, projectDataResponse := downloadTextFile(projectURL, token)
	if statusCode == 200 {
		var projectData GitLabAPIResponse
		json.Unmarshal([]byte(projectDataResponse), &projectData)
		prettyPrint(projectData)
		projectPath := strings.Split(projectData.PathWithNamespace, "/")
		projectPathLength := len(projectPath)
		remoteUsername = strings.Join(projectPath[:projectPathLength-1], "/")
		remoteRepo = projectPath[projectPathLength-1]
	} else {
		log.Warn("An error occurred while retrieving project information from ", projectURL)
	}
	match, err := regexp.MatchString(`v\d+(\.\d+)*`, remoteRef)
	if match {
		log.Debug("remoteRef = ", remoteRef, " matches tag name regexp.")
		tagURL := projectURL + "/repository/tags/" + remoteRef
		statusCode, _, tagDataResponse := downloadTextFile(tagURL, token)
		if statusCode == 200 {
			var tagData GitLabTagOrBranchResponse
			json.Unmarshal([]byte(tagDataResponse), &tagData)
			prettyPrint(tagData)
			remoteSha = tagData.Commit.Id
		} else {
			log.Warn("An error occurred while retrieving tag information from ", tagURL)
		}
	} else {
		log.Debug("remoteRef = ", remoteRef, " doesn't match tag name regexp.")
		branchURL := projectURL + "/repository/branches/" + remoteRef
		statusCode, _, branchDataResponse := downloadTextFile(branchURL, token)
		if statusCode == 200 {
			var branchData GitLabTagOrBranchResponse
			json.Unmarshal([]byte(branchDataResponse), &branchData)
			prettyPrint(branchData)
			remoteSha = branchData.Commit.Id
		} else {
			log.Warn("An error occurred while retrieving branch information from ", branchURL)
		}
	}
	checkError(err)
	return remoteUsername, remoteRepo, remoteSha
}

func getGitHubSha(remoteUsername string, remoteRepo string, remoteRef string,
	token map[string]string) string {
	var remoteSha string
	log.Debug("Downloading information for GitHub project ", remoteUsername, "/", remoteRepo)
	match, err := regexp.MatchString(`v\d+(\.\d+)*`, remoteRef)
	if match {
		log.Debug("remoteRef = ", remoteRef, " matches tag name regexp.")
		tagURL := "https://api.github.com/repos/" + remoteUsername + "/" + remoteRepo + "/git/ref/tags/" + remoteRef
		statusCode, _, tagDataResponse := downloadTextFile(tagURL, token)
		if statusCode == 200 {
			var tagData GitHubTagOrBranchResponse
			json.Unmarshal([]byte(tagDataResponse), &tagData)
			prettyPrint(tagData)
			remoteSha = tagData.Object.Sha
		} else {
			log.Warn("An error occurred while retrieving tag information from ", tagURL)
		}
	} else {
		log.Debug("remoteRef = ", remoteRef, " doesn't match tag name regexp.")
		branchURL := "https://api.github.com/repos/" + remoteUsername + "/" + remoteRepo + "/git/ref/heads/" + remoteRef
		statusCode, _, branchDataResponse := downloadTextFile(branchURL, token)
		if statusCode == 200 {
			var branchData GitHubTagOrBranchResponse
			json.Unmarshal([]byte(branchDataResponse), &branchData)
			prettyPrint(branchData)
			remoteSha = branchData.Object.Sha
		} else {
			log.Warn("An error occurred while retrieving branch information from ", branchURL)
		}
	}
	checkError(err)
	return remoteSha
}

func processDescriptionURL(descriptionURL string) (map[string]string, string, string, string, string, string, string, string, string) {
	token := make(map[string]string)
	var remoteType string
	var remoteRef string
	var remoteHost string
	var remoteUsername string
	var remoteRepo string
	var remoteSubdir string
	var remoteSha string
	var packageSource string
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
		remoteSha = getGitHubSha(remoteUsername, remoteRepo, remoteRef, token)
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
		if strings.Contains(strings.Split(shorterURL, "/")[7], "%2F") {
			// DESCRIPTION is in a directory within the repository.
			descriptionPath := strings.Split(strings.ReplaceAll(strings.Split(shorterURL, "/")[7], "%2F", "/"), "/")
			remoteSubdir = strings.Join(descriptionPath[:len(descriptionPath)-1], "/")
		}
		remoteUsername, remoteRepo, remoteSha = getGitLabProjectAndSha(projectURL, remoteRef, token)
	}
	return token, remoteType, packageSource, remoteHost, remoteUsername, remoteRepo, remoteSubdir, remoteRef, remoteSha
}

func downloadDescriptionFiles(packageDescriptionList []string) []DescriptionFile {
	var inputDescriptionFiles []DescriptionFile
	for _, packageDescriptionURL := range packageDescriptionList {
		token, remoteType, packageSource, remoteHost, remoteUsername, remoteRepo, remoteSubdir, remoteRef, remoteSha :=
			processDescriptionURL(packageDescriptionURL)
		log.Info(
			"Downloading ", packageDescriptionURL, "\nremoteType = ", remoteType,
			", remoteUsername = ", remoteUsername, ", remoteRepo = ", remoteRepo,
			", remoteSubdir = ", remoteSubdir, ", remoteRef = ", remoteRef, ", remoteSha = ", remoteSha,
		)
		statusCode, _, descriptionContent := downloadTextFile(packageDescriptionURL, token)
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
	os.Exit(0)
	return inputDescriptionFiles
}

func downloadPackagesFiles(repositoryList []string) map[string]string {
	inputPackagesFiles := make(map[string]string)
	for _, repository := range repositoryList {
		statusCode, _, packagesFileContent := downloadTextFile(repository+"/src/contrib/PACKAGES", map[string]string{})
		if statusCode == 200 {
			inputPackagesFiles[repository] = packagesFileContent
		} else {
			log.Warn("An error occurred while downloading ", repository+"/src/contrib/PACKAGES")
		}
	}
	return inputPackagesFiles
}
