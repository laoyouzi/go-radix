# go-radix — 带有序迭代的可变 Radix 树

基于 [armon/go-radix](https://github.com/armon/go-radix)的增强：**完全兼容原有 API**，新增带 **`SeekLowerBound`** 的 `Iterator`，适用于大范围、连续的字典序遍历。

[English](./README.en.md)

---

## 解决什么问题

[armon/go-radix](https://github.com/armon/go-radix) 是常用的**可变**有序 Radix 树，适合 `Insert` / `Delete` / `Get` 以及一次性全量 `Walk`。

但在很多业务里，访问模式不是「查几个点」，而是反复问：

> **第一个 `key >= seek` 是谁？** 然后 **按顺序一直往后读**。

典型场景包括：分页 cursor、范围扫描、流式导出、按 key 递增的消费位点、路由/规则表的有序匹配等。

原版库没有 **lower-bound seek**，只能每次从根节点 `Walk` 整棵树，再比较 key。结果是：

| 现象 | 原因 |
|------|------|
| 单次「找 `>= seek`」慢 | 每次 O(n) 全树扫描 |
| 连续顺序读更慢 | n 步 × 每步 O(n) ≈ **O(n²)** |
| 与 immutable 版不对等 | [go-immutable-radix](https://github.com/hashicorp/go-immutable-radix) 已有 `Iterator.SeekLowerBound`，mutable 版长期缺失 |
| `WalkPrefix` 不够用 | 沿 Radix **边** 走子树，与「按完整 key 的字节前缀区间过滤」语义并不总一致 |

## 本仓库做了什么

移植 [hashicorp/go-immutable-radix](https://github.com/hashicorp/go-immutable-radix) 的迭代器算法（见 `iter.go` 注释），为 **mutable** 树补上同一能力：

```go
it := tree.Iterator()
it.SeekLowerBound(seekKey) // O(len(key)) 定位到第一个 key >= seekKey
k, v, ok := it.Next()      // 顺序下一个，摊销 O(1)
```

| API | 说明 |
|-----|------|
| `(*Tree).Iterator()` | 创建迭代器 |
| `(*Iterator).SeekLowerBound(key)` | 定位到最小 `key' >= key` |
| `(*Iterator).SeekPrefix(prefix)` | 定位到 prefix 下第一项 |
| `(*Iterator).Next()` | 返回下一个 key / value |

## 优势

| 维度 | 原版 go-radix | 本仓库 |
|------|---------------|--------|
| 单次 lower-bound | O(n) `Walk` | O(len(seek)) 树上下探 |
| 连续顺序遍历 n 项 | O(n²) | **O(n)** |
| 额外内存 | 无 | 迭代栈（无 duplicate key 索引） |
| API 兼容 | — | **纯新增，无 breaking change** |
| 可变树原地更新 | 支持 | 支持（相对 immutable 版的 CoW 更省导入内存） |

## 快速开始

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

## 与上游的关系

- 核心树实现来自 [armon/go-radix](https://github.com/armon/go-radix)（MIT）
- 新增 `iter.go`、`iter_test.go`
- 欢迎向上游提交 PR；也可直接使用本仓库作为 fork / 长期维护分支

## 测试

```bash
go test ./...
```

- 表驱动：混合长度 key、前缀 key、二进制 key
- `testing/quick` fuzz：迭代结果与排序 slice 过滤一致

## 许可

- 树实现：[armon/go-radix](https://github.com/armon/go-radix) — **MIT**
- Iterator 算法：改编自 [hashicorp/go-immutable-radix](https://github.com/hashicorp/go-immutable-radix) — **MPL-2.0**（见 `iter.go`）
