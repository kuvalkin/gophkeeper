package middleware

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
)

func SecretSet(container container.Container) MW {
	return func(f CobraRunE) CobraRunE {
		return func(cmd *cobra.Command, args []string) error {
			secret, err := container.GetSecretService(cmd.Context())
			if err != nil {
				return fmt.Errorf("cant get secret service: %w", err)
			}

			_, exists, err := secret.Get(cmd.Context())
			if err != nil {
				return fmt.Errorf("cant check if secret is set: %w", err)
			}

			if !exists {
				return errors.New("secret is not set, set it first")
			}

			if f == nil {
				return nil
			}

			return f(cmd, args)
		}
	}
}
