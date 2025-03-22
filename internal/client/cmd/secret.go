package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/support/keyring"
	"github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
)

func newSecretCommand() *cobra.Command {
	secret := &cobra.Command{
		Use:   "secret",
		Short: "Set encryption secret",
		Long:  "Set encryption secret. It's necessary to set it only once. It's used to encrypt and decrypt your data. You can't change secret after setting it",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			_, ok, _ := keyring.Get("secret")
			if ok {
				return errors.New("secret already set")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			secret, err := prompts.AskPassword(cmd.Context(), "Enter secret", "Secret")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking secret: %w", err)
			}

			err = keyring.Set("secret", secret)
			if err != nil {
				return fmt.Errorf("cant set secret: %w", err)
			}

			cmd.Println("Secret set!")

			return nil
		},
	}

	return secret
}
