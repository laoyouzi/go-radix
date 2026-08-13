# go-radix — mutable radix tree with ordered iteration

A drop-in enhancement of [armon/go-radix](https://github.com/armon/go-radix): **all original APIs unchanged**, plus `Iterator` with **`SeekLowerBound`** for efficient ordered scans.

[中文](./README.zh.md)

---

## Problem

[armon/go-radix](https://github.com/armon/go-radix) is a widely used **mutable** radix tree: great for point lookups and one-shot `Walk`.

Many workloads instead repeat:

> **First key `>= seek`**, then **read forward in order**.

Examples: paginated cursors, range scans, streaming export, monotonic offsets, ordered rule tables — not point lookups alone.

The stock library has **no lower-bound seek**. Callers simulate it with a full-tree `Walk` every time:

| Symptom | Cause |
|---------|--------|
| Slow single “`>= seek`” query | O(n) scan from root |
| Slow sequential reads | n steps × O(n) per step ≈ **O(n²)** |
| Feature gap vs immutable radix | [go-immutable-radix](https://github.com/hashicorp/go-immutable-radix) has `Iterator.SeekLowerBound`; mutable tree did not |
| `WalkPrefix` limitations | Follows radix **edges**, not always identical to a **byte-wise key prefix range** |

## What this repo adds

Ports the iterator algorithm from [hashicorp/go-immutable-radix](https://github.com/hashicorp/go-immutable-radix) (see `iter.go`) to the **mutable** tree:

```go
it := tree.Iterator()
it.SeekLowerBound(seekKey) // O(len(key)) — first key >= seekKey
k, v, ok := it.Next()      // amortized O(1) forward step
```

| API | Purpose |
|-----|---------|
| `(*Tree).Iterator()` | Create an iterator |
| `(*Iterator).SeekLowerBound(key)` | Seek to smallest `key' >= key` |
| `(*Iterator).SeekPrefix(prefix)` | Seek to first key under prefix |
| `(*Iterator).Next()` | Next key / value in order |

## Advantages

| Aspect | Stock go-radix | This repo |
|--------|----------------|-----------|
| Single lower-bound | O(n) `Walk` | O(len(seek)) descent |
| Sequential pass over n keys | O(n²) | **O(n)** |
| Extra memory | None | Iterator stack only (no duplicate key index) |
| Compatibility | — | **Additive API, no breaking changes** |
| In-place mutation | Yes | Yes (vs immutable CoW — lower alloc on bulk load) |

## Quick start

```bash
git clone <this-repo>
cd go-radix
go test ./...
```

```go
import "github.com/armon/go-radix"

r := radix.New()
r.Insert("app:1001", "a")
r.Insert("app:1002", "b")
r.Insert("app:2000", "c")

it := r.Iterator()
it.SeekLowerBound("app:1001")
for {
    k, v, ok := it.Next()
    if !ok {
        break
    }
    fmt.Println(k, v)
}
```

## Upstream

- Tree: [armon/go-radix](https://github.com/armon/go-radix) (MIT)
- Added: `iter.go`, `iter_test.go`
- PRs back to upstream welcome; this repo can also be used as a maintained fork

## Tests

```bash
go test ./...
```

- Table-driven: mixed-length, prefix, and binary keys
- `testing/quick` fuzz: iterator output matches sorted-slice filter

## License

- Tree: [armon/go-radix](https://github.com/armon/go-radix) — **MIT**
- Iterator algorithm: adapted from [hashicorp/go-immutable-radix](https://github.com/hashicorp/go-immutable-radix) — **MPL-2.0** (see `iter.go`)
