package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/cmd/entries"
	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

func newGetCommand(container container.Container) *cobra.Command {
	set := &cobra.Command{
		Use:   "get",
		Short: "Get value",
		Long:  "Get value from the cloud, decrypt it and show it",
	}

	set.AddCommand(newGetLoginCommand(container))
	set.AddCommand(newGetFileCommand(container))
	set.AddCommand(newGetCardCommand(container))
	set.AddCommand(newGetTextCommand(container))

	return set
}

func newGetLoginCommand(container container.Container) *cobra.Command {
	getLogin := &cobra.Command{
		Use:   "login",
		Short: "Get login and password pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				return fmt.Errorf("name is empty")
			}

			cmd.Println("Getting login...")

			notes, content, exists, err := get(cmd.Context(), container, utils.GetEntryKey("login", name))
			if err != nil {
				return fmt.Errorf("error getting login: %w", err)
			}

			if !exists {
				cmd.Println("Login not found")

				return nil
			}

			defer utils.CloseAndLogError(content, nil)

			entry := &entries.LoginPasswordPair{}
			err = entry.Unmarshal(content)
			if err != nil {
				return fmt.Errorf("error unmarshaling login: %w", err)
			}

			cmd.Printf("Login: %s\nPassword: %s\nNotes: %s\n", entry.Login, entry.Password, notes)

			return nil
		},
	}

	return getLogin
}

func newGetFileCommand(container container.Container) *cobra.Command {
	getFile := &cobra.Command{
		Use:   "file",
		Short: "Get file",
		Long:  "Download file from the cloud, decrypts it and stores in a provided path. If the file in path already exists, it will be overwritten.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				return fmt.Errorf("name is empty")
			}
			pathToDst := args[1]
			if pathToDst == "" {
				return fmt.Errorf("path is empty")
			}

			cmd.Println("Getting file...")
			cmd.Println("It may take a while depending on the file size")

			notes, content, exists, err := get(cmd.Context(), container, utils.GetEntryKey("file", name))
			if err != nil {
				return fmt.Errorf("error getting file: %w", err)
			}

			if !exists {
				cmd.Println("File not found")

				return nil
			}

			defer utils.CloseAndLogError(content, nil)

			cmd.Println("Downloaded entry from server")
			cmd.Println("Your notes:", notes)

			cmd.Println("Storing file...")

			err = os.MkdirAll(filepath.Dir(pathToDst), 0755)
			if err != nil {
				return fmt.Errorf("error creating directory: %w", err)
			}

			dst, err := os.OpenFile(pathToDst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("error opening file: %w", err)
			}

			_, err = utils.CopyContext(cmd.Context(), dst, content)
			if err != nil {
				return fmt.Errorf("error copying file: %w", err)
			}

			cmd.Println("File stored successfully!")

			return nil
		},
	}

	return getFile
}

func newGetCardCommand(container container.Container) *cobra.Command {
	getLogin := &cobra.Command{
		Use:   "card",
		Short: "Get bank card details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				return fmt.Errorf("name is empty")
			}

			cmd.Println("Getting bank card...")

			notes, content, exists, err := get(cmd.Context(), container, utils.GetEntryKey("card", name))
			if err != nil {
				return fmt.Errorf("error getting card: %w", err)
			}

			if !exists {
				cmd.Println("Bank card not found")

				return nil
			}

			defer utils.CloseAndLogError(content, nil)

			entry := &entries.BankCard{}
			err = entry.Unmarshal(content)
			if err != nil {
				return fmt.Errorf("error unmarshaling card: %w", err)
			}

			cmd.Println("Number:", entry.Number)
			cmd.Println("Holder:", entry.HolderName)
			cmd.Println("Expiration Year:", entry.ExpirationDate.Year)
			cmd.Println("Expiration Month:", entry.ExpirationDate.Month)
			cmd.Println("CVV:", entry.CVV)
			cmd.Println("")
			cmd.Println("Notes:", notes)

			return nil
		},
	}

	return getLogin
}

func newGetTextCommand(container container.Container) *cobra.Command {
	getLogin := &cobra.Command{
		Use:   "text",
		Short: "Get text",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				return fmt.Errorf("name is empty")
			}

			cmd.Println("Getting text...")

			notes, content, exists, err := get(cmd.Context(), container, utils.GetEntryKey("text", name))
			if err != nil {
				return fmt.Errorf("error getting text: %w", err)
			}

			if !exists {
				cmd.Println("Text not found")

				return nil
			}

			defer utils.CloseAndLogError(content, nil)

			text, err := io.ReadAll(content)
			if err != nil {
				return fmt.Errorf("error reading text: %w", err)
			}

			cmd.Println("Text:")
			cmd.Println(string(text))
			cmd.Println("")
			cmd.Println("Notes:", notes)

			return nil
		},
	}

	return getLogin
}

// get retrieves an entry from the cloud by its key.
// It returns the entry's notes, content, and existence status.
func get(ctx context.Context, container container.Container, key string) (string, io.ReadCloser, bool, error) {
	service, err := container.GetEntryService(ctx)
	if err != nil {
		return "", nil, false, fmt.Errorf("error getting entry service: %w", err)
	}

	authService, err := container.GetAuthService(ctx)
	if err != nil {
		return "", nil, false, fmt.Errorf("error getting auth service: %w", err)
	}

	ctxWithToken, err := authService.AddAuthorizationHeader(ctx)
	if err != nil {
		return "", nil, false, fmt.Errorf("error setting token: %w", err)
	}

	return service.GetEntry(ctxWithToken, key)
}
