package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
)

func newLogoutCommand(container container.Container) *cobra.Command {
	logout := &cobra.Command{
		Use:   "logout",
		Short: "Logout",
		Long:  "Logout from the remote server",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := container.GetAuthService(cmd.Context())
			if err != nil {
				return fmt.Errorf("cant get auth service: %w", err)
			}

			err = service.Logout(cmd.Context())
			if err != nil {
				return fmt.Errorf("cant logout: %w", err)
			}

			cmd.Println("You are logged out!")

			return nil
		},
	}

	return logout
}
