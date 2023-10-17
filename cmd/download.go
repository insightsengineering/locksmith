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
	"net/http"
	"io"
	"strings"
)

// Returns HTTP status code for downloaded file, number of bytes in downloaded content,
// and the downloaded content itself.
func downloadFile(url string, parameters map[string]string) (int, int64, string) {
	tr := &http.Transport{ // #nosec
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec
	} // #nosec
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	checkError(err)
	for k, v := range parameters {
		req.Header.Add(k, v)
	}

	log.Debug("HTTP request = ", req)

	resp, err := client.Do(req)
	checkError(err)

	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			body, err2 := io.ReadAll(resp.Body)
			checkError(err2)
			log.Debug("HTTP response = ", string(body))
			return resp.StatusCode, resp.ContentLength, string(body)
		}
	}
	return -1, 0, ""
}

func downloadDescriptionFiles(repositoryList []string) {
	for _, repository := range repositoryList {
		if strings.HasPrefix(repository, "https://github.com") {
			repository = strings.ReplaceAll(repository, "https://github.com", "https://raw.githubusercontent.com")
			downloadFile(repository + "/main/DESCRIPTION", map[string]string{"Authorization": "token " + gitHubToken})
		}
		// TODO else if gitlab
	}

}
