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

type DescriptionFile struct {
	Contents string `json:"contents"`
	// GitHub or GitLab
	PackageSource  string `json:"source"`
	RemoteType     string `json:"remoteType"`
	RemoteHost     string `json:"remoteHost"`
	RemoteUsername string `json:"remoteUsername"`
	RemoteRepo     string `json:"remoteRepo"`
	RemoteSubdir   string `json:"remoteSubdir"`
	RemoteRef      string `json:"remoteRef"`
	RemoteSha      string `json:"remoteSha"`
}

type PackagesFile struct {
	Packages []PackageDescription `json:"packages"`
}

type PackageDescription struct {
	Package        string       `json:"Package"`
	Version        string       `json:"Version"`
	Source         string       `json:"Source"`
	Repository     string       `json:"Repository"`
	Dependencies   []Dependency `json:"Requirements"`
	RemoteType     string       `json:"RemoteType"`
	RemoteHost     string       `json:"RemoteHost"`
	RemoteUsername string       `json:"RemoteUsername"`
	RemoteRepo     string       `json:"RemoteRepo"`
	RemoteSubdir   string       `json:"RemoteSubdir"`
	RemoteRef      string       `json:"RemoteRef"`
	RemoteSha      string       `json:"RemoteSha"`
}

type Dependency struct {
	DependencyType  string `json:"type"`
	DependencyName  string `json:"name"`
	VersionOperator string `json:"operator"`
	VersionValue    string `json:"value"`
}
