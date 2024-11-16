package errorx

type temporaryErr struct{ error }

func Temporary(err error) bool {
	temp := UnwrapTo[interface{ Temporary() bool }](err)
	if temp == nil {
		return false
	} else {
		return temp.Temporary()
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
	timeout := UnwrapTo[interface{ Timeout() bool }](err)
	if timeout == nil {
		return false
	} else {
		return timeout.Timeout()
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

func UnwrapTo[To any](err error) To {
	for {
		switch x := err.(type) {
		case To:
			return x
		case interface{ Unwrap() error }:
			err = x.Unwrap()
		default:
			var to To
			return to
		}
	}
}

func NotFound(err error) bool {
	timeout := UnwrapTo[interface{ Timeout() bool }](err)
	if timeout == nil {
		return false
	} else {
		return timeout.Timeout()
	}
}

func ConnectRefused(err error) bool { return connectRefused(err) }
func NetworkUnreach(err error) bool { return networkUnreach(err) }
func BuffTooSmall(err error) bool   { return buffTooSmall(err) }
