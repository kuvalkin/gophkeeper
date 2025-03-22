package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
)

func newLoginCommand(container container.Container) *cobra.Command {
	login := &cobra.Command{
		Use:   "login",
		Short: "Login on server",
		Long:  "Login on a remote server. You need to have an account. If you don't have one, you can register",
		RunE: func(cmd *cobra.Command, args []string) error {
			login, err := prompts.AskString(cmd.Context(), "Enter login", "login")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking login: %w", err)
			}

			password, err := prompts.AskPassword(cmd.Context(), "Enter password", "password")
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

			err = service.Login(cmd.Context(), login, password)
			if err != nil {
				return fmt.Errorf("cant login: %w", err)
			}

			cmd.Println("You are logged in!")

			return nil
		},
	}

	return login
}
