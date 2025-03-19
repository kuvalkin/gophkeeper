package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	entryService "github.com/kuvalkin/gophkeeper/internal/client/service/entry"
	"github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
	pbSerialize "github.com/kuvalkin/gophkeeper/internal/proto/serialize/v1"
)

type EntryService interface {
	Set(ctx context.Context, key string, name string, entry entryService.Entry, onConflict func(errMsg string) bool) error
	Get(ctx context.Context, key string, entry entryService.Entry) (bool, error)
	Delete(ctx context.Context, key string) error
}

type TokenService interface {
	SetToken(ctx context.Context) (context.Context, error)
}

func newSetCommand(container Container) *cobra.Command {
	set := &cobra.Command{
		Use:   "set",
		Short: "Store value",
		Long:  "Store new value and sync it to the cloud. Value is E2E encrypted",
	}

	set.AddCommand(newSetLoginCommand(container))

	return set
}

func newSetLoginCommand(container Container) *cobra.Command {
	setLogin := &cobra.Command{
		Use:   "login",
		Short: "Store login and password",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			entry := &loginPasswordEntry{}

			var err error
			entry.notes, err = cmd.Flags().GetString("notes")
			if err != nil {
				return fmt.Errorf("error getting notes flag: %w", err)
			}

			entry.login, err = prompts.AskString("Enter login you want to save", "Login")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking login: %w", err)
			}

			entry.password, err = prompts.AskPassword("Enter password you want to save", "Password")
			if err != nil {
				if errors.Is(err, prompts.ErrCanceled) {
					return nil
				}

				return fmt.Errorf("error asking password: %w", err)
			}

			service, err := container.GetEntryService(cmd.Context())
			if err != nil {
				return fmt.Errorf("error getting entry service: %w", err)
			}

			tokenService, err := container.GetTokenService(cmd.Context())
			if err != nil {
				return fmt.Errorf("error getting token service: %w", err)
			}

			ctxWithToken, err := tokenService.SetToken(cmd.Context())
			if err != nil {
				return fmt.Errorf("error setting token: %w", err)
			}

			cmd.Println("Storing login...")

			// todo extract name conversion
			// todo provide feedback as service runs
			err = service.Set(ctxWithToken, fmt.Sprintf("%x", sha256.Sum256([]byte(name))), name, entry, nil)
			if err != nil {
				return fmt.Errorf("error setting login: %w", err)
			}

			cmd.Println("Login stored successfully!")

			return nil
		},
	}

	setLogin.Flags().String("notes", "", "Notes for the login entry. Will be stored encrypted along with login and password. Optional")

	return setLogin
}

type loginPasswordEntry struct {
	login    string
	password string
	notes    string
}

func (l *loginPasswordEntry) Bytes() (io.ReadCloser, error) {
	m := &pbSerialize.Login{
		Login:    l.login,
		Password: l.password,
	}

	b, err := proto.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("error marshaling login entry: %w", err)
	}

	return io.NopCloser(bytes.NewReader(b)), nil
}

func (l *loginPasswordEntry) FromBytes(reader io.Reader) error {
	b, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading login entry: %w", err)
	}

	m := &pbSerialize.Login{}
	err = proto.Unmarshal(b, m)
	if err != nil {
		return fmt.Errorf("error unmarshaling login entry: %w", err)
	}

	l.login = m.Login
	l.password = m.Password

	return nil
}

func (l *loginPasswordEntry) Notes() string {
	return l.notes
}

func (l *loginPasswordEntry) SetNotes(notes string) error {
	l.notes = notes

	return nil
}
