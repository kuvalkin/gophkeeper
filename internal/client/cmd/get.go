package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/cmd/entries"
	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/support/utils"
)

func newGetCommand(container container.Container) *cobra.Command {
	set := &cobra.Command{
		Use:   "get",
		Short: "Get value",
		Long:  "Get value from the cloud, decrypt it and show it",
	}

	set.AddCommand(newGetLoginCommand(container))

	return set
}

func newGetLoginCommand(container container.Container) *cobra.Command {
	setLogin := &cobra.Command{
		Use:   "login",
		Short: "Get login and password pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			service, err := container.GetEntryService(cmd.Context())
			if err != nil {
				return fmt.Errorf("error getting entry service: %w", err)
			}

			authService, err := container.GetAuthService(cmd.Context())
			if err != nil {
				return fmt.Errorf("error getting token service: %w", err)
			}

			ctxWithToken, err := authService.SetToken(cmd.Context())
			if err != nil {
				return fmt.Errorf("error setting token: %w", err)
			}

			cmd.Println("Getting login...")

			notes, content, exists, err := service.Get(ctxWithToken, utils.GetEntryKey("login", name))
			if err != nil {
				return fmt.Errorf("error getting login: %w", err)
			}

			if !exists {
				cmd.Println("Login not found")

				return nil
			}

			defer content.Close()

			entry := &entries.LoginPasswordPair{}
			err = entry.Unmarshal(content)
			if err != nil {
				return fmt.Errorf("error unmarshaling login: %w", err)
			}

			cmd.Printf("Login: %s\nPassword: %s\nNotes: %s\n", entry.Login, entry.Password, notes)

			return nil
		},
	}

	return setLogin
}
