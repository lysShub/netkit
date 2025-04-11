package pid2path

import "github.com/pkg/errors"

type Pid2Path struct {
}

func New() (*Pid2Path, error) {
	return nil, errors.New("not support")
}

func (p *Pid2Path) Path(pid uint32) (path string, err error) { panic("") }
