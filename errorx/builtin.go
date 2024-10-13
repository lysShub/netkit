package errorx

type temporaryErr struct{ error }

func Temporary(err error) bool {
	for {
		switch x := err.(type) {
		case interface{ Temporary() bool }:
			return x.Temporary()
		case interface{ Unwrap() error }:
			err = x.Unwrap()
		default:
			return false
		}
	}
}

func WrapTemp(err error) error {
	if err == nil {
		return nil
	}
	return &temporaryErr{error: err}
}
func (t *temporaryErr) Error() string   { return t.error.Error() }
func (t *temporaryErr) Unwrap() error   { return t.error }
func (t *temporaryErr) Temporary() bool { return true }

type timeoutErr struct{ error }

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

func WrapTimeout(err error) error {
	if err == nil {
		return nil
	}
	return &timeoutErr{error: err}
}
func (t *timeoutErr) Error() string { return t.error.Error() }
func (t *timeoutErr) Unwrap() error { return t.error }
func (t *timeoutErr) Timeout() bool { return true }
