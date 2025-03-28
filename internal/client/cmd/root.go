package cmd

import (
	"fmt"
	"os"

	"github.com/apparentlymart/go-userdirs/userdirs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kuvalkin/gophkeeper/internal/client/cmd/middleware"
	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

var buildDate string

func NewRootCommand(container container.Container) *cobra.Command {
	rootCmd := &cobra.Command{
		Version: fmt.Sprintf("v0.0.1 (Build Date %s)", buildDate),
		Use:     "gkeep",
		Long:    "Gophkeeper (gkeep) is a CLI password and secret manager. Store your stuff securely in the cloud",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return fmt.Errorf("error getting verbose flag: %w", err)
			}

			if verbose {
				err = log.InitClientLogger()
				if err != nil {
					return fmt.Errorf("error initializing logger: %w", err)
				}
			}

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return fmt.Errorf("error getting verbose flag: %w", err)
			}

			if verbose {
				_ = log.Logger().Sync()
			}

			return nil
		},
	}
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output. Useful for debugging")

	secret := newSecretCommand(container)
	rootCmd.AddCommand(secret)

	ensureNotLoggedIn := middleware.EnsureNotLoggedIn(container)

	register := newRegisterCommand(container)
	register.PreRunE = ensureNotLoggedIn(register.PreRunE)
	rootCmd.AddCommand(register)

	login := newLoginCommand(container)
	login.PreRunE = ensureNotLoggedIn(login.PreRunE)
	rootCmd.AddCommand(login)

	rootCmd.AddCommand(newLogoutCommand(container))

	ensureSecretSet := middleware.EnsureSecretSet(container)
	ensureLoggedIn := middleware.EnsureLoggedIn(container)

	set := newSetCommand(container)
	set.PersistentPreRunE = middleware.Combine(rootCmd.PersistentPreRunE, ensureSecretSet(ensureLoggedIn(set.PersistentPreRunE)))
	rootCmd.AddCommand(set)

	getCmd := newGetCommand(container)
	getCmd.PersistentPreRunE = middleware.Combine(rootCmd.PersistentPreRunE, ensureSecretSet(ensureLoggedIn(getCmd.PersistentPreRunE)))
	rootCmd.AddCommand(getCmd)

	deleteCmd := newDeleteCommand(container)
	deleteCmd.PersistentPreRunE = middleware.Combine(rootCmd.PersistentPreRunE, ensureLoggedIn(deleteCmd.PersistentPreRunE))
	rootCmd.AddCommand(deleteCmd)

	rootCmd.AddCommand(newConfigPathCommand())

	return rootCmd
}

func NewConfig() (*viper.Viper, error) {
	conf := viper.New()

	dirs := getUserDirs()

	defaultConfig(conf)

	// config.yaml, config.json, ....
	conf.SetConfigName("config")

	// PWD config overrides other
	conf.AddConfigPath(".")
	for _, dir := range dirs.ConfigDirs {
		// standard OS user config paths
		conf.AddConfigPath(dir)
	}

	if err := conf.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cant read config: %w", err)
	}

	return conf, nil
}

func defaultConfig(conf *viper.Viper) {
	conf.SetDefault("server.insecure", false)
	conf.SetDefault("stream.chunk_size", 1024*1024) // 1 MB
}

func getUserDirs() userdirs.Dirs {
	return userdirs.ForApp("gophkeeper", "kuvalkin", "com.kuvalkin.gophkeeper")
}

func newConfigPathCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "paths",
		Short:  "Print config path",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Config dirs:")

			pwd, err := os.Getwd()
			if err == nil {
				cmd.Println(pwd)
			}
			dirs := getUserDirs()
			for _, dir := range dirs.ConfigDirs {
				cmd.Println(dir)
			}
		},
	}
}
