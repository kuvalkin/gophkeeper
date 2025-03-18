package cmd

import (
	"fmt"

	"github.com/apparentlymart/go-userdirs/userdirs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
)

var buildDate string

func NewRootCommand() *cobra.Command {
	var conf *viper.Viper
	// todo interface
	var c *container.Container

	rootCmd := &cobra.Command{
		Version: fmt.Sprintf("v0.0.1 (Build Date %s)", buildDate),
		Use:     "gkeep",
		Long:    "Gophkeeper (gkeep) is a CLI password and secret manager. Store your stuff securely both locally and in the cloud",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			conf, err = initConfig()
			if err != nil {
				return fmt.Errorf("cant init config: %w", err)
			}

			c, err = container.New(conf)
			if err != nil {
				return fmt.Errorf("cant init application service container: %w", err)
			}

			return nil
		},
		// todo start tui
		//RunE: func(cmd *cobra.Command, args []string) error {
		//	return nil
		//},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return c.Close()
		},
	}

	rootCmd.AddCommand(newRegisterCommand(c))

	set := newSetCommand(c)
	// todo guard
	//set.PersistentPreRunE = middleware.WithParentPersistentPreRunE(ensureFullSetup(set.PersistentPreRunE))

	rootCmd.AddCommand(set)
	//todo vacuum command to delete blobs without metadata

	return rootCmd
}

func initConfig() (*viper.Viper, error) {
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
}
