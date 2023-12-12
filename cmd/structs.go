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

// DescriptionFile represents the input package DESCRIPTION file together with related
// information about the git repository where the package is stored.
// This structure represents data about input git packages before
// it is parsed into a PackageDescription struct.
type DescriptionFile struct {
	// Contents stores the DESCRIPTION file.
	Contents string `json:"contents"`
	// PackageSource can be either 'GitHub' or 'GitLab'.
	PackageSource string `json:"source"`
	// RemoteType can be either 'github' or 'gitlab'.
	RemoteType string `json:"remoteType"`
	// RemoteHost can be 'api.github.com' or the URL of GitLab instance,
	// for example: 'https://gitlab.example.com'.
	RemoteHost string `json:"remoteHost"`
	// RemoteUsername represents the organization or the owner in case of a GitHub
	// repository, or the path to the repository in the project tree in case of
	// a GitLab repository.
	RemoteUsername string `json:"remoteUsername"`
	// RemoteRepo contains the name of git repository.
	RemoteRepo string `json:"remoteRepo"`
	// RemoteSubdir is an optional field storing the path to the package inside
	// the git repository in case the package is not located in the root of the
	// git repository.
	RemoteSubdir string `json:"remoteSubdir"`
	// RemoteRef is tag or branch name representing the verion of the provided
	// package DESCRIPTION file. If RemoteRef matches `v\d+(\.\d+)*` regex,
	// it is treated as a git tag, otherwise it is treated as a git branch.
	RemoteRef string `json:"remoteRef"`
	// RemoteSha is the commit SHA for the RemoteRef.
	RemoteSha string `json:"remoteSha"`
}

type PackagesFile struct {
	Packages []PackageDescription `json:"packages"`
}

type RenvLock struct {
	R        RenvLockContents              `json:"R"`
	Packages map[string]PackageDescription `json:"Packages"`
}

type RenvLockRepository struct {
	Name string `json:"Name"`
	URL  string `json:"URL"`
}

type RenvLockContents struct {
	Repositories []RenvLockRepository `json:"Repositories"`
}

// PackageDescrition represents an R package.
type PackageDescription struct {
	// Package stores the package name.
	Package string `json:"Package"`
	// Version stores the package version.
	Version string `json:"Version"`
	// Source can be one of: 'GitHub', 'GitLab' (for packages from git repositories)
	// or 'Repository' (for packages from package repositories).
	Source string `json:"Source"`
	// Repository stores the URL or the name (depending on the stage of processing)
	// of the package repository, in case Source is 'Repository'.
	Repository string `json:"Repository,omitempty"`
	// Dependencies contains the list of package dependencies.
	Dependencies []Dependency `json:"Requirements,omitempty"`
	// When processing packages stored in package repositories, the fields below are empty.
	// These fields are documented in the DescriptionFile struct.
	RemoteType     string `json:"RemoteType,omitempty"`
	RemoteHost     string `json:"RemoteHost,omitempty"`
	RemoteUsername string `json:"RemoteUsername,omitempty"`
	RemoteRepo     string `json:"RemoteRepo,omitempty"`
	RemoteSubdir   string `json:"RemoteSubdir,omitempty"`
	RemoteRef      string `json:"RemoteRef,omitempty"`
	RemoteSha      string `json:"RemoteSha,omitempty"`
}

type Dependency struct {
	DependencyType  string `json:"type"`
	DependencyName  string `json:"name"`
	VersionOperator string `json:"operator"`
	VersionValue    string `json:"value"`
}

type DependencyVersion struct {
	VersionOperator string `json:"operator"`
	VersionValue    string `json:"value"`
}
