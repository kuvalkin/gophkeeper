package entries

import (
	"bytes"
	"fmt"
	"io"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/proto"

	pb "github.com/kuvalkin/gophkeeper/internal/proto/serialize/v1"
)

type LoginPasswordPair struct {
	Login    string
	Password string
}

func (l *LoginPasswordPair) Marshal() (io.ReadCloser, error) {
	m := &pb.Login{
		Login:    l.Login,
		Password: l.Password,
	}

	validator, err := protovalidate.New()
	if err != nil {
		return nil, fmt.Errorf("error creating validator: %w", err)
	}

	err = validator.Validate(m)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	b, err := proto.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("error marshaling Login entry: %w", err)
	}

	return io.NopCloser(bytes.NewReader(b)), nil
}

func (l *LoginPasswordPair) Unmarshal(content io.Reader) error {
	b, err := io.ReadAll(content)
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
