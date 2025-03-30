package middleware

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
)

// EnsureLoggedIn ensures that the user is logged in before executing the provided CobraRunE function.
// If the user is not logged in, it returns an error prompting the user to log in or register first.
func EnsureLoggedIn(container container.Container) MW {
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

			if !isLoggedIn {
				return errors.New("not logged in, login or register first")
			}

			if f == nil {
				return nil
			}

			return f(cmd, args)
		}
	}
}

// EnsureNotLoggedIn ensures that the user is not logged in before executing the provided CobraRunE function.
// If the user is already logged in, it returns an error prompting the user to log out first.
func EnsureNotLoggedIn(container container.Container) MW {
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
