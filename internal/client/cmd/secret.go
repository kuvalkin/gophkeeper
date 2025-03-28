package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
)

func newSecretCommand(container container.Container) *cobra.Command {
	secret := &cobra.Command{
		Use:   "secret",
		Short: "Set encryption secret",
		Long:  "Set encryption secret. It's necessary to set it only once. It's used to encrypt and decrypt your data.",
		RunE: func(cmd *cobra.Command, args []string) error {
			srv, err := container.GetSecretService(cmd.Context())
			if err != nil {
				return fmt.Errorf("cant get secret service: %w", err)
			}

			promter, err := container.GetPrompter(cmd.Context())
			if err != nil {
				return fmt.Errorf("cant get prompter: %w", err)
			}

			_, exists, err := srv.GetSecret(cmd.Context())
			if err != nil {
				return fmt.Errorf("cant get secret: %w", err)
			}
			if exists && !promter.Confirm(cmd.Context(), "Secret already set. Do you want to overwrite it? ANY EXISTING DATA WILL BECOME UNREADABLE!") {
				return nil
			}

			secret, err := promter.AskPassword(cmd.Context(), "Enter secret", "Secret")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking secret: %w", err)
			}

			err = srv.SetSecret(cmd.Context(), secret)
			if err != nil {
				return fmt.Errorf("cant set secret: %w", err)
			}

			cmd.Println("Secret set!")

			return nil
		},
	}

	return secret
}
