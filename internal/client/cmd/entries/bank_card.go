package entries

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/proto"

	pb "github.com/kuvalkin/gophkeeper/pkg/proto/serialize/v1"
)

type BankCard struct {
	Number         string
	HolderName     string
	CVV            int
	ExpirationDate ExpirationDate
}

type ExpirationDate struct {
	Year  int
	Month int
}

func (b *BankCard) Marshal() (io.ReadCloser, error) {
	m := &pb.BankCard{
		Number:     b.Number,
		HolderName: b.HolderName,
		Cvv:        int32(b.CVV),
		ExpirationDate: &pb.BankCard_ExpirationDate{
			Year:  int32(b.ExpirationDate.Year),
			Month: int32(b.ExpirationDate.Month),
		},
	}

	err := validate(m)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	bts, err := proto.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("error marshaling entry: %w", err)
	}

	return io.NopCloser(bytes.NewReader(bts)), nil
}

func validate(m *pb.BankCard) error {
	validator, err := protovalidate.New()
	if err != nil {
		return fmt.Errorf("error creating validator: %w", err)
	}

	validationErrors := make([]error, 0)

	err = validator.Validate(m)
	if err != nil {
		validationErrors = append(validationErrors, err)
	}

	err = goluhn.Validate(m.Number)
	if err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("invalid card number: %w", err))
	}

	return errors.Join(validationErrors...)
}

func (b *BankCard) Unmarshal(content io.Reader) error {
	bts, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("error reading entry: %w", err)
	}

	m := &pb.BankCard{}
	err = proto.Unmarshal(bts, m)
	if err != nil {
		return fmt.Errorf("error unmarshaling entry: %w", err)
	}

	b.Number = m.Number
	b.HolderName = m.HolderName
	b.CVV = int(m.Cvv)
	b.ExpirationDate.Year = int(m.ExpirationDate.Year)
	b.ExpirationDate.Month = int(m.ExpirationDate.Month)

	return nil
}
