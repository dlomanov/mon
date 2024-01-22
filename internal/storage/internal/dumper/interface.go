package dumper

import "github.com/dlomanov/mon/internal/storage/internal/mem"

type Dumper interface {
	Load(dest *mem.Storage) error
	Dump(source mem.Storage) error
}
