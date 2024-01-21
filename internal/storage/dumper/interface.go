package dumper

type Dumper interface {
	Load(dest *map[string]string) error
	Dump(source map[string]string) error
}
