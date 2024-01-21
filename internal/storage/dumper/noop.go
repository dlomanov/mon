package dumper

func init() {
	var _ Dumper = (*NoopDumper)(nil)
}

func NewNoopDumper() *NoopDumper {
	return &NoopDumper{}
}

type NoopDumper struct {
}

func (n NoopDumper) Load(_ *map[string]string) error {
	return nil
}

func (n NoopDumper) Dump(_ map[string]string) error {
	return nil
}
