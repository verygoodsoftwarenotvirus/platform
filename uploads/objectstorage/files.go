package objectstorage

import (
	"context"
	"fmt"
	"io"
)

// SaveFile saves a file to the blob.
func (u *Uploader) SaveFile(ctx context.Context, path string, content []byte) error {
	ctx, span := u.tracer.StartSpan(ctx)
	defer span.End()

	if err := u.bucket.WriteAll(ctx, path, content, nil); err != nil {
		return fmt.Errorf("writing file content: %w", err)
	}

	return nil
}

// ReadFile reads a file from the blob.
func (u *Uploader) ReadFile(ctx context.Context, path string) ([]byte, error) {
	ctx, span := u.tracer.StartSpan(ctx)
	defer span.End()

	r, err := u.bucket.NewReader(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching file: %w", err)
	}

	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			u.logger.Error("error closing file reader", closeErr)
		}
	}()

	fileBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return fileBytes, nil
}
