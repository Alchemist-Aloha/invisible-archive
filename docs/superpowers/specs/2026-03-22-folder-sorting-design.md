# Design Spec: Folder-Context Sorting for API and UI

## Overview
Add sorting functionality to the `invisible-archive` project to allow users to sort items by `name`, `name_natural`, `random`, and `size`. The sorting should always prioritize "browsable" items (folders and ZIP archives) by keeping them at the top of the list, regardless of the chosen sort mode (even for `random` and `size`).

## Architecture & Components

### 1. Backend API (Go)
- **Endpoint:** `GET /api/ls`
- **Parameters:**
  - `path` (string): Path to list.
  - `sort` (string): Sort mode (`name`, `natural`, `random`, `size`).
  - `order` (string): Sort direction (`asc`, `desc`).
- **Sorting Logic:**
  - Partition items into two groups:
    - Group A (Browsable): `(Capabilities & CapBrowse) != 0`
    - Group B (Other Files): Everyone else
  - Within each group, apply the requested sort:
    - `name`: Lexical comparison.
    - `natural`: Natural numeric comparison (e.g., "file2" < "file10").
    - `random`: Fisher-Yates shuffle.
    - `size`: Numeric comparison (default `desc` for size).
  - Apply `order` (asc/desc) within each group (except for `random`).

### 2. Frontend (Vue.js)
- **Component:** `Breadcrumbs.vue`
- **State Management:** Update `useNavigation.ts` composable to store `sortMode` and `isDescending`.
- **API Integration:** Update `fetchList` in `api/index.ts` to include the sorting parameters.
- **UI:** Add a dropdown menu or modal in the breadcrumbs area for selecting sort mode and a toggle button for order.

### 3. Mobile (Flutter)
- **Component:** `breadcrumbs_widget.dart`
- **State Management:** Update `ExplorerProvider` to manage `sortMode` and `isDescending`.
- **API Integration:** Update `ApiService.fetchList` to include the sorting parameters.
- **UI:** Add a sort icon button in the breadcrumbs row that opens a menu/sheet to change the sort mode and direction.

## Data Flow
1. User interacts with the sort widget in the UI.
2. The UI updates its local state (sort mode/order) and triggers a new `fetchList` call.
3. The backend receives the request with `sort` and `order` parameters.
4. The backend retrieves the directory listing, partitions it by "browsable-ness," sorts each partition, and returns the combined list.
5. The UI renders the updated list.

## Testing Strategy
- **Backend:** Add unit tests for the sorting logic in `internal/api/handlers_test.go` and `internal/vfs/vfs_test.go` (if applicable).
- **Frontend/Mobile:** Manual verification of the UI components and the list order after changing sort settings.
