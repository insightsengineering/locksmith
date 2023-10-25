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
		},
		{
			"",
			"",
			"",
			"",
			[]Dependency{},
			"",
			"",
			"",
			"",
			"",
			"",
			"",
		},
		{
			"package3",
			"3.2.7.8",
			"Repository",
			"https://repo1.example.com/repo1",
			[]Dependency{},
			"",
			"",
			"",
			"",
			"",
			"",
			"",
		},
		{
			"package4",
			"4.1.2",
			"Repository",
			"https://repo2.example.com/repo2",
			[]Dependency{},
			"",
			"",
			"",
			"",
			"",
			"",
			"",
		},
		{
			"package5",
			"0.0.5",
			"Repository",
			"https://repo3.example.com/repo3",
			[]Dependency{},
			"",
			"",
			"",
			"",
			"",
			"",
			"",
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
			},
			"package3": {
				"package3",
				"3.2.7.8",
				"Repository",
				"Repo1",
				[]Dependency{},
				"",
				"",
				"",
				"",
				"",
				"",
				"",
			},
			"package4": {
				"package4",
				"4.1.2",
				"Repository",
				"Repo2",
				[]Dependency{},
				"",
				"",
				"",
				"",
				"",
				"",
				"",
			},
			"package5": {
				"package5",
				"0.0.5",
				"Repository",
				"Repo3",
				[]Dependency{},
				"",
				"",
				"",
				"",
				"",
				"",
				"",
			},
		},
	})
}
