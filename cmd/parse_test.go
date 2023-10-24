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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_processPackagesFile(t *testing.T) {
	byteValue, err := os.ReadFile("testdata/PACKAGES")
	checkError(err)
	allPackages := processPackagesFile(string(byteValue))
	assert.Equal(t, allPackages,
		PackagesFile{
			[]PackageDescription{
				{
					"somePackage1",
					"1.0.0",
					"", "",
					[]Dependency{
						{
							"Depends",
							"R",
							">=",
							"2.15.0",
						},
					},
					"", "", "", "", "", "", "",
				},
				{
					"somePackage2",
					"2.0.0",
					"", "",
					[]Dependency{
						{
							"Depends",
							"R",
							">=",
							"3.6.0",
						},
						{
							"Imports",
							"magrittr",
							"",
							"",
						},
						{
							"Imports",
							"dplyr",
							"",
							"",
						},
					},
					"", "", "", "", "", "", "",
				},
				{
					"somePackage3",
					"0.0.1",
					"", "",
					[]Dependency{
						{
							"Depends",
							"R",
							">=",
							"3.1.0",
						},
						{
							"Imports",
							"ggplot2",
							">=",
							"3.1.0",
						},
						{
							"Imports",
							"shiny",
							">=",
							"1.3.1",
						},
						{
							"Suggests",
							"rmarkdown",
							">=",
							"1.13",
						},
						{
							"Suggests",
							"knitr",
							">=",
							"1.22",
						},
					},
					"", "", "", "", "", "", "",
				},
				{
					"somePackage4",
					"0.2",
					"", "",
					[]Dependency{
						{
							"Suggests",
							"testthat",
							">=",
							"3.0.0",
						},
						{
							"Suggests",
							"ggplot2",
							">=",
							"3.4.0",
						},
						{
							"Suggests",
							"knitr",
							">=",
							"1.30",
						},
						{
							"Suggests",
							"mockery",
							">=",
							"0.4.2",
						},
						{
							"Suggests",
							"rmarkdown",
							">=",
							"2.6",
						},
						{
							"Suggests",
							"roxygen2",
							">=",
							"7.1.0",
						},
					},
					"", "", "", "", "", "", "",
				},
			},
		},
	)
}

func Test_parseDescriptionFileList(t *testing.T) {
	byteValue1, err := os.ReadFile("testdata/DESCRIPTION1")
	checkError(err)
	byteValue2, err := os.ReadFile("testdata/DESCRIPTION2")
	checkError(err)
	descriptionFileList := []DescriptionFile{
		{string(byteValue1), "GitHub", "", "", "", "", "", "", ""},
		{string(byteValue2), "GitHub", "", "", "", "", "", "", ""},
	}
	allPackages := parseDescriptionFileList(descriptionFileList)
	assert.Equal(t, allPackages,
		[]PackageDescription{
			{
				"my.awesome.package",
				"0.14.0.9012",
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
						"shiny",
						">=",
						"1.7.0",
					},
					{
						"Imports",
						"checkmate",
						"",
						"",
					},
					{
						"Imports",
						"lifecycle",
						"",
						"",
					},
					{
						"Imports",
						"logger",
						">=",
						"0.2.0",
					},
					{
						"Imports",
						"magrittr",
						"",
						"",
					},
					{
						"Imports",
						"rlang",
						"",
						"",
					},
					{
						"Imports",
						"shinyjs",
						"",
						"",
					},
					{
						"Imports",
						"rmarkdown",
						">=",
						"0.1.1",
					},
					{
						"Imports",
						"MultiAssayExperiment",
						">=",
						"0.2.0",
					},
					{
						"Imports",
						"yaml",
						">=",
						"0.4.0",
					},
					{
						"Imports",
						"utils",
						"",
						"",
					},
					{
						"Suggests",
						"covr",
						"",
						"",
					},
					{
						"Suggests",
						"dplyr",
						"",
						"",
					},
					{
						"Suggests",
						"knitr",
						"",
						"",
					},
					{
						"Suggests",
						"testthat",
						">=",
						"3.1.5",
					},
					{
						"Suggests",
						"withr",
						"",
						"",
					},
				},
				"", "", "", "", "", "", "",
			},
			{
				"my.awesome.package.2",
				"0.9.1.9013",
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
						"rtables",
						">",
						"0.6.4",
					},
					{
						"Imports",
						"dplyr",
						"",
						"",
					},
					{
						"Imports",
						"forcats",
						">=",
						"1.0.0",
					},
					{
						"Imports",
						"formatters",
						">=",
						"0.5.3",
					},
					{
						"Imports",
						"ggplot2",
						">=",
						"3.4.0",
					},
					{
						"Imports",
						"stats",
						"",
						"",
					},
					{
						"Imports",
						"survival",
						">=",
						"3.2-13",
					},
					{
						"Imports",
						"tibble",
						"",
						"",
					},
					{
						"Imports",
						"tidyr",
						"",
						"",
					},
					{
						"Imports",
						"utils",
						"",
						"",
					},
					{
						"Suggests",
						"knitr",
						"",
						"",
					},
					{
						"Suggests",
						"lattice",
						"",
						"",
					},
					{
						"Suggests",
						"lubridate",
						"",
						"",
					},
					{
						"Suggests",
						"rmarkdown",
						"",
						"",
					},
					{
						"Suggests",
						"stringr",
						"",
						"",
					},
					{
						"Suggests",
						"testthat",
						">=",
						"3.0",
					},
					{
						"Suggests",
						"vdiffr",
						">=",
						"1.0.0",
					},
				},
				"", "", "", "", "", "", "",
			},
		},
	)
}
