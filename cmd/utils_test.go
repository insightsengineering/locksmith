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

func Test_ParseInput(t *testing.T) {
	inputPackageList = "https://raw.githubusercontent.com/insightsengineering/tern/main/DESCRIPTION,https://raw.githubusercontent.com/insightsengineering/rlistings/v0.2.6/DESCRIPTION"
	inputRepositoryList = "Repo1=https://repo1.example.com/repo1,Repo2=https://repo2.example.com/repo2,Repo3=https://repo3.example.com/repo3"
	packageList, repositoryList, repositoryMap := ParseInput()
	assert.Equal(t, packageList, []string{
		"https://raw.githubusercontent.com/insightsengineering/tern/main/DESCRIPTION",
		"https://raw.githubusercontent.com/insightsengineering/rlistings/v0.2.6/DESCRIPTION",
	})
	assert.Equal(t, repositoryList, []string{
		"https://repo1.example.com/repo1",
		"https://repo2.example.com/repo2",
		"https://repo3.example.com/repo3",
	})
	assert.Equal(t, repositoryMap, map[string]string{
		"Repo1": "https://repo1.example.com/repo1",
		"Repo2": "https://repo2.example.com/repo2",
		"Repo3": "https://repo3.example.com/repo3",
	})
}
