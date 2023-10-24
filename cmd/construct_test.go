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

func Test_checkIfVersionSufficient(t *testing.T) {
	assert.True(t, checkIfVersionSufficient("2", ">=", "1"))
	assert.True(t, checkIfVersionSufficient("2", ">", "1"))
	assert.False(t, checkIfVersionSufficient("1", ">=", "2"))
	assert.False(t, checkIfVersionSufficient("1", ">", "2"))
	assert.False(t, checkIfVersionSufficient("2", ">", "2"))
	assert.True(t, checkIfVersionSufficient("2", ">=", "2"))
	assert.True(t, checkIfVersionSufficient("1.2", ">=", "1.2"))
	assert.False(t, checkIfVersionSufficient("1.2", ">", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.3", ">=", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.3", ">", "1.2"))
	assert.False(t, checkIfVersionSufficient("1.2", ">=", "1.3"))
	assert.False(t, checkIfVersionSufficient("1.2", ">", "1.3"))
	assert.False(t, checkIfVersionSufficient("1", ">=", "1.2"))
	assert.False(t, checkIfVersionSufficient("1", ">", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.2", ">=", "1"))
	assert.True(t, checkIfVersionSufficient("1.2", ">", "1"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">=", "1.2.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">", "1.2.4"))
	assert.True(t, checkIfVersionSufficient("1.2.3", ">=", "1.2.3"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.2.4", ">=", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.2.4", ">", "1.2.3"))
	assert.False(t, checkIfVersionSufficient("1.2", ">=", "1.2.3"))
	assert.False(t, checkIfVersionSufficient("1.2", ">", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.2.3", ">=", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.2.3", ">", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.3", ">=", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.3", ">", "1.2.3"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">=", "1.3"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">", "1.3"))
	assert.False(t, checkIfVersionSufficient("1", ">=", "1.2.3"))
	assert.False(t, checkIfVersionSufficient("1", ">", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.2.3", ">=", "1"))
	assert.True(t, checkIfVersionSufficient("1.2.3", ">", "1"))
	assert.True(t, checkIfVersionSufficient("2", ">=", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("2", ">", "1.2.3"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">=", "2"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">", "2"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">=", "1.2.3.5"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">", "1.2.3.5"))
	assert.True(t, checkIfVersionSufficient("1.2.3.5", ">=", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.2.3.5", ">", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">=", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">=", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">=", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.2.4", ">=", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.2.4", ">", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">=", "1.2.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">", "1.2.4"))
	assert.False(t, checkIfVersionSufficient("1.2", ">=", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1.2", ">", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">=", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.3", ">=", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.3", ">", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">=", "1.3"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">", "1.3"))
	assert.False(t, checkIfVersionSufficient("1", ">=", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1", ">", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("2", ">=", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("2", ">", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">=", "1"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">", "1"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">=", "2"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">", "2"))
}

func Test_constructOutputPackageList(t *testing.T) {
	var repositoryList = []string{
		"https://repo1.example.com/ExampleRepo1",
		"https://repo2.example.com/ExampleRepo2",
		"https://repo3.example.com/ExampleRepo3",
	}
	packagesFiles := make(map[string]PackagesFile)
	packagesFiles["https://repo1.example.com/ExampleRepo1"] = PackagesFile{
		[]PackageDescription{
			{
				"package3",
				"1.2.0",
				"", "",
				[]Dependency{
					{
						"Depends",
						"package11",
						">=",
						"0.7",
					},
					{
						"Imports",
						"package12",
						"",
						"",
					},
					{
						"Suggests",
						"package13",
						"",
						"",
					},
				},
				"", "", "", "", "", "", "",
			},
			{
				"package4",
				"0.7.5",
				"", "",
				[]Dependency{
					{
						"Imports",
						"package11",
						">=",
						"4.5",
					},
					{
						"Imports",
						"package14",
						"",
						"",
					},
				},
				"", "", "", "", "", "", "",
			},
			{
				"package11",
				"0.7.8",
				"", "",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package14",
				"2.5.8",
				"", "",
				[]Dependency{
					{
						"Depends",
						"package15",
						">=",
						"3.2",
					},
					{
						"Imports",
						"package16",
						">=",
						"2.2",
					},
				},
				"", "", "", "", "", "", "",
			},
			{
				"package16",
				"2.4.5",
				"", "",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package6",
				"3.0.1",
				"", "",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package10",
				"3.0.2",
				"", "",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
		},
	}
	packagesFiles["https://repo2.example.com/ExampleRepo2"] = PackagesFile{
		[]PackageDescription{
			{
				"package4",
				"1.1.1",
				"", "",
				[]Dependency{
					{
						"Imports",
						"package11",
						">=",
						"4.5",
					},
					{
						"Imports",
						"package14",
						"",
						"",
					},
					{
						"Imports",
						"package12",
						"",
						"",
					},
				},
				"", "", "", "", "", "", "",
			},
			{
				"package5",
				"3.2.0",
				"", "",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package7",
				"1.6.2",
				"", "",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package9",
				"2.4",
				"", "",
				[]Dependency{
					{
						"Imports",
						"R",
						">=",
						"3.6",
					},
				},
				"", "", "", "", "", "", "",
			},
			{
				"package11",
				"5.4.7",
				"", "",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package12",
				"1.2.3",
				"", "",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package15",
				"3.3.4.5",
				"", "",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
		},
	}
	packagesFiles["https://repo3.example.com/ExampleRepo3"] = PackagesFile{
		[]PackageDescription{
			{
				"package8",
				"1.9.2",
				"", "",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
		},
	}
	outputPackageList := constructOutputPackageList(
		[]PackageDescription{
			{
				"package1",
				"1.2.3",
				"GitHub",
				"",
				[]Dependency{
					{
						"Depends",
						"R",
						">=",
						"4.0",
					},
					{
						"Depends",
						"package3",
						"",
						"",
					},
					{
						"Imports",
						"package4",
						">=",
						"1.0",
					},
					{
						"Suggests",
						"package5",
						"",
						"",
					},
					{
						"LinkingTo",
						"package6",
						"",
						"",
					},
				},
				"", "", "", "", "", "", "",
			},
			{
				"package2",
				"2.3.4",
				"GitHub",
				"",
				[]Dependency{
					{
						"Depends",
						"R",
						">=",
						"3.6",
					},
					{
						"Depends",
						"package7",
						"",
						"",
					},
					{
						"Suggests",
						"package16",
						"",
						"",
					},
					{
						"Suggests",
						"package1",
						"",
						"",
					},
					{
						"Imports",
						"package8",
						">",
						"1.8.3",
					},
					{
						"Suggests",
						"package9",
						">=",
						"2.3",
					},
					{
						"LinkingTo",
						"package10",
						"",
						"",
					},
				},
				"", "", "", "", "", "", "",
			},
		},
		packagesFiles, repositoryList,
	)
	assert.Equal(t, outputPackageList,
		[]PackageDescription{
			{
				"package1",
				"1.2.3",
				"GitHub",
				"",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package2",
				"2.3.4",
				"GitHub",
				"",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package3",
				"1.2.0",
				"Repository",
				"https://repo1.example.com/ExampleRepo1",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				// package11 removed from here
				// First it was required by package3 in version >= 0.7
				// so a compatible version was found in repo1.
				// However afterwards, package4 requested package11 >= 4.5
				// so it had to be retrieved from repo2.
				// The reference to repo1 was overwritten here.
				"", "", "", "", []Dependency{}, "", "", "", "", "", "", "",
			},
			{
				"package12",
				"1.2.3",
				"Repository",
				"https://repo2.example.com/ExampleRepo2",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package4",
				"1.1.1",
				"Repository",
				"https://repo2.example.com/ExampleRepo2",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package11",
				"5.4.7",
				"Repository",
				"https://repo2.example.com/ExampleRepo2",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package14",
				"2.5.8",
				"Repository",
				"https://repo1.example.com/ExampleRepo1",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package15",
				"3.3.4.5",
				"Repository",
				"https://repo2.example.com/ExampleRepo2",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package16",
				"2.4.5",
				"Repository",
				"https://repo1.example.com/ExampleRepo1",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package5",
				"3.2.0",
				"Repository",
				"https://repo2.example.com/ExampleRepo2",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package6",
				"3.0.1",
				"Repository",
				"https://repo1.example.com/ExampleRepo1",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package7",
				"1.6.2",
				"Repository",
				"https://repo2.example.com/ExampleRepo2",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package8",
				"1.9.2",
				"Repository",
				"https://repo3.example.com/ExampleRepo3",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package9",
				"2.4",
				"Repository",
				"https://repo2.example.com/ExampleRepo2",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
			{
				"package10",
				"3.0.2",
				"Repository",
				"https://repo1.example.com/ExampleRepo1",
				[]Dependency{},
				"", "", "", "", "", "", "",
			},
		},
	)
}
