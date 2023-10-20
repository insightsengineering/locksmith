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
)

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

func processDescriptionURL(descriptionURL string) (map[string]string, string, string, string, string, string) {
	token := make(map[string]string)
	var repositoryType string
	var remoteRef string
	var remoteHost string
	var remoteUsername string
	var remoteRepo string
	if strings.HasPrefix(descriptionURL, "https://raw.githubusercontent.com") {
		// Expecting URL in form:
		// https://raw.githubusercontent.com/<organization>/<repo-name>/<ref-name>/DESCRIPTION
		token["Authorization"] = "token " + gitHubToken
		repositoryType = "GitHub"
		shorterURL := strings.TrimPrefix(descriptionURL, "https://raw.githubusercontent.com/")
		remoteHost = "api.github.com"
		remoteUsername = strings.Split(shorterURL, "/")[0]
		remoteRepo = strings.Split(shorterURL, "/")[1]
		remoteRef = strings.Split(shorterURL, "/")[2]
		// TODO get remoteSha based on remoteRef
	} else {
		// Expecting URL in form:
		// https://example.gitlab.com/api/v4/projects/<project-id>/repository/files/DESCRIPTION/raw?ref=<ref-name>
		re := regexp.MustCompile(`ref=.*$`)
		token["Private-Token"] = gitLabToken
		repositoryType = "GitLab"
		shorterURL := strings.TrimPrefix(descriptionURL, "https://")
		remoteHost = strings.Split(shorterURL, "/")[0]
		remoteRef = strings.TrimPrefix(re.FindString(descriptionURL), "ref=")
		// TODO get remoteSha based on remoteRef
		// TODO get remoteUsername and remoteRepo from GitLab API based on project ID
	}
	return token, repositoryType, remoteHost, remoteUsername, remoteRepo, remoteRef
}

func downloadDescriptionFiles(packageDescriptionList []string) []DescriptionFile {
	var inputDescriptionFiles []DescriptionFile
	for _, packageDescriptionURL := range packageDescriptionList {
		token, repositoryType, _, remoteUsername, remoteRepo, remoteRef :=
			processDescriptionURL(packageDescriptionURL)
		log.Info(
			"Downloading ", packageDescriptionURL, "\nremoteRef = ", remoteRef, ", ",
			"repositoryType = ", repositoryType, ", remoteUsername = ", remoteUsername,
			", remoteRepo = ", remoteRepo,
		)
		statusCode, _, descriptionContent := downloadTextFile(packageDescriptionURL, token)
		if statusCode == 200 {
			inputDescriptionFiles = append(
				inputDescriptionFiles,
				DescriptionFile{descriptionContent, repositoryType, remoteRef},
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
