# go-radix

Mutable radix tree with **`Iterator`** and **`SeekLowerBound`** — a drop-in enhancement of [armon/go-radix](https://github.com/armon/go-radix).

## Documentation

| Language | File |
|----------|------|
| English | [README.en.md](./README.en.md) |
| 中文 | [README.zh.md](./README.zh.md) |

## Quick example

```go
it := tree.Iterator()
it.SeekLowerBound(seekKey)
k, v, ok := it.Next()
```

```bash
go test ./...
```
