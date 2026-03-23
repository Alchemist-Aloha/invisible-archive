# Folder-Context Sorting Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add sorting functionality to the server and UI (Vue and Flutter) with a "Folders First" priority.

**Architecture:**
- **VFS handles sorting logic to ensure consistency and avoid API-level dependency on vips for testing.**
- Frontend and Mobile components (Breadcrumbs) allow users to toggle sort mode and order.
- "Browsable" items (folders and ZIPs) are always grouped at the top.

**Tech Stack:** Go (Backend), Vue 3 + TypeScript (Frontend), Flutter (Mobile).

---

## Chunk 1: Backend Sorting Logic (in VFS)

### Task 1: Implement Natural Sort Utility in VFS

**Files:**
- Create: `internal/vfs/sort_util.go`

- [ ] **Step 1: Create `sort_util.go` with `NaturalCompare` function**

```go
package vfs

import (
	"strconv"
	"unicode"
)

func NaturalCompare(s1, s2 string) bool {
	i, j := 0, 0
	for i < len(s1) && j < len(s2) {
		c1, c2 := rune(s1[i]), rune(s2[j])
		if unicode.IsDigit(c1) && unicode.IsDigit(c2) {
			// Find numeric segments
			n1Start := i
			for i < len(s1) && unicode.IsDigit(rune(s1[i])) {
				i++
			}
			n1, _ := strconv.Atoi(s1[n1Start:i])

			n2Start := j
			for j < len(s2) && unicode.IsDigit(rune(s2[j])) {
				j++
			}
			n2, _ := strconv.Atoi(s2[n2Start:j])

			if n1 != n2 {
				return n1 < n2
			}
			continue
		}
		if c1 != c2 {
			return c1 < c2
		}
		i++
		j++
	}
	return len(s1) < len(s2)
}
```

- [ ] **Step 2: Commit utility**

```bash
git add internal/vfs/sort_util.go
git commit -m "feat(vfs): add natural sort utility"
```

### Task 2: Update Manager.ReadDir to support sorting

**Files:**
- Modify: `internal/vfs/vfs.go`

- [ ] **Step 1: Update `ReadDir` signature to handle `sort` and `desc`**

```go
// In internal/vfs/vfs.go
func (m *Manager) ReadDir(path string, sortMode string, desc bool) ([]os.FileInfo, string, error) {
    // ...
}
```

- [ ] **Step 2: Implement sorting in `ReadDir`**

```go
// Add sorting logic inside ReadDir before returning
// Note: need to import "sort" and "math/rand/v2" or similar
```

- [ ] **Step 3: Update `Handler.List` in `internal/api/handlers.go`**

```go
// Update call to m.vfs.ReadDir(requestPath, sortParam, descParam)
```

- [ ] **Step 4: Add unit tests for sorting in VFS**

Modify `internal/vfs/vfs_test.go` or create `internal/vfs/sort_test.go`.

- [ ] **Step 5: Commit backend changes**

```bash
git add internal/vfs/vfs.go internal/vfs/sort_util.go internal/api/handlers.go
git commit -m "feat(backend): implement folder-context sorting in VFS and API"
```

---

## Chunk 2: Vue.js Frontend Updates

### Task 3: Update API and State

**Files:**
- Modify: `frontend/src/api/index.ts`
- Modify: `frontend/src/composables/useNavigation.ts`

- [ ] **Step 1: Update `fetchList` signature**
- [ ] **Step 2: Update `useNavigation.ts` to include sorting state**
- [ ] **Step 3: Commit frontend core changes**

### Task 4: Add Sorting UI to Breadcrumbs

**Files:**
- Modify: `frontend/src/components/explorer/Breadcrumbs.vue`

- [ ] **Step 1: Add sort dropdown/toggle UI**
- [ ] **Step 2: Commit UI changes**

---

## Chunk 3: Flutter Mobile Updates

### Task 5: Update API and Provider

**Files:**
- Modify: `mobile/lib/api.dart`
- Modify: `mobile/lib/providers/explorer_provider.dart`

- [ ] **Step 1: Update `fetchList` in `ApiService`**
- [ ] **Step 2: Update `ExplorerProvider` state**
- [ ] **Step 3: Commit mobile core changes**

### Task 6: Add Sorting UI to Breadcrumbs

**Files:**
- Modify: `mobile/lib/widgets/breadcrumbs_widget.dart`

- [ ] **Step 1: Add sort action button**
- [ ] **Step 2: Commit UI changes**
