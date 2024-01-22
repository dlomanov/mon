package dumper

import (
	"github.com/dlomanov/mon/internal/storage/internal/mem"
)

func init() {
	var _ Dumper = (*NoopDumper)(nil)
}

func NewNoopDumper() *NoopDumper {
	return &NoopDumper{}
}

type NoopDumper struct {
}

func (n NoopDumper) Load(_ *mem.Storage) error {
	return nil
}

func (n NoopDumper) Dump(_ mem.Storage) error {
	return nil
}
