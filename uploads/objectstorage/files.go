package objectstorage

import (
	"context"
	"io"

	"github.com/verygoodsoftwarenotvirus/platform/v4/errors"
)

// SaveFile saves a file to the blob.
func (u *Uploader) SaveFile(ctx context.Context, path string, content []byte) error {
	ctx, span := u.tracer.StartSpan(ctx)
	defer span.End()

	if err := u.bucket.WriteAll(ctx, path, content, nil); err != nil {
		return errors.Wrap(err, "writing file content")
	}

	return nil
}

// ReadFile reads a file from the blob.
func (u *Uploader) ReadFile(ctx context.Context, path string) ([]byte, error) {
	ctx, span := u.tracer.StartSpan(ctx)
	defer span.End()

	r, err := u.bucket.NewReader(ctx, path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "fetching file")
	}

	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			u.logger.Error("error closing file reader", closeErr)
		}
	}()

	fileBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "reading file")
	}

	return fileBytes, nil
}
