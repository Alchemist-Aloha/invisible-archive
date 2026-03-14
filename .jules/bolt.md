## 2024-05-20 - Zero-Allocation String Processing in tight VFS loops
**Learning:** Parsing large ZIP directory structures (`readZipDir` in `vfs.go`) can be significantly slowed down by unnecessary slice allocations using `strings.Split` just to extract the first directory component.
**Action:** Always prefer `strings.IndexByte` or slicing over `strings.Split` in tight VFS path traversal loops. This codebase relies heavily on the VFS layer for fast navigation, making string allocation overhead a critical path to optimize.

## 2026-03-14 - [Optimize Path Parsing in ZIP Indexing Loop]
**Learning:** `strings.Split` combined with `strings.Join` inside a tight loop (like indexing thousands of ZIP file entries) causes a major performance bottleneck due to unnecessary string allocations and array creations.
**Action:** Replace `strings.Split` and `strings.Join` with `strings.LastIndexByte` for zero-allocation path parsing in indexing loops. This yielded an ~7x performance gain in benchmarks.
