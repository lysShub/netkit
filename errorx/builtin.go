package errorx

type temporaryErr struct{ error }

func Temporary(err error) error {
	if err == nil {
		return nil
	}
	return &temporaryErr{error: err}
}
func IsTemporary(err error) bool {
	temp := UnwrapTo[interface{ Temporary() bool }](err)
	if temp == nil {
		return false
	} else {
		return temp.Temporary()
	}
}
func (t *temporaryErr) Error() string   { return t.error.Error() }
func (t *temporaryErr) Unwrap() error   { return t.error }
func (t *temporaryErr) Temporary() bool { return true }

type timeoutErr struct{ error }

func Timeout(err error) error {
	if err == nil {
		return nil
	}
	return &timeoutErr{error: err}
}
func IsTimeout(err error) bool {
	timeout := UnwrapTo[interface{ Timeout() bool }](err)
	if timeout == nil {
		return false
	} else {
		return timeout.Timeout()
	}
}
func (t *timeoutErr) Error() string { return t.error.Error() }
func (t *timeoutErr) Unwrap() error { return t.error }
func (t *timeoutErr) Timeout() bool { return true }

type messageErr struct {
	error
	msg string
}

func Message(err error, msgs ...string) error {
	if err == nil && len(msgs) == 0 {
		return nil
	} else {
		if len(msgs) > 0 {
			return &messageErr{error: err, msg: msgs[0]}
		} else {
			return &messageErr{error: err, msg: err.Error()}
		}
	}
}
func IsMessage(err error) bool {
	message := UnwrapTo[interface{ Message() string }](err)
	return message == nil
}
func (m *messageErr) Error() string {
	if m.error != nil {
		return m.error.Error()
	} else {
		return m.msg
	}
}
func (m *messageErr) Message() string { return m.msg }
func (t *messageErr) Unwrap() error   { return t.error }

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

type notfoundErr struct{ error }

func NotFound(err error) bool {
	timeout := UnwrapTo[interface{ Timeout() bool }](err)
	if timeout == nil {
		return false
	} else {
		return timeout.Timeout()
	}
}
func WrapNotfound(err error) error {
	if err == nil {
		return nil
	}
	return &notfoundErr{error: err}
}
func (t *notfoundErr) Error() string  { return t.error.Error() }
func (t *notfoundErr) Unwrap() error  { return t.error }
func (t *notfoundErr) NotFound() bool { return true }

type emptyErr struct{}

// used by CloseErr
var _emptyErr error = emptyErr{}

// empty but not is nil error
var ErrEmpty error = _emptyErr

func (emptyErr) Error() string { return "" }

func ConnectResed(err error) bool   { return connectResed(err) }
func ConnectRefused(err error) bool { return connectRefused(err) }
func NetworkUnreach(err error) bool { return networkUnreach(err) }
func BuffTooSmall(err error) bool   { return buffTooSmall(err) }
func AddrNotAvail(err error) bool   { return addrNotAvail(err) }
func NetTimeout(err error) bool     { return netTimeout(err) }
