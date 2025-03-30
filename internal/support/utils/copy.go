package utils

import (
	"context"
	"errors"
	"io"
)

const chunkSize = 32 * 1024

func CopyContext(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	copied := int64(0)

	for {
		select {
		case <-ctx.Done():
			return copied, ctx.Err()
		default:
			n, err := io.CopyN(dst, src, chunkSize)
			copied += n

			if errors.Is(err, io.EOF) {
				return copied, nil
			}

			if err != nil {
				return copied, err
			}
		}
	}
}
