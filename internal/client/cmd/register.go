package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/service/auth"
	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
)

func newRegisterCommand(container container.Container) *cobra.Command {
	register := &cobra.Command{
		Use:   "register",
		Short: "Register on server",
		Long:  "Register on a remote server. It's necessary only if you've never had an account. Existing users can simply login",
		RunE: func(cmd *cobra.Command, args []string) error {
			login, err := prompts.AskString(cmd.Context(), "Enter login", "New login")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking login: %w", err)
			}

			password, err := prompts.AskPassword(cmd.Context(), "Enter password", "New password")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking password: %w", err)
			}

			service, err := container.GetAuthService(cmd.Context())
			if err != nil {
				return fmt.Errorf("cant get auth service: %w", err)
			}

			err = service.Register(cmd.Context(), login, password)
			if err != nil {
				if errors.Is(err, auth.ErrLoginTaken) {
					return fmt.Errorf("user with this login already exists, pick another one")
				}

				return fmt.Errorf("cant register: %w", err)
			}

			cmd.Println("Registered successfully!")

			return nil
		},
	}

	return register
}
