package cmd

import (
	"context"
	"fmt"

	"github.com/apparentlymart/go-userdirs/userdirs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

var buildDate string

type Container interface {
	GetEntryService(ctx context.Context) (EntryService, error)
	GetAuthService(ctx context.Context) (RegisterService, error)
}

func NewRootCommand(container Container) *cobra.Command {
	rootCmd := &cobra.Command{
		Version: fmt.Sprintf("v0.0.1 (Build Date %s)", buildDate),
		Use:     "gkeep",
		Long:    "Gophkeeper (gkeep) is a CLI password and secret manager. Store your stuff securely both locally and in the cloud",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return fmt.Errorf("error getting verbose flag: %w", err)
			}

			if verbose {
				err = log.InitLogger()
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

	rootCmd.AddCommand(newSecretCommand())
	rootCmd.AddCommand(newRegisterCommand(container))

	set := newSetCommand(container)
	// todo guard
	//set.PersistentPreRunE = middleware.WithParentPersistentPreRunE(ensureFullSetup(set.PersistentPreRunE))

	rootCmd.AddCommand(set)
	//todo vacuum command to delete blobs without metadata

	return rootCmd
}

func NewConfig() (*viper.Viper, error) {
	conf := viper.New()

	dirs := userdirs.ForApp("gophkeeper", "kuvalkin", "com.kuvalkin.gophkeeper")

	defaultConfig(conf, &dirs)

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

func defaultConfig(conf *viper.Viper, dirs *userdirs.Dirs) {
	conf.SetDefault("server.insecure", false)

	conf.SetDefault("storage.sqlite.path", dirs.NewDataPath("metadata.sqlite"))
	conf.SetDefault("storage.blobs.path", dirs.NewDataPath("blobs"))

	conf.SetDefault("log.path", "")
}
