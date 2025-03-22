package cmd

import (
	"bytes"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	pbSerialize "github.com/kuvalkin/gophkeeper/internal/proto/serialize/v1"
)

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
