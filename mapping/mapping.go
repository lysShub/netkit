package mapping

type ErrNotRecord struct{}

func (e ErrNotRecord) Error() string { return "not record" }
func (ErrNotRecord) Temporary() bool { return true }
