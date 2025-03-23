package cmd

import (
	"fmt"

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

	notLoggedIn := middleware.NotLoggedIn(container)

	register := newRegisterCommand(container)
	register.PreRunE = notLoggedIn(register.PreRunE)
	rootCmd.AddCommand(register)

	login := newLoginCommand(container)
	login.PreRunE = notLoggedIn(login.PreRunE)
	rootCmd.AddCommand(login)

	rootCmd.AddCommand(newLogoutCommand(container))

	secretSet := middleware.SecretSet(container)
	loggedIn := middleware.LoggedIn(container)

	set := newSetCommand(container)
	set.PersistentPreRunE = middleware.Combine(rootCmd.PersistentPreRunE, secretSet(loggedIn(set.PersistentPreRunE)))
	rootCmd.AddCommand(set)

	getCmd := newGetCommand(container)
	getCmd.PersistentPreRunE = middleware.Combine(rootCmd.PersistentPreRunE, secretSet(loggedIn(getCmd.PersistentPreRunE)))
	rootCmd.AddCommand(getCmd)

	deleteCmd := newDeleteCommand(container)
	deleteCmd.PersistentPreRunE = middleware.Combine(rootCmd.PersistentPreRunE, loggedIn(deleteCmd.PersistentPreRunE))
	rootCmd.AddCommand(deleteCmd)

	return rootCmd
}

func NewConfig() (*viper.Viper, error) {
	conf := viper.New()

	dirs := userdirs.ForApp("gophkeeper", "kuvalkin", "com.kuvalkin.gophkeeper")

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
