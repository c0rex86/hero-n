-- группы
CREATE TABLE IF NOT EXISTS groups (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    creator_id TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    group_key BLOB NOT NULL, -- симметричный ключ группы
    key_version INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY(creator_id) REFERENCES users(id) ON DELETE CASCADE
);

-- участники групп
CREATE TABLE IF NOT EXISTS group_members (
    group_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    joined_at INTEGER NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('admin', 'member')),
    encrypted_key BLOB NOT NULL, -- групповой ключ зашифрованный для участника
    key_version INTEGER NOT NULL,
    PRIMARY KEY(group_id, user_id),
    FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- индексы
CREATE INDEX IF NOT EXISTS idx_group_members_user ON group_members(user_id);
CREATE INDEX IF NOT EXISTS idx_groups_creator ON groups(creator_id);
CREATE INDEX IF NOT EXISTS idx_groups_created ON groups(created_at DESC);
