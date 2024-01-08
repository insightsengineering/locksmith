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

func Test_GenerateRenvLock(t *testing.T) {
	renvLock := GenerateRenvLock([]PackageDescription{
		{
			"package1",
			"1.0.2",
			"GitHub",
			"",
			[]Dependency{},
			"github",
			"api.github.com",
			"group1/group2",
			"package1",
			"subdirectory1",
			"main",
			"aaabbb444333",
			[]string{},
		},
		{
			"package2",
			"2.5.4.3",
			"GitLab",
			"",
			[]Dependency{},
			"gitlab",
			"https://gitlab.example.com",
			"group3/group4",
			"package2",
			"subdirectory2",
			"v2.5.4.3",
			"eee888222aaa",
			[]string{},
		},
		{
			"", "", "", "", []Dependency{}, "", "", "", "", "", "", "", []string{},
		},
		{
			"package3",
			"3.2.7.8",
			"Repository",
			"https://repo1.example.com/repo1",
			[]Dependency{},
			"", "", "", "", "", "", "", []string{},
		},
		{
			"package4",
			"4.1.2",
			"Repository",
			"https://repo2.example.com/repo2",
			[]Dependency{},
			"", "", "", "", "", "", "", []string{},
		},
		{
			"package5",
			"0.0.5",
			"Repository",
			"https://repo3.example.com/repo3",
			[]Dependency{},
			"", "", "", "", "", "", "", []string{},
		},
	}, map[string]string{
		"Repo1": "https://repo1.example.com/repo1",
		"Repo2": "https://repo2.example.com/repo2",
		"Repo3": "https://repo3.example.com/repo3",
	})
	assert.Equal(t, renvLock, RenvLock{
		RenvLockContents{
			[]RenvLockRepository{
				{"Repo1", "https://repo1.example.com/repo1"},
				{"Repo2", "https://repo2.example.com/repo2"},
				{"Repo3", "https://repo3.example.com/repo3"},
			},
		},
		map[string]PackageDescription{
			"package1": {
				"package1",
				"1.0.2",
				"GitHub",
				"",
				[]Dependency{},
				"github",
				"api.github.com",
				"group1/group2",
				"package1",
				"subdirectory1",
				"main",
				"aaabbb444333",
				[]string{},
			},
			"package2": {
				"package2",
				"2.5.4.3",
				"GitLab",
				"",
				[]Dependency{},
				"gitlab",
				"https://gitlab.example.com",
				"group3/group4",
				"package2",
				"subdirectory2",
				"v2.5.4.3",
				"eee888222aaa",
				[]string{},
			},
			"package3": {
				"package3",
				"3.2.7.8",
				"Repository",
				"Repo1",
				[]Dependency{},
				"", "", "", "", "", "", "", []string{},
			},
			"package4": {
				"package4",
				"4.1.2",
				"Repository",
				"Repo2",
				[]Dependency{},
				"", "", "", "", "", "", "", []string{},
			},
			"package5": {
				"package5",
				"0.0.5",
				"Repository",
				"Repo3",
				[]Dependency{},
				"", "", "", "", "", "", "", []string{},
			},
		},
	})
}

func Test_GetPackageRegex(t *testing.T) {
	packageRegex := GetPackageRegex("package*,*some.Package,test1,my*awesome*package")
	assert.Equal(t, packageRegex, `^package.*$|^.*some\.Package$|^test1$|^my.*awesome.*package$`)
}

func Test_GetPackageVersionFromDescription(t *testing.T) {
	version1 := GetPackageVersionFromDescription("testdata/DESCRIPTION1")
	assert.Equal(t, version1, "0.14.0.9012")
	version2 := GetPackageVersionFromDescription("testdata/DESCRIPTION2")
	assert.Equal(t, version2, "0.9.1.9013")
	version3 := GetPackageVersionFromDescription("testdata/NON_EXISTENT_DESCRIPTION")
	assert.Equal(t, version3, "")
}

func Test_GetGitRepositoryURL(t *testing.T) {
	repoURL1 := GetGitRepositoryURL(PackageDescription{
		"", "", "GitHub", "", []Dependency{}, "",
		"api.github.com", "github-org-1", "repo-name-1", "", "", "", []string{},
	})
	assert.Equal(t, repoURL1, "https://github.com/github-org-1/repo-name-1")
	repoURL2 := GetGitRepositoryURL(PackageDescription{
		"", "", "GitLab", "", []Dependency{}, "",
		"https://gitlab.example.com", "org1/org2", "repo-name-2", "", "", "", []string{},
	})
	assert.Equal(t, repoURL2, "https://gitlab.example.com/org1/org2/repo-name-2")
	repoURL3 := GetGitRepositoryURL(PackageDescription{
		"", "", "GitLab", "", []Dependency{}, "",
		"gitlab.example.com", "org3/org4", "repo-name-3", "", "", "", []string{},
	})
	assert.Equal(t, repoURL3, "https://gitlab.example.com/org3/org4/repo-name-3")
}

func mockedGetDefaultBranchSha(_ string, repoURL string, _ string) string {
	switch {
	case repoURL == "https://github.com/group1/group2/package11":
		return "eee111555bbbccc"
	case repoURL == "https://gitlab.example.com/group3/group4/package12":
		return "888444dddbbbaaa"
	}
	return ""
}

func Test_UpdateGitPackages(t *testing.T) {
	renvLock := RenvLock{
		RenvLockContents{
			[]RenvLockRepository{
				{"Repo1", "https://repo1.example.com/repo1"},
			},
		},
		map[string]PackageDescription{
			"package11": {
				"package11",
				"1.0.2",
				"GitHub",
				"",
				[]Dependency{},
				"github",
				"api.github.com",
				"group1/group2",
				"package11",
				"subdirectory1",
				"main",
				"aaabbb444333",
				[]string{},
			},
			"package12": {
				"package12",
				"2.5.4.3",
				"GitLab",
				"",
				[]Dependency{},
				"gitlab",
				"https://gitlab.example.com",
				"group3/group4",
				"package12",
				"subdirectory2",
				"v2.5.4.3",
				"eee888222aaa",
				[]string{},
			},
			"package3": {
				"package3",
				"3.2.7.8",
				"Repository",
				"Repo1",
				[]Dependency{},
				"", "", "", "", "", "", "", []string{},
			},
			"package4": {
				"package4",
				"3.7.0",
				"GitLab",
				"",
				[]Dependency{},
				"gitlab",
				"https://gitlab.example.com",
				"group6/group7",
				"package4",
				"",
				"v3.7.0",
				"ccceee444999",
				[]string{},
			},
		},
	}
	UpdateGitPackages(&renvLock, "package1*", mockedGetDefaultBranchSha, "testdata/git_updates/")
	assert.Equal(t, renvLock.Packages["package11"].Version, "1.0.4")
	assert.Equal(t, renvLock.Packages["package12"].Version, "2.6.1.1")
	assert.Equal(t, renvLock.Packages["package4"].Version, "3.7.0")
	assert.Equal(t, renvLock.Packages["package11"].RemoteSha, "eee111555bbbccc")
	assert.Equal(t, renvLock.Packages["package12"].RemoteSha, "888444dddbbbaaa")
	assert.Equal(t, renvLock.Packages["package4"].RemoteSha, "ccceee444999")
}
