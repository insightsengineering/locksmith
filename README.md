# locksmith

[![build](https://github.com/insightsengineering/locksmith/actions/workflows/test.yml/badge.svg)](https://github.com/insightsengineering/locksmith/actions/workflows/test.yml)

`locksmith` is a utility to generate `renv.lock` file containing all dependencies of given set of R packages.

Given the input list of R packages or git repositories containing the R packages, as well as a list of R package repositories (e.g. in a package manager, CRAN, BioConductor etc.), `locksmith` will try to determine the list of all dependencies and their versions required to make the input list of packages work. It will then save the result in an `renv.lock`-compatible file.

## Installing

Simply download the project for your distribution from the [releases](https://github.com/insightsengineering/locksmith/releases) page. `locksmith` is distributed as a single binary file and does not require any additional system requirements.

## Usage

`locksmith` is a command line utility, so after installing the binary in your `PATH`, simply run the following command to view its capabilities:

```bash
locksmith --help
```

Example usage with multiple flags:

```bash
locksmith --logLevel debug --exampleParameter 'exampleValue'
```

Real-life example with multiple input packages and repositories.

```bash
locksmith --inputPackageList https://raw.githubusercontent.com/insightsengineering/formatters/main/DESCRIPTION,https://raw.githubusercontent.com/insightsengineering/rtables/main/DESCRIPTION,https://raw.githubusercontent.com/insightsengineering/scda/main/DESCRIPTION,https://raw.githubusercontent.com/insightsengineering/scda.2022/main/DESCRIPTION,https://raw.githubusercontent.com/insightsengineering/nestcolor/main/DESCRIPTION,https://raw.githubusercontent.com/insightsengineering/tern/main/DESCRIPTION,https://raw.githubusercontent.com/insightsengineering/rlistings/main/DESCRIPTION --inputRepositoryList BioC=https://bioconductor.org/packages/release/bioc,CRAN=https://cran.rstudio.com/
```

In order to download the packages from GitHub or GitLab repositories, please set the environment variables containing the Personal Access Tokens.
* For GitHub, set the `LOCKSMITH_GITHUBTOKEN` environment variable.
* For GitLab, set the `LOCKSMITH_GITLABTOKEN` environment variable.

By default `locksmith` will save the resulting output file to `renv.lock`.

## Configuration file

If you'd like to set the above options in a configuration file, by default `locksmith` checks `~/.locksmith`, `~/.locksmith.yaml` and `~/.locksmith.yml` files.
If this file exists, `locksmith` uses options defined there, unless they are overridden by command line flags.

You can also specify custom path to configuration file with `--config <your-configuration-file>.yml` command line flag.
When using custom configuration file, if you specify command line flags, the latter will still take precedence.

Example contents of configuration file:

```yaml
logLevel: trace
exampleParameter: exampleValue
```

## Environment variables

`locksmith` reads environment variables with `LOCKSMITH_` prefix and tries to match them with CLI flags.
For example, setting the following variables will override the respective values from configuration file:
`LOCKSMITH_LOGLEVEL`, `LOCKSMITH_EXAMPLEPARAMETER` etc.

The order of precedence is:

CLI flag → environment variable → configuration file → default value.

## Development

This project is built with the [Go programming language](https://go.dev/).

### Development Environment

It is recommended to use Go 1.21+ for developing this project. This project uses a pre-commit configuration and it is recommended to [install and use pre-commit](https://pre-commit.com/#install) when you are developing this project.

### Common Commands

Run `make help` to list all related targets that will aid local development.

## License

`locksmith` is licensed under the Apache 2.0 license. See [LICENSE](LICENSE) for details.
