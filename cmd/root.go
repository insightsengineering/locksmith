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
	"fmt"
	"os"

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
var outputRenvLock string
var allowIncompleteRenvLock string

// In case the lists are provided as arrays in YAML configuration file:
var inputPackages []string
var inputRepositories []string

// In case the lists are provided as strings of comma-separated values
// via CLI flag or in an environment variable:
var inputPackageList string
var inputRepositoryList string

var log = logrus.New()

func setLogLevel() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.ForceColors = true
	log.SetFormatter(customFormatter)
	log.SetReportCaller(false)
	customFormatter.FullTimestamp = false
	fmt.Println("logLevel =", logLevel)
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

			fmt.Println("config =", cfgFile)
			fmt.Println("inputPackageList =", inputPackageList)
			fmt.Println("inputRepositoryList =", inputRepositoryList)
			fmt.Println("inputPackages =", inputPackages)
			fmt.Println("inputRepositories =", inputRepositories)
			fmt.Println("outputRenvLock =", outputRenvLock)
			fmt.Println("allowIncompleteRenvLock =", allowIncompleteRenvLock)

			packageDescriptionList, repositoryList, repositoryMap, allowedMissingDependencyTypes := ParseInput()
			inputDescriptionFiles := DownloadDescriptionFiles(packageDescriptionList, DownloadTextFile)
			inputPackages := ParseDescriptionFileList(inputDescriptionFiles)
			repositoryPackagesFiles := DownloadPackagesFiles(repositoryList, DownloadTextFile)
			packagesFiles := ParsePackagesFiles(repositoryPackagesFiles)
			outputPackageList := ConstructOutputPackageList(inputPackages, packagesFiles, repositoryList, allowedMissingDependencyTypes)
			renvLock := GenerateRenvLock(outputPackageList, repositoryMap)
			writeJSON(outputRenvLock, renvLock)
		},
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.locksmith.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "logLevel", "info",
		"Logging level (trace, debug, info, warn, error). ")
	rootCmd.PersistentFlags().StringVar(&inputPackageList, "inputPackageList", "",
		"Comma-separated list of URLs for raw DESCRIPTION files in git repositories for input packages.")
	rootCmd.PersistentFlags().StringVar(&inputRepositoryList, "inputRepositoryList", "",
		"Comma-separated list of package repositories URLs, sorted according to their priorities (descending).")
	rootCmd.PersistentFlags().StringVar(&gitHubToken, "gitHubToken", "",
		"Token to download non-public files from GitHub.")
	rootCmd.PersistentFlags().StringVar(&gitLabToken, "gitLabToken", "",
		"Token to download non-public files from GitLab.")
	rootCmd.PersistentFlags().StringVar(&outputRenvLock, "outputRenvLock", "renv.lock",
		"File name to save the output renv.lock file.")
	rootCmd.PersistentFlags().StringVar(&allowIncompleteRenvLock, "allowIncompleteRenvLock", "",
		"Locksmith will fail if any of dependencies of input packages cannot be found in the repositories. "+
			"However, it will not fail for comma-separated dependency types listed in this argument, e.g.: "+
			"'Imports,Depends,Suggests,LinkingTo'")

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

		// Search config in home directory with name ".locksmith" (without extension).
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
		"logLevel",
		"inputPackageList",
		"inputRepositoryList",
		"gitHubToken",
		"gitLabToken",
		"outputRenvLock",
		"allowIncompleteRenvLock",
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
