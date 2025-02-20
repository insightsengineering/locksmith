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

// Package cmd implements the core functionality of locksmith - the renv.lock generator.
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/jamiealquiza/envy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.szostok.io/version/extension"
)

var cfgFile string
var logLevel string
var gitHubToken string
var gitLabToken string
var inputRenvLock string
var outputRenvLock string
var allowIncompleteRenvLock string
var updatePackages string
var reportFileName string

// In case the lists are provided as arrays in YAML configuration file:
var inputPackages []string
var inputRepositories []string

// In case the lists are provided as strings of comma-separated values
// via CLI flag or in an environment variable:
var inputPackageList string
var inputRepositoryList string

var localTempDirectory string

// Messages with level warning are saved to warnBuffer.
// Messages with level error or higher are saved to errorBuffer.
// This is required for the HTML report.
var warnBuffer, errorBuffer bytes.Buffer

var log = logrus.New()

type WarningCaptureHook struct {
	WarnEntries *bytes.Buffer
}

func (hook *WarningCaptureHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel}
}

func (hook *WarningCaptureHook) Fire(entry *logrus.Entry) error {
	if entry.Level == logrus.WarnLevel {
		hook.WarnEntries.WriteString(strings.TrimSpace(entry.Message) + "\n")
	}
	return nil
}

type ErrorCaptureHook struct {
	ErrorEntries *bytes.Buffer
}

func (hook *ErrorCaptureHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}

func (hook *ErrorCaptureHook) Fire(entry *logrus.Entry) error {
	if entry.Level >= logrus.ErrorLevel {
		hook.ErrorEntries.WriteString(strings.TrimSpace(entry.Message) + "\n")
	}
	return nil
}

func setLogLevel() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.ForceColors = true
	log.SetFormatter(customFormatter)
	log.SetReportCaller(false)
	customFormatter.FullTimestamp = false

	// Log messages with level exactly equal to warning to warnBuffer.
	warnCaptureHook := &WarningCaptureHook{
		WarnEntries: &warnBuffer,
	}
	// Log messages with level equal to or higher than error to errorBuffer.
	errorCaptureHook := &ErrorCaptureHook{
		ErrorEntries: &errorBuffer,
	}
	// Add the custom hooks.
	log.AddHook(warnCaptureHook)
	log.AddHook(errorCaptureHook)

	fmt.Println(`logLevel = "` + logLevel + `"`)
	switch logLevel {
	case "trace":
		log.SetLevel(logrus.TraceLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

var rootCmd *cobra.Command

//nolint:revive
func newRootCommand() {
	rootCmd = &cobra.Command{
		Use:   "locksmith",
		Short: "renv.lock generator",
		Long: `locksmith is a utility to generate renv.lock file containing all dependencies
of given set of R packages. Given the input list of git repositories containing the R packages,
as well as a list of R package repositories (e.g. in a package manager, CRAN,
BioConductor etc.), locksmith will try to determine the list of all dependencies and their
versions required to make the input list of packages work. It will then save the result
in an renv.lock-compatible file.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initializeConfig()
		},
		Run: func(cmd *cobra.Command, args []string) {
			setLogLevel()

			fmt.Println(`config = "` + cfgFile + `"`)
			fmt.Println(`inputPackageList = "` + inputPackageList + `"`)
			fmt.Println(`inputRepositoryList = "` + inputRepositoryList + `"`)
			fmt.Println("inputPackages =", inputPackages)
			fmt.Println("inputRepositories =", inputRepositories)
			fmt.Println(`inputRenvLock = "` + inputRenvLock + `"`)
			fmt.Println(`outputRenvLock = "` + outputRenvLock + `"`)
			fmt.Println(`allowIncompleteRenvLock = "` + allowIncompleteRenvLock + `"`)
			fmt.Println(`updatePackages = "` + updatePackages + `"`)
			fmt.Println(`reportFileName = "` + reportFileName + `"`)

			if runtime.GOOS == "windows" {
				localTempDirectory = os.Getenv("TMP") + `\tmp\locksmith`
			} else {
				localTempDirectory = "/tmp/locksmith"
			}

			if inputRenvLock != "" {
				renvLock := UpdateRenvLock(inputRenvLock, updatePackages)
				writeJSON(outputRenvLock, renvLock)
			} else {
				packageDescriptionList, repositoryList, repositoryMap, allowedMissingDependencyTypes := ParseInput()
				inputDescriptionFiles := DownloadDescriptionFiles(packageDescriptionList, DownloadTextFile)
				inputPackages := ParseDescriptionFileList(inputDescriptionFiles)
				repositoryPackagesFiles := DownloadPackagesFiles(repositoryList, DownloadTextFile)
				packagesFiles := ParsePackagesFiles(repositoryPackagesFiles)
				outputPackageList := ConstructOutputPackageList(inputPackages, packagesFiles, repositoryList, allowedMissingDependencyTypes)
				renvLock := GenerateRenvLock(outputPackageList, repositoryMap)
				GenerateHTMLReport(outputPackageList, inputPackages, packagesFiles, renvLock, repositoryMap)
				writeJSON(outputRenvLock, renvLock)
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		"config file (default is $HOME/.locksmith.yaml)")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "logLevel", "l", "info",
		"Logging level (trace, debug, info, warn, error). ")
	rootCmd.PersistentFlags().StringVarP(&inputPackageList, "inputPackageList", "p", "",
		"Comma-separated list of URLs for raw DESCRIPTION files in git repositories for input packages.")
	rootCmd.PersistentFlags().StringVarP(&inputRepositoryList, "inputRepositoryList", "r", "",
		"Comma-separated list of package repositories URLs, sorted according to their priorities (descending).")
	rootCmd.PersistentFlags().StringVarP(&gitHubToken, "gitHubToken", "t", "",
		"Token to download non-public files from GitHub.")
	rootCmd.PersistentFlags().StringVarP(&gitLabToken, "gitLabToken", "g", "",
		"Token to download non-public files from GitLab.")
	rootCmd.PersistentFlags().StringVarP(&inputRenvLock, "inputRenvLock", "n", "",
		"Lockfile which should be read and updated to include the newest versions of the packages.")
	rootCmd.PersistentFlags().StringVarP(&outputRenvLock, "outputRenvLock", "k", "renv.lock",
		"File name to save the output renv.lock file.")
	rootCmd.PersistentFlags().StringVarP(&allowIncompleteRenvLock, "allowIncompleteRenvLock", "i", "",
		"Locksmith will fail if any of dependencies of input packages cannot be found in the repositories. "+
			"However, it will not fail for comma-separated dependency types listed in this argument, e.g.: "+
			"'Imports,Depends,Suggests,LinkingTo'")
	rootCmd.PersistentFlags().StringVarP(&updatePackages, "updatePackages", "u", "*",
		"Expression with wildcards indicating which packages from the inputRenvLock should be updated to the newest version. "+
			"The expression follows the pattern: \"expression1,expression2,...\" where \"expressionN\" can be: "+
			"literal package name and/or * symbol(s) meaning any set of characters. Example: "+
			`'package*,*abc,a*b,someOtherPackage'. By default all packages are updated.`)
	rootCmd.PersistentFlags().StringVarP(&reportFileName, "reportFileName", "f", "locksmithReport.html",
		"File name to save the output report.")

	// Add version command.
	rootCmd.AddCommand(extension.NewVersionCobraCmd())

	cfg := envy.CobraConfig{
		Prefix:     "LOCKSMITH",
		Persistent: true,
	}
	envy.ParseCobra(rootCmd, cfg)
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search for config in home directory.
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".locksmith")
	}
	// Read in environment variables that match.
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println(err)
	}
}

func Execute() {
	newRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initializeConfig() {
	for _, v := range []string{
		"logLevel", "inputPackageList", "inputRepositoryList", "gitHubToken", "gitLabToken",
		"inputRenvLock", "outputRenvLock", "allowIncompleteRenvLock", "updatePackages",
		"reportFileName",
	} {
		// If the flag has not been set in newRootCommand() and it has been set in initConfig().
		// In other words: if it's not been provided in command line, but has been
		// provided in config file.
		// Helpful project where it's explained:
		// https://github.com/carolynvs/stingoftheviper
		if !rootCmd.PersistentFlags().Lookup(v).Changed && viper.IsSet(v) {
			err := rootCmd.PersistentFlags().Set(v, fmt.Sprintf("%v", viper.Get(v)))
			checkError(err)
		}
	}
	// Check if a YAML list of input packages or input repositories has been provided in the configuration file.
	inputPackages = viper.GetStringSlice("inputPackages")
	inputRepositories = viper.GetStringSlice("inputRepositories")
}
