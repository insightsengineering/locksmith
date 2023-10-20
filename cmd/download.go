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
)

type DescriptionFile struct {
	Contents   string `json:"contents"`
	Repository string `json:"repository"`
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

func downloadDescriptionFiles(packageDescriptionList []string) []DescriptionFile {
	var inputDescriptionFiles []DescriptionFile
	for _, packageDescriptionURL := range packageDescriptionList {
		token := make(map[string]string)
		if strings.HasPrefix(packageDescriptionURL, "https://raw.githubusercontent.com") {
			token["Authorization"] = "token " + gitHubToken
		} else {
			token["Private-Token"] = gitLabToken
		}
		log.Info("Downloading ", packageDescriptionURL)
		statusCode, _, descriptionContent := downloadTextFile(packageDescriptionURL, token)
		if statusCode == 200 {
			inputDescriptionFiles = append(inputDescriptionFiles, DescriptionFile{descriptionContent, "GitHub"})
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
