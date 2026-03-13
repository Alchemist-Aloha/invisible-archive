## 2024-03-13 - Path parsing in readZipDir
**Learning:** `strings.Split` causes slice allocations inside inner loops for path processing in `readZipDir`. This causes slower performance for large ZIP directories, despite paths logically being treated as single string components divided by slashes.
**Action:** Use `strings.IndexByte` instead of `strings.Split` when peeling directory paths and names in the archive virtual file system, returning slice substrings to prevent heap allocations.
