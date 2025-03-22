package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/kuvalkin/gophkeeper/internal/client/cmd/entries"
	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/support/utils"
	"github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
)

func newSetCommand(container container.Container) *cobra.Command {
	set := &cobra.Command{
		Use:   "set",
		Short: "Store entry",
		Long:  "Store new entry and sync it to the cloud. Entry is E2E encrypted",
	}

	set.PersistentFlags().String("notes", "", "Notes for the entry. Will be stored encrypted along with the entry itself. Optional")

	set.AddCommand(newSetLoginCommand(container))
	set.AddCommand(newSetFileCommand(container))
	set.AddCommand(newSetCardCommand(container))

	return set
}

func newSetLoginCommand(container container.Container) *cobra.Command {
	setLogin := &cobra.Command{
		Use:   "login",
		Short: "Store login and password pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			notes, err := cmd.Flags().GetString("notes")
			if err != nil {
				return fmt.Errorf("error getting notes flag: %w", err)
			}

			entry := &entries.LoginPasswordPair{}

			entry.Login, err = prompts.AskString(cmd.Context(), "Enter login you want to save", "Login")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking login: %w", err)
			}

			entry.Password, err = prompts.AskPassword(cmd.Context(), "Enter password you want to save", "Password")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking password: %w", err)
			}

			content, err := entry.Marshal()
			if err != nil {
				return fmt.Errorf("error getting entry bytes: %w", err)
			}

			cmd.Println("Storing login...")

			err = store(cmd.Context(), container, utils.GetEntryKey("login", name), name, notes, content)
			if err != nil {
				return fmt.Errorf("error storing login: %w", err)
			}

			cmd.Println("Login stored successfully!")

			return nil
		},
	}

	return setLogin
}

func newSetFileCommand(container container.Container) *cobra.Command {
	setFile := &cobra.Command{
		Use:   "file",
		Short: "Store file with any content",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			path := args[1]

			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("error opening file: %w", err)
			}

			notes, err := cmd.Flags().GetString("notes")
			if err != nil {
				return fmt.Errorf("error getting notes flag: %w", err)
			}

			cmd.Println("Storing file...")
			cmd.Println("It may take a while depending on the file size")

			err = store(cmd.Context(), container, utils.GetEntryKey("file", name), name, notes, file)
			if err != nil {
				return fmt.Errorf("error storing file: %w", err)
			}

			cmd.Println("File stored successfully!")

			return nil
		},
	}

	return setFile
}

func newSetCardCommand(container container.Container) *cobra.Command {
	setCard := &cobra.Command{
		Use:   "card",
		Short: "Store bank card info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			notes, err := cmd.Flags().GetString("notes")
			if err != nil {
				return fmt.Errorf("error getting notes flag: %w", err)
			}

			entry := &entries.BankCard{}

			entry.Number, err = prompts.AskString(cmd.Context(), "Enter bank card number", "xxxxxxxxxxxxxxxx")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking card number: %w", err)
			}

			entry.HolderName, err = prompts.AskString(cmd.Context(), "Enter card holder name", "John Doe")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking card holder name: %w", err)
			}

			entry.ExpirationDate.Year, err = prompts.AskInt(cmd.Context(), "Enter card expiration year", "2030")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking expiration year: %w", err)
			}

			entry.ExpirationDate.Month, err = prompts.AskInt(cmd.Context(), "Enter card expiration month", "05")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking expiration month: %w", err)
			}

			entry.CVV, err = prompts.AskInt(cmd.Context(), "Enter card CVV", "123")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking cvv: %w", err)
			}

			content, err := entry.Marshal()
			if err != nil {
				return fmt.Errorf("error getting entry bytes: %w", err)
			}

			cmd.Println("Storing bank card...")

			err = store(cmd.Context(), container, utils.GetEntryKey("card", name), name, notes, content)
			if err != nil {
				return fmt.Errorf("error storing card: %w", err)
			}

			cmd.Println("Bank card stored successfully!")

			return nil
		},
	}

	return setCard
}

func store(ctx context.Context, container container.Container, key string, name string, notes string, content io.ReadCloser) error {
	service, err := container.GetEntryService(ctx)
	if err != nil {
		return fmt.Errorf("error getting entry service: %w", err)
	}

	tokenService, err := container.GetAuthService(ctx)
	if err != nil {
		return fmt.Errorf("error getting token service: %w", err)
	}

	ctxWithToken, err := tokenService.SetToken(ctx)
	if err != nil {
		return fmt.Errorf("error setting token: %w", err)
	}

	// todo provide feedback as service runs
	err = service.Set(ctxWithToken, key, name, notes, content)
	if err != nil {
		return fmt.Errorf("error setting entry: %w", err)
	}

	return nil
}
