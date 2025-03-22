package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/support/utils"
	"github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
)

func newSetCommand(container container.Container) *cobra.Command {
	set := &cobra.Command{
		Use:   "set",
		Short: "Store value",
		Long:  "Store new value and sync it to the cloud. Value is E2E encrypted",
	}

	set.AddCommand(newSetLoginCommand(container))

	return set
}

func newSetLoginCommand(container container.Container) *cobra.Command {
	setLogin := &cobra.Command{
		Use:   "login",
		Short: "Store login and password",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			entry := &loginPasswordEntry{}

			var err error
			entry.notes, err = cmd.Flags().GetString("notes")
			if err != nil {
				return fmt.Errorf("error getting notes flag: %w", err)
			}

			entry.login, err = prompts.AskString("Enter login you want to save", "Login")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking login: %w", err)
			}

			entry.password, err = prompts.AskPassword("Enter password you want to save", "Password")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking password: %w", err)
			}

			service, err := container.GetEntryService(cmd.Context())
			if err != nil {
				return fmt.Errorf("error getting entry service: %w", err)
			}

			tokenService, err := container.GetAuthService(cmd.Context())
			if err != nil {
				return fmt.Errorf("error getting token service: %w", err)
			}

			ctxWithToken, err := tokenService.SetToken(cmd.Context())
			if err != nil {
				return fmt.Errorf("error setting token: %w", err)
			}

			cmd.Println("Storing login...")

			// todo provide feedback as service runs
			err = service.Set(ctxWithToken, utils.GetEntryKey("login", name), name, entry)
			if err != nil {
				return fmt.Errorf("error setting login: %w", err)
			}

			cmd.Println("Login stored successfully!")

			return nil
		},
	}

	setLogin.Flags().String("notes", "", "Notes for the login entry. Will be stored encrypted along with login and password. Optional")

	return setLogin
}
