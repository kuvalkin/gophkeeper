package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/support/utils"
	"github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
)

func newDeleteCommand(container container.Container) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Remove value",
		Long:  "Remove value from the cloud. This operation is final and can't be undone",
	}

	deleteCmd.AddCommand(newDeleteLoginCommand(container))
	deleteCmd.AddCommand(newDeleteFileCommand(container))
	deleteCmd.AddCommand(newDeleteCardCommand(container))
	deleteCmd.AddCommand(newDeleteTextCommand(container))

	return deleteCmd
}

func newDeleteLoginCommand(container container.Container) *cobra.Command {
	deleteLogin := &cobra.Command{
		Use:   "login",
		Short: "Delete login and password pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteEntry(cmd, container, args[0], "login")
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
			return deleteEntry(cmd, container, args[0], "file")
		},
	}

	return deleteFile
}

func newDeleteCardCommand(container container.Container) *cobra.Command {
	deleteCard := &cobra.Command{
		Use:   "card",
		Short: "Delete bank card info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteEntry(cmd, container, args[0], "card")
		},
	}

	return deleteCard
}

func newDeleteTextCommand(container container.Container) *cobra.Command {
	deleteText := &cobra.Command{
		Use:   "text",
		Short: "Delete text",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteEntry(cmd, container, args[0], "text")
		},
	}

	return deleteText
}

func deleteEntry(cmd *cobra.Command, container container.Container, name string, entryType string) error {
	if !prompts.Confirm(cmd.Context(), fmt.Sprintf("Are you sure you want to delete %s?", entryType)) {
		return nil
	}

	cmd.Printf("Deleting %s...\n", entryType)

	service, err := container.GetEntryService(cmd.Context())
	if err != nil {
		return fmt.Errorf("error getting entry service: %w", err)
	}

	authService, err := container.GetAuthService(cmd.Context())
	if err != nil {
		return fmt.Errorf("error getting token service: %w", err)
	}

	ctxWithToken, err := authService.AddAuthorizationHeader(cmd.Context())
	if err != nil {
		return fmt.Errorf("error setting token: %w", err)
	}

	key := utils.GetEntryKey(entryType, name)

	err = service.Delete(ctxWithToken, key)
	if err != nil {
		return fmt.Errorf("error deleting entry: %w", err)
	}

	cmd.Println("Successfully deleted!")

	return nil
}
