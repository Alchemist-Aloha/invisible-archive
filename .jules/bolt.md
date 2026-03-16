## 2024-05-20 - Zero-Allocation String Processing in tight VFS loops
**Learning:** Parsing large ZIP directory structures (`readZipDir` in `vfs.go`) can be significantly slowed down by unnecessary slice allocations using `strings.Split` just to extract the first directory component.
**Action:** Always prefer `strings.IndexByte` or slicing over `strings.Split` in tight VFS path traversal loops. This codebase relies heavily on the VFS layer for fast navigation, making string allocation overhead a critical path to optimize.

## 2024-05-23 - [Zero-Allocation Path Parsing in ZIP Indexer]
**Learning:** The `IndexZip` loop processes thousands of entries in a single operation. Utilizing allocation-heavy functions like `strings.Split` and `strings.Join` inside this tight loop causes measurable performance bottlenecks. Since ZIP internal paths are reliably forward-slash separated, we can achieve O(1) allocation overhead by replacing string array manipulations with simple string slicing using `strings.LastIndexByte`.
**Action:** When extracting parents and filenames from absolute paths in performance-critical areas, prioritize zero-allocation slice slicing (e.g., `strings.LastIndexByte`) over allocation-generating `Split`/`Join` sequences.

## 2024-05-28 - Zero-Allocation Path Parsing in ZIP Path Generation
**Learning:** `filepath.Join` requires OS-specific path parsing and causes multiple allocations behind the scenes. Using it in tight loops like `IndexZip` which traverses thousands of files is a noticeable performance bottleneck (~2600ns per loop iteration).
**Action:** Since we know ZIPs only contain forward slashes and we can guarantee `relZipPath` doesn't end in a slash, use simple string concatenation to construct paths instead. Doing so avoids path parsing overhead and reduces execution time by roughly ~4x (~670ns).
