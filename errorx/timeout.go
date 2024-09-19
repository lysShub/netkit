package errorx

func Timeout(err error) bool {
	for {
		switch x := err.(type) {
		case interface{ Timeout() bool }:
			return x.Timeout()
		case interface{ Unwrap() error }:
			err = x.Unwrap()
		default:
			return false
		}
	}
}

type timeoutErr struct {
	error
}

func WrapTimeout(err error) error {
	if err == nil {
		return nil
	}
	return &timeoutErr{error: err}
}
func (t *timeoutErr) Error() string { return t.error.Error() }
func (t *timeoutErr) Unwrap() error { return t.error }
func (t *timeoutErr) Timeout() bool { return true }
