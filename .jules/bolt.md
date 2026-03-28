## 2024-05-20 - Zero-Allocation String Processing in tight VFS loops
**Learning:** Parsing large ZIP directory structures (`readZipDir` in `vfs.go`) can be significantly slowed down by unnecessary slice allocations using `strings.Split` just to extract the first directory component.
**Action:** Always prefer `strings.IndexByte` or slicing over `strings.Split` in tight VFS path traversal loops. This codebase relies heavily on the VFS layer for fast navigation, making string allocation overhead a critical path to optimize.

## 2024-05-23 - [Zero-Allocation Path Parsing in ZIP Indexer]
**Learning:** The `IndexZip` loop processes thousands of entries in a single operation. Utilizing allocation-heavy functions like `strings.Split` and `strings.Join` inside this tight loop causes measurable performance bottlenecks. Since ZIP internal paths are reliably forward-slash separated, we can achieve O(1) allocation overhead by replacing string array manipulations with simple string slicing using `strings.LastIndexByte`.
**Action:** When extracting parents and filenames from absolute paths in performance-critical areas, prioritize zero-allocation slice slicing (e.g., `strings.LastIndexByte`) over allocation-generating `Split`/`Join` sequences.

## 2024-05-23 - [Zero-Allocation Path Encoding in Frontend/Mobile]
**Learning:** Using `split('/').map(encodeURIComponent).join('/')` (or equivalent in Dart) creates unnecessary string arrays and allocations just to encode URL paths. In tight loops or frequent API calls, this can cause measurable performance bottlenecks due to Garbage Collection (GC) overhead.
**Action:** Optimize path encoding by avoiding array allocations entirely. In JavaScript/TypeScript, use `encodeURIComponent(path).replaceAll('%2F', '/')`. In Dart, use `Uri.encodeComponent(path).replaceAll('%2F', '/')`. Both approaches significantly reduce allocations and are measurably faster.
