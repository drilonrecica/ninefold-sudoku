# Puzzle engine benchmark baseline

Run from `apps/server`:

```sh
go test -run '^$' -bench 'Benchmark(Validation|Uniqueness|LogicalGrading|Hints|Canonicalization|Generation)$' -benchmem ./internal/puzzle/engine
```

Validation, uniqueness, logical grading, and hint selection are bounded gameplay-support operations.
Canonicalization and generation are offline-only and must never be called from an HTTP, WebSocket,
Room actor, or gameplay request path.

Initial regression limits on the pinned CI runner:

| Operation | Limit |
|---|---:|
| Validation | 50 µs |
| Uniqueness | 20 ms |
| Logical grading | 20 ms |
| Nudge | 5 ms |
| Canonicalization | 5 s |
| Easy generation | 5 s |

The limits are deliberately broad cross-machine guards. Tightening them requires representative CI
history; increasing them requires an investigated benchmark report.

## Initial local baseline

Recorded 2026-07-24 on Linux/amd64, Go 1.26.5, Intel i5-7500:

| Operation | Time | Allocations |
|---|---:|---:|
| Validation | 201 ns | 0 |
| Uniqueness | 161 µs | 0 |
| Logical grading | 2.10 ms | 33,884 |
| Nudge | 3.24 µs | 3 |
| Canonicalization | 1.74 s | 2 |
| Easy generation | 775 µs | 1,200 |

Canonicalization and generation were measured with `-benchtime=1x`; the remaining operations used
the Go benchmark harness default duration.
