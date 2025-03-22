package middleware

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
)

func NotLoggedIn(container container.Container) func(e CobraRunE) CobraRunE {
	return func(f CobraRunE) CobraRunE {
		return func(cmd *cobra.Command, args []string) error {
			auth, err := container.GetAuthService(cmd.Context())
			if err != nil {
				return fmt.Errorf("cant get auth service: %w", err)
			}

			isLoggedIn, err := auth.IsLoggedIn(cmd.Context())
			if err != nil {
				return fmt.Errorf("cant check if logged in: %w", err)
			}

			if isLoggedIn {
				return errors.New("you are already logged in, logout first")
			}

			if f == nil {
				return nil
			}

			return f(cmd, args)
		}
	}
}
