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
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockedDownloadTextFile(url string, _ map[string]string) (int, int64, string) { // nolint: gocyclo
	switch {
	case url == "https://gitlab.example.com/api/v4/projects/37706/repository/tags/v1.3.1":
		return 200, 0, `{
			"commit": {
				"id": "aaabbbcccddd111"
			}
		}`
	case url == "https://gitlab.example.com/api/v4/projects/37706":
		return 200, 0, `{
			"path_with_namespace": "group1/group2/project1"
		}`
	case url == "https://gitlab.example.com/api/v4/projects/38706/repository/tags/v1.4.2":
		return 200, 0, `{
			"commit": {
				"id": "aaa222ccc444111"
			}
		}`
	case url == "https://gitlab.example.com/api/v4/projects/38706":
		return 200, 0, `{
			"path_with_namespace": "group3/group4/group5/project4"
		}`
	case url == "https://gitlab.example.com/api/v4/projects/30176/repository/tags/v0.2.0":
		return 200, 0, `{
			"commit": {
				"id": "fff222ccc444eee"
			}
		}`
	case url == "https://gitlab.example.com/api/v4/projects/30176":
		return 200, 0, `{
			"path_with_namespace": "group6/project7"
		}`
	case url == "https://gitlab.example.com/api/v4/projects/39307/repository/branches/main":
		return 200, 0, `{
			"commit": {
				"id": "fff555ddd888eee"
			}
		}`
	case url == "https://gitlab.example.com/api/v4/projects/39307":
		return 200, 0, `{
			"path_with_namespace": "group7/project8"
		}`
	case url == "https://gitlab.example.com/api/v4/projects/39211/repository/branches/main":
		return 200, 0, `{
			"commit": {
				"id": "fffeee999888aaa"
			}
		}`
	case url == "https://gitlab.example.com/api/v4/projects/39211":
		return 200, 0, `{
			"path_with_namespace": "group9/subgroup10/subgroup11/project9"
		}`
	case url == "https://api.github.com/repos/insightsengineering/formatters/git/ref/tags/v0.5.4":
		return 200, 0, `{
			"object": {
				"sha": "444eee222111eee"
			}
		}`
	case url == "https://api.github.com/repos/insightsengineering/rtables/git/ref/tags/v0.6.5":
		return 200, 0, `{
			"object": {
				"sha": "555ddd222111ddd"
			}
		}`
	case url == "https://api.github.com/repos/insightsengineering/nestcolor/git/ref/heads/main":
		return 200, 0, `{
			"object": {
				"sha": "555333aaabbbddd"
			}
		}`
	case url == "https://api.github.com/repos/insightsengineering/tern/git/ref/heads/main":
		return 200, 0, `{
			"object": {
				"sha": "555333aaaeeefff"
			}
		}`
	case url == "https://api.github.com/repos/insightsengineering/rlistings/git/ref/tags/v0.2.6":
		return 200, 0, `{
			"object": {
				"sha": "111444999eee222"
			}
		}`
	case url == "https://gitlab.example.com/api/v4/projects/37706/repository/files/subdirectory%2FDESCRIPTION/raw?ref=v1.3.1":
		return 200, 0, "DESCRIPTION contents 1"
	case url == "https://gitlab.example.com/api/v4/projects/38706/repository/files/subdirectory1%2Fsubdirectory2%2FDESCRIPTION/raw?ref=v1.4.2":
		return 200, 0, "DESCRIPTION contents 2"
	case url == "https://gitlab.example.com/api/v4/projects/30176/repository/files/DESCRIPTION/raw?ref=v0.2.0":
		return 200, 0, "DESCRIPTION contents 3"
	case url == "https://gitlab.example.com/api/v4/projects/39307/repository/files/DESCRIPTION/raw?ref=main":
		return 200, 0, "DESCRIPTION contents 4"
	case url == "https://gitlab.example.com/api/v4/projects/39211/repository/files/subdirectory1%2FDESCRIPTION/raw?ref=main":
		return 200, 0, "DESCRIPTION contents 5"
	case url == "https://raw.githubusercontent.com/insightsengineering/formatters/v0.5.4/subdirectory/DESCRIPTION":
		return 200, 0, "DESCRIPTION contents 6"
	case url == "https://raw.githubusercontent.com/insightsengineering/rtables/v0.6.5/subdirectory1/subdirectory2/DESCRIPTION":
		return 200, 0, "DESCRIPTION contents 7"
	case url == "https://raw.githubusercontent.com/insightsengineering/nestcolor/main/subdirectory/DESCRIPTION":
		return 200, 0, "DESCRIPTION contents 8"
	case url == "https://raw.githubusercontent.com/insightsengineering/tern/main/DESCRIPTION":
		return 200, 0, "DESCRIPTION contents 9"
	case url == "https://raw.githubusercontent.com/insightsengineering/rlistings/v0.2.6/DESCRIPTION":
		return 200, 0, "DESCRIPTION contents 10"
	case url == "https://repo1.example.com/repo1/src/contrib/PACKAGES":
		return 200, 0, "PACKAGES contents 1"
	case url == "https://repo2.example.com/repo2/src/contrib/PACKAGES":
		return 200, 0, "PACKAGES contents 2"
	case url == "https://repo3.example.com/repo3/src/contrib/PACKAGES":
		return 200, 0, "PACKAGES contents 3"
	}
	return 200, 0, ""
}

func Test_downloadDescriptionFiles(t *testing.T) {
	gitHubToken = "contentIrrelevantForTheTest"
	gitLabToken = "contentIrrelevantForTheTest"
	descriptionFileList := downloadDescriptionFiles([]string{
		"https://gitlab.example.com/api/v4/projects/37706/repository/files/subdirectory%2FDESCRIPTION/raw?ref=v1.3.1",
		"https://gitlab.example.com/api/v4/projects/38706/repository/files/subdirectory1%2Fsubdirectory2%2FDESCRIPTION/raw?ref=v1.4.2",
		"https://gitlab.example.com/api/v4/projects/30176/repository/files/DESCRIPTION/raw?ref=v0.2.0",
		"https://gitlab.example.com/api/v4/projects/39307/repository/files/DESCRIPTION/raw?ref=main",
		"https://gitlab.example.com/api/v4/projects/39211/repository/files/subdirectory1%2FDESCRIPTION/raw?ref=main",
		"https://raw.githubusercontent.com/insightsengineering/formatters/v0.5.4/subdirectory/DESCRIPTION",
		"https://raw.githubusercontent.com/insightsengineering/rtables/v0.6.5/subdirectory1/subdirectory2/DESCRIPTION",
		"https://raw.githubusercontent.com/insightsengineering/nestcolor/main/subdirectory/DESCRIPTION",
		"https://raw.githubusercontent.com/insightsengineering/tern/main/DESCRIPTION",
		"https://raw.githubusercontent.com/insightsengineering/rlistings/v0.2.6/DESCRIPTION",
	}, mockedDownloadTextFile)
	assert.Equal(t, descriptionFileList, []DescriptionFile{
		{
			"DESCRIPTION contents 1",
			"GitLab",
			"gitlab",
			"https://gitlab.example.com",
			"group1/group2",
			"project1",
			"subdirectory",
			"v1.3.1",
			"aaabbbcccddd111",
		},
		{
			"DESCRIPTION contents 2",
			"GitLab",
			"gitlab",
			"https://gitlab.example.com",
			"group3/group4/group5",
			"project4",
			"subdirectory1/subdirectory2",
			"v1.4.2",
			"aaa222ccc444111",
		},
		{
			"DESCRIPTION contents 3",
			"GitLab",
			"gitlab",
			"https://gitlab.example.com",
			"group6",
			"project7",
			"",
			"v0.2.0",
			"fff222ccc444eee",
		},
		{
			"DESCRIPTION contents 4",
			"GitLab",
			"gitlab",
			"https://gitlab.example.com",
			"group7",
			"project8",
			"",
			"main",
			"fff555ddd888eee",
		},
		{
			"DESCRIPTION contents 5",
			"GitLab",
			"gitlab",
			"https://gitlab.example.com",
			"group9/subgroup10/subgroup11",
			"project9",
			"subdirectory1",
			"main",
			"fffeee999888aaa",
		},
		{
			"DESCRIPTION contents 6",
			"GitHub",
			"github",
			"api.github.com",
			"insightsengineering",
			"formatters",
			"subdirectory",
			"v0.5.4",
			"444eee222111eee",
		},
		{
			"DESCRIPTION contents 7",
			"GitHub",
			"github",
			"api.github.com",
			"insightsengineering",
			"rtables",
			"subdirectory1/subdirectory2",
			"v0.6.5",
			"555ddd222111ddd",
		},
		{
			"DESCRIPTION contents 8",
			"GitHub",
			"github",
			"api.github.com",
			"insightsengineering",
			"nestcolor",
			"subdirectory",
			"main",
			"555333aaabbbddd",
		},
		{
			"DESCRIPTION contents 9",
			"GitHub",
			"github",
			"api.github.com",
			"insightsengineering",
			"tern",
			"",
			"main",
			"555333aaaeeefff",
		},
		{
			"DESCRIPTION contents 10",
			"GitHub",
			"github",
			"api.github.com",
			"insightsengineering",
			"rlistings",
			"",
			"v0.2.6",
			"111444999eee222",
		},
	})
}

func Test_downloadPackagesFiles(t *testing.T) {
	packagesFiles := downloadPackagesFiles([]string{
		"https://repo1.example.com/repo1",
		"https://repo2.example.com/repo2",
		"https://repo3.example.com/repo3",
	}, mockedDownloadTextFile)
	assert.Equal(t, packagesFiles, map[string]string{
		"https://repo1.example.com/repo1": "PACKAGES contents 1",
		"https://repo2.example.com/repo2": "PACKAGES contents 2",
		"https://repo3.example.com/repo3": "PACKAGES contents 3",
	})
}
