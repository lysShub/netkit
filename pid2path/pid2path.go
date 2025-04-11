package pid2path

import "fmt"

type ErrNotfound uint32

func (e ErrNotfound) NotFound() bool { return true }
func (e ErrNotfound) Error() string  { return fmt.Sprintf("not found process %d path", uint32(e)) }
