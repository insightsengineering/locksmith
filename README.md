# locksmith

[![build](https://github.com/insightsengineering/locksmith/actions/workflows/test.yml/badge.svg)](https://github.com/insightsengineering/locksmith/actions/workflows/test.yml)

`locksmith` is a utility to generate `renv.lock` file containing all dependencies of given set of R packages.

Given the input list of git repositories containing the R packages, as well as a list of R package
repositories (e.g. in a package manager, CRAN, BioConductor etc.), `locksmith` will try to determine
the list of all dependencies and their versions required to make the input list of packages work.
It will then save the result in an `renv.lock`-compatible file.

For additional information about `renv.lock`, please refer to the [`renv` documentation](https://rstudio.github.io/renv/articles/renv.html).

## Installation

Simply download the project for your distribution from the
[releases](https://github.com/insightsengineering/locksmith/releases) page. `locksmith` is
distributed as a single binary file and does not need any additional system requirements.

Alternatively, you can install the latest version by running:

```shell
go install github.com/insightsengineering/locksmith@latest
```

## Usage

`locksmith` is a command line utility, so after installing the binary in your `PATH`, simply run the
following command to view its capabilities:

```bash
locksmith --help
```

Real-life example with multiple input packages and repositories.
Please see below for [an example](#configuration-file) how to set package and repository lists more
easily in a configuration file.

```bash
locksmith --inputPackageList https://raw.githubusercontent.com/insightsengineering/formatters/main/DESCRIPTION,https://raw.githubusercontent.com/insightsengineering/rtables/main/DESCRIPTION,https://raw.githubusercontent.com/insightsengineering/scda/main/DESCRIPTION,https://raw.githubusercontent.com/insightsengineering/scda.2022/main/DESCRIPTION,https://raw.githubusercontent.com/insightsengineering/nestcolor/main/DESCRIPTION,https://raw.githubusercontent.com/insightsengineering/tern/main/DESCRIPTION,https://raw.githubusercontent.com/insightsengineering/rlistings/main/DESCRIPTION,https://gitlab.example.com/api/v4/projects/123456/repository/files/DESCRIPTION/raw?ref=main,https://gitlab.example.com/api/v4/projects/234567/repository/files/directory%2Fsubdirectory%2FDESCRIPTION/raw?ref=main --inputRepositoryList BioC=https://bioconductor.org/packages/release/bioc,CRAN=https://cran.rstudio.com
```

In order to download the input `DESCRIPTION` files from GitHub or GitLab repositories, please set the environment
variables containing the Personal Access Tokens.

* For GitHub, set the `LOCKSMITH_GITHUBTOKEN` environment variable.
* For GitLab, set the `LOCKSMITH_GITLABTOKEN` environment variable.

By default `locksmith` will save the resulting output file to `renv.lock`.

## Configuration file

If you'd like to set the above options in a configuration file, by default `locksmith` checks
`~/.locksmith`, `~/.locksmith.yaml` and `~/.locksmith.yml` files.

If any of these files exist, `locksmith` will use options defined there, unless they are overridden
by command line flags or environment variables.

You can also specify custom path to configuration file with `--config <your-configuration-file>.yml`
command line flag. When using custom configuration file, if you specify command line flags,
the latter will still take precedence.

Example contents of configuration file:

```yaml
logLevel: debug
inputPackages:
  - https://raw.githubusercontent.com/insightsengineering/formatters/main/DESCRIPTION
  - https://raw.githubusercontent.com/insightsengineering/rtables/main/DESCRIPTION
  - https://raw.githubusercontent.com/insightsengineering/scda/main/DESCRIPTION
  - https://raw.githubusercontent.com/insightsengineering/scda.2022/main/DESCRIPTION
  - https://gitlab.example.com/api/v4/projects/123456/repository/files/DESCRIPTION/raw?ref=main
  # Forward slashes in 'directory/subdirectory/DESCRIPTION' path are replaced by '%2F' due to URL encoding
  - https://gitlab.example.com/api/v4/projects/234567/repository/files/directory%2Fsubdirectory%2FDESCRIPTION/raw?ref=main
inputRepositories:
  - Bioconductor.BioCsoft=https://bioconductor.org/packages/release/bioc
  - CRAN=https://cran.rstudio.com
```

The example above shows an alternative way of providing input packages, and input repositories,
as opposed to `inputPackageList` and `inputRepositoryList` CLI flags/YAML keys.

Additionally, `inputPackageList`/`inputRepositoryList` CLI flags take precendence over
`inputPackages`/`inputRepositories` YAML keys.

Please note that package repository URLs should be provided without the trailing `/`.

## Environment variables

`locksmith` reads environment variables with `LOCKSMITH_` prefix and tries to match them with CLI
flags. For example, setting the following variables will override the respective values from the
configuration file: `LOCKSMITH_LOGLEVEL`, `LOCKSMITH_INPUTPACKAGELIST`, `LOCKSMITH_INPUTREPOSITORYLIST` etc.

The order of precedence is:

CLI flag → environment variable → configuration file → default value.

To check the available names of environment variables, please run `locksmith --help`.

## Binary dependencies

If `locksmith` should generate an `renv.lock` with binary R packages,
it is necessary to provide URLs to binary repositories via `inputRepositories`/`inputRepositoryList`.

Examples illustrating the expected format of URLs to repositories with binary packages:

* Linux:
  * `https://packagemanager.posit.co/cran/__linux__/<distribution-name>/latest`
* Windows:
  * `https://cloud.r-project.org/bin/windows/contrib/<r-version>`
  * `https://www.bioconductor.org/packages/release/bioc/bin/windows/contrib/<r-version>`
  * `https://packagemanager.posit.co/cran/latest/bin/windows/contrib/<r-version>`
* macOS:
  * `https://cloud.r-project.org/bin/macosx/contrib/<r-version>`
  * `https://www.bioconductor.org/packages/release/bioc/bin/macosx/big-sur-arm64/contrib/<r-version>`
  * `https://www.bioconductor.org/packages/release/bioc/bin/macosx/big-sur-x86_64/contrib/<r-version>`
  * `https://packagemanager.posit.co/cran/latest/bin/macosx/big-sur-x86_64/contrib/<r-version>`
  * `https://packagemanager.posit.co/cran/latest/bin/macosx/big-sur-arm64/contrib/<r-version>`

where `<r-version>` is e.g. `4.2`, `4.3` etc.

In all cases the URL points to a directory where the `PACKAGES` file is located, without the trailing `/`.

As a result, the configuration file could look like this:

* for macOS:

    ```yaml
    inputRepositories:
      - CRAN-macOS=https://cloud.r-project.org/bin/macosx/contrib/4.2
      - Bioc-macOS=https://www.bioconductor.org/packages/release/bioc/bin/macosx/big-sur-x86_64/contrib/4.3
    ```

* for Windows:

    ```yaml
    inputRepositories:
      - CRAN-Windows=https://cloud.r-project.org/bin/windows/contrib/4.2
      - Bioc-Windows=https://www.bioconductor.org/packages/release/bioc/bin/windows/contrib/4.3
    ```

## Packages not found in the repositories

It may happen that some of the dependencies required by the input packages cannot be found in any of
the input repositories. By default, `locksmith` will fail in such case and show a list of such dependencies.

However, it is possible to override this behavior by using the `--allowIncompleteRenvLock` flag.
Simply list the types of dependencies which should not cause the `renv.lock` generation to fail:

```bash
locksmith --allowIncompleteRenvLock 'Imports,Depends,Suggests,LinkingTo'
```

## Updating existing `renv.lock`

`locksmith` has the capability to update an existing lockfile with the newest available package versions.

To ensure that the `input.renv.lock` has all the packages in the newest versions from the respective repositories (git, CRAN-like or BioConductor-like), and to save such updated file to `output.renv.lock`, you can run:

```bash
locksmith --inputRenvLock input.renv.lock --outputRenvLock output.renv.lock
```

For git packages, a reference to the latest commit on the default branch will be saved.

For packages which, according to the input lockfile, should be downloaded from CRAN-like or BioConductor-like repositories, a reference to the latest available package version in the respective repository will be saved.

The packages can be updated selectively by using the `--updatePackages` flag.

Please note that `renv` might have saved the information in the input lockfile that a package `P` should be downloaded from `CRAN`, `RSPM` or BioConductor repository, but at the same time the definition of that repository in the `renv.lock` header (in the `Repositories` section) might be missing.
In this case, `locksmith` will replicate seemingly undocumented `renv` behavior: the version of package `P` in the lockfile will be updated to the latest version found in any of the repositories **defined** in the lockfile.

Please also note that `locksmith` will not verify whether the dependencies of some packages have changed - this means that the set of package names present in the lockfile will stay the same.

## Development

This project is built with the [Go programming language](https://go.dev/).

### Development Environment

It is recommended to use Go 1.21+ for developing this project. This project uses a pre-commit
configuration and it is recommended to [install and use pre-commit](https://pre-commit.com/#install)
when you are developing this project.

### Common Commands

Run `make help` to list all related targets that will aid local development.

## License

`locksmith` is licensed under the Apache 2.0 license. See [LICENSE](LICENSE) for details.
