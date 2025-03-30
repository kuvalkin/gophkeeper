package entries

import (
	"bytes"
	"fmt"
	"io"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/proto"

	pb "github.com/kuvalkin/gophkeeper/pkg/proto/serialize/v1"
)

// LoginPasswordPair represents a pair of login credentials with a login and password.
type LoginPasswordPair struct {
	Login    string // Login is the username or identifier for authentication.
	Password string // Password is the secret key associated with the login.
}

// Marshal serializes the LoginPasswordPair into a protobuf message and validates it.
// Returns an io.ReadCloser containing the serialized data or an error if validation or marshaling fails.
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
		return nil, fmt.Errorf("error marshaling entry: %w", err)
	}

	return io.NopCloser(bytes.NewReader(b)), nil
}

// Unmarshal deserializes the content from an io.Reader into the LoginPasswordPair.
// Returns an error if reading or unmarshaling fails.
func (l *LoginPasswordPair) Unmarshal(content io.Reader) error {
	b, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("error reading entry: %w", err)
	}

	m := &pb.Login{}
	err = proto.Unmarshal(b, m)
	if err != nil {
		return fmt.Errorf("error unmarshaling entry: %w", err)
	}

	l.Login = m.Login
	l.Password = m.Password

	return nil
}
