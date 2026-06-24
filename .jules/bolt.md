## 2025-06-23 - [Optimization] Efficient Multi-line Text Processing in Go
**Learning:** Iterative line-by-line processing using `strings.Split`, `regexp.ReplaceAllString`, and `strings.Join` on large files (e.g., YAML workflows or logs) creates $O(N)$ allocations. Using a single `regexp.ReplaceAllString` with the multi-line flag `(?m)` on the entire string reduces allocations to $O(1)$ and significantly improves performance.
**Action:** Prefer single-pass regex with `(?m)` for text normalization and cleaning tasks involving line-end patterns.

## 2025-06-23 - [Pattern] Go Benchmarking with Single Files
**Learning:** Running `go test -bench` on individual files (e.g., `go test -bench . file_test.go`) often fails in Go if the tests depend on other files in the same package (e.g., shared types or variables).
**Action:** Use `go test -bench . ./pkg/path/to/package/` to ensure all package dependencies are included in the benchmark run.
