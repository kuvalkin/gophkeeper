package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/support/utils"
)

func newDeleteCommand(container container.Container) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Remove value",
		Long:  "Remove value from the cloud. This operation is final and can't be undone",
	}

	deleteCmd.AddCommand(newDeleteLoginCommand(container))
	deleteCmd.AddCommand(newDeleteFileCommand(container))

	return deleteCmd
}

func newDeleteLoginCommand(container container.Container) *cobra.Command {
	deleteLogin := &cobra.Command{
		Use:   "login",
		Short: "Delete login and password pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cmd.Println("Deleting login...")

			err := deleteEntry(cmd.Context(), container, utils.GetEntryKey("login", name))
			if err != nil {
				return fmt.Errorf("error deleting login: %w", err)
			}

			cmd.Println("Login deleted!")

			return nil
		},
	}

	return deleteLogin
}

func newDeleteFileCommand(container container.Container) *cobra.Command {
	deleteFile := &cobra.Command{
		Use:   "file",
		Short: "Delete file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cmd.Println("Deleting file...")

			err := deleteEntry(cmd.Context(), container, utils.GetEntryKey("file", name))
			if err != nil {
				return fmt.Errorf("error deleting file: %w", err)
			}

			cmd.Println("File deleted!")

			return nil
		},
	}

	return deleteFile
}

func deleteEntry(ctx context.Context, container container.Container, key string) error {
	service, err := container.GetEntryService(ctx)
	if err != nil {
		return fmt.Errorf("error getting entry service: %w", err)
	}

	authService, err := container.GetAuthService(ctx)
	if err != nil {
		return fmt.Errorf("error getting token service: %w", err)
	}

	ctxWithToken, err := authService.SetToken(ctx)
	if err != nil {
		return fmt.Errorf("error setting token: %w", err)
	}

	err = service.Delete(ctxWithToken, key)
	if err != nil {
		return fmt.Errorf("error deleting entry: %w", err)
	}

	return nil
}
