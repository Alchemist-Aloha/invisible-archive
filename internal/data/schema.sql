CREATE TABLE items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_path TEXT NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    is_dir BOOLEAN NOT NULL,
    size INTEGER NOT NULL,
    mod_time INTEGER NOT NULL,
    capabilities INTEGER NOT NULL,
    is_inside_zip BOOLEAN NOT NULL,
    indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_parent_path ON items(parent_path);

-- FTS5 Virtual Table for searching
CREATE VIRTUAL TABLE items_fts USING fts5(
    name,
    path,
    content='items',
    content_rowid='id'
);

-- Triggers to keep FTS in sync
CREATE TRIGGER items_ai AFTER INSERT ON items BEGIN
  INSERT INTO items_fts(rowid, name, path) VALUES (new.id, new.name, new.path);
END;

CREATE TRIGGER items_ad AFTER DELETE ON items BEGIN
  INSERT INTO items_fts(items_fts, rowid, name, path) VALUES('delete', old.id, old.name, old.path);
END;

CREATE TRIGGER items_au AFTER UPDATE ON items BEGIN
  INSERT INTO items_fts(items_fts, rowid, name, path) VALUES('delete', old.id, old.name, old.path);
  INSERT INTO items_fts(rowid, name, path) VALUES (new.id, new.name, new.path);
END;
