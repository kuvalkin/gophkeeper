package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/cmd/entries"
	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/support/utils"
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

	return set
}

func newGetLoginCommand(container container.Container) *cobra.Command {
	getLogin := &cobra.Command{
		Use:   "login",
		Short: "Get login and password pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cmd.Println("Getting login...")

			notes, content, exists, err := get(cmd.Context(), container, utils.GetEntryKey("login", name))
			if err != nil {
				return fmt.Errorf("error getting login: %w", err)
			}

			if !exists {
				cmd.Println("Login not found")

				return nil
			}

			defer content.Close()

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
		Long:  "Download file from the cloud, decrypts it and stores in a provided path",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			pathToDst := args[1]

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

			defer content.Close()

			cmd.Println("Downloaded entry from server")
			cmd.Println("Your notes:", notes)

			cmd.Println("Storing file...")

			err = os.MkdirAll(path.Dir(pathToDst), 0755)
			if err != nil {
				return fmt.Errorf("error creating directory: %w", err)
			}

			dst, err := os.OpenFile(pathToDst, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("error opening file: %w", err)
			}

			_, err = io.Copy(dst, content)
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

			cmd.Println("Getting bank card...")

			notes, content, exists, err := get(cmd.Context(), container, utils.GetEntryKey("card", name))
			if err != nil {
				return fmt.Errorf("error getting card: %w", err)
			}

			if !exists {
				cmd.Println("Bank card not found")

				return nil
			}

			defer content.Close()

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

func get(ctx context.Context, container container.Container, key string) (string, io.ReadCloser, bool, error) {
	service, err := container.GetEntryService(ctx)
	if err != nil {
		return "", nil, false, fmt.Errorf("error getting entry service: %w", err)
	}

	authService, err := container.GetAuthService(ctx)
	if err != nil {
		return "", nil, false, fmt.Errorf("error getting auth service: %w", err)
	}

	ctxWithToken, err := authService.SetToken(ctx)
	if err != nil {
		return "", nil, false, fmt.Errorf("error setting token: %w", err)
	}

	return service.Get(ctxWithToken, key)
}
