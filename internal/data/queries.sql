-- name: UpsertItem :exec
INSERT INTO items (
    parent_path, name, path, is_dir, size, mod_time, capabilities, is_inside_zip
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
) ON CONFLICT(path) DO UPDATE SET
    size = excluded.size,
    mod_time = excluded.mod_time,
    capabilities = excluded.capabilities,
    indexed_at = CURRENT_TIMESTAMP;

-- name: ListItemsByParent :many
SELECT * FROM items
WHERE parent_path = ?
ORDER BY is_dir DESC, name ASC;

-- name: SearchItems :many
SELECT *
FROM items
WHERE name LIKE ? OR path LIKE ?
LIMIT 50;

-- name: DeleteItemsByPathPrefix :exec
DELETE FROM items
WHERE path LIKE ? || '%';

-- name: GetItemByPath :one
SELECT * FROM items
WHERE path = ? LIMIT 1;
