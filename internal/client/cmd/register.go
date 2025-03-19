package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
)

var ErrNoSecret = errors.New("secret not set")

type RegisterService interface {
	Register(ctx context.Context, login string, password string) error
}

func newRegisterCommand(container Container) *cobra.Command {
	register := &cobra.Command{
		Use:   "register",
		Short: "Register on server",
		Long:  "Register on a remote server. It's necessary only if you've never had an account. Existing users can simply login",
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

			service, err := container.GetRegisterService(cmd.Context())
			if err != nil {
				return fmt.Errorf("cant get auth service: %w", err)
			}

			err = service.Register(cmd.Context(), login, password)
			if err != nil {
				return fmt.Errorf("cant register: %w", err)
			}

			cmd.Println("Registered successfully!")

			return nil
		},
	}

	return register
}
