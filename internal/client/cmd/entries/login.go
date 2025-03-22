package entries

import (
	"bytes"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	pb "github.com/kuvalkin/gophkeeper/internal/proto/serialize/v1"
)

type LoginPasswordPair struct {
	Login    string
	Password string
	notes    string
}

func (l *LoginPasswordPair) Bytes() (io.ReadCloser, error) {
	m := &pb.Login{
		Login:    l.Login,
		Password: l.Password,
	}

	b, err := proto.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("error marshaling Login entry: %w", err)
	}

	return io.NopCloser(bytes.NewReader(b)), nil
}

func (l *LoginPasswordPair) FromBytes(reader io.Reader) error {
	b, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading Login entry: %w", err)
	}

	m := &pb.Login{}
	err = proto.Unmarshal(b, m)
	if err != nil {
		return fmt.Errorf("error unmarshaling Login entry: %w", err)
	}

	l.Login = m.Login
	l.Password = m.Password

	return nil
}

func (l *LoginPasswordPair) Notes() string {
	return l.notes
}

func (l *LoginPasswordPair) SetNotes(notes string) error {
	l.notes = notes

	return nil
}
