package errorx

import (
	"io"

	"github.com/pkg/errors"
)

func ShortBuff(expLen, actLen int) error {
	return WrapTemp(errors.WithStack(
		errors.WithMessagef(io.ErrShortBuffer, "packet size %d, buff size %d", expLen, actLen),
	))
}
