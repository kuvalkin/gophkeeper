package cmd

import (
	"errors"
	"fmt"

	"github.com/apparentlymart/go-userdirs/userdirs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kuvalkin/gophkeeper/internal/client/support/keyring"
	"github.com/kuvalkin/gophkeeper/internal/client/transport"
	"github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
	pbAuth "github.com/kuvalkin/gophkeeper/internal/proto/auth/v1"
)

func NewRootCommand() *cobra.Command {
	dirs := userdirs.ForApp("gophkeeper", "kuvalkin", "com.kuvalkin.gophkeeper")
	v := viper.New()
	v.SetDefault("server.insecure", false)

	rootCmd := &cobra.Command{
		Version: "v0.0.1",
		Use:     "gkeep",
		Long:    "Gophkeeper (gkeep) is a CLI password and secret manager. Store your stuff securely locally and in the cloud",
		Example: "gkeep",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := v.BindPFlags(cmd.Flags())
			if err != nil {
				return fmt.Errorf("cant bind flags: %w", err)
			}

			// config.yaml, config.json, ....
			v.SetConfigName("config")

			// PWD config overrides other
			v.AddConfigPath(".")
			for _, dir := range dirs.ConfigDirs {
				// standard OS user config paths
				v.AddConfigPath(dir)
			}

			if err = v.ReadInConfig(); err != nil {
				return fmt.Errorf("cant read config: %w", err)
			}

			return nil
		},
		// todo start tui
		//RunE: func(cmd *cobra.Command, args []string) error {
		//	return nil
		//},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return transport.CloseConnection()
		},
	}

	rootCmd.AddCommand(newRegisterCommand(v))

	return rootCmd
}

func newRegisterCommand(conf *viper.Viper) *cobra.Command {
	register := &cobra.Command{
		Use:   "register",
		Short: "Register on server",
		Long:  "Register on a remote server. It's necessary only if you've never had an account. Existing users can simply register",
		RunE: func(cmd *cobra.Command, args []string) error {
			login, err := prompts.AskString("Enter login", "New login")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking login: %w", err)
			}

			password, err := prompts.AskPassword("Enter password", "New password")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking password: %w", err)
			}

			client, err := transport.GetAuthClient(conf)
			if err != nil {
				return fmt.Errorf("cant get auth client: %w", err)
			}

			s, err := client.Register(cmd.Context(), &pbAuth.RegisterRequest{Login: login, Password: password})

			if err != nil {
				//todo handle errors (conflict)
				return fmt.Errorf("error registering user: %w", err)
			}

			err = keyring.Set("token", s.Token)
			if err != nil {
				return fmt.Errorf("error saving token: %w", err)
			}

			cmd.Println("Registered successfully!")

			return nil
		},
	}

	return register
}
