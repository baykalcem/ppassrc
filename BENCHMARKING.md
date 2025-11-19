# Benchmarking notes

Use the existing benchmarks under `tests/bench_extended_test.go` just as before:

```bash
go test ./tests -run=^$ -bench=. > bench.log
```

To visualize the results run the new plotting helper. It accepts a benchmark log either via `-in` or stdin and emits one `benchplot_<group>.svg` chart per benchmark group:

```bash
go run ./cmd/benchplot -in bench.log -out-dir bench-plots
```

If you omit `-in` you can pipe the benchmark output directly:

```bash
go test ./tests -run=^$ -bench=. | go run ./cmd/benchplot -out-dir bench-plots
```

The default canvas size is `1200×640` but you can adjust it with `-width`/`-height`. Each chart highlights the ns/op values collected for its group and labels the axes with the benchmark names.
