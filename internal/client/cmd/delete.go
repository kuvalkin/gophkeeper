package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/support/utils"
)

func newDeleteCommand(container container.Container) *cobra.Command {
	set := &cobra.Command{
		Use:   "delete",
		Short: "Remove value",
		Long:  "Remove value from the cloud. This operation is final and can't be undone",
	}

	set.AddCommand(newDeleteLoginCommand(container))

	return set
}

func newDeleteLoginCommand(container container.Container) *cobra.Command {
	setLogin := &cobra.Command{
		Use:   "login",
		Short: "Delete login and password pair",
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

			cmd.Println("Deleting login...")

			err = service.Delete(ctxWithToken, utils.GetEntryKey("login", name))
			if err != nil {
				return fmt.Errorf("error deleting login: %w", err)
			}

			cmd.Println("Login deleted!")

			return nil
		},
	}

	return setLogin
}
