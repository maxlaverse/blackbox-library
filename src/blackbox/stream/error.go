package stream

import (
	"github.com/pkg/errors"
	"io"
)

func errorWithStack(err error) error {
	if err == io.EOF {
		return err
	} else {
		return errors.WithStack(err)
	}
}
