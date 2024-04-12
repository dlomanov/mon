
Профайлы сделаны на основе `pg_benchmark_test.go`

Команды:

- `go tool pprof -http=":9090" base.test base.pprof`

- `go tool pprof -http=":9090" result.test result.pprof`

- `go tool pprof -top -diff_base="base.pprof" result.pprof`