## 2026-06-21 - [Optimize string normalization]
**Learning:** Functions that process multi-line strings using `strings.Split` and `strings.Join` can incur significant allocation overhead, especially for large inputs. Replacing them with `strings.Builder` and manual iteration using `strings.IndexByte` significantly reduces memory allocations and improves performance.
**Action:** Prefer `strings.Builder` and manual iteration for high-frequency string processing tasks in the core utility packages.
