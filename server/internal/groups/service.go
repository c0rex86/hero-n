package groups

import (
    "context"
    "crypto/rand"
    "database/sql"
    "errors"
    "time"
    
    "github.com/google/uuid"
)

// группа
type Group struct {
    ID          string
    Name        string
    CreatorID   string
    CreatedAt   time.Time
    GroupKey    []byte // симметричный ключ группы
    KeyVersion  int
    MemberCount int
}

// участник группы
type Member struct {
    GroupID       string
    UserID        string
    JoinedAt      time.Time
    Role          string // admin, member
    EncryptedKey  []byte // групповой ключ зашифрованный публичным ключом участника
    KeyVersion    int
}

// сервис групп
type Service struct {
    db *sql.DB
}

func NewService(db *sql.DB) *Service {
    return &Service{db: db}
}

// создать группу
func (s *Service) CreateGroup(ctx context.Context, creatorID, name string) (*Group, error) {
    // генерим id и ключ
    groupID := uuid.NewString()
    groupKey := make([]byte, 32)
    if _, err := rand.Read(groupKey); err != nil {
        return nil, err
    }
    
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()
    
    // создаем группу
    _, err = tx.ExecContext(ctx, `
        INSERT INTO groups (id, name, creator_id, created_at, group_key, key_version)
        VALUES (?, ?, ?, ?, ?, ?)
    `, groupID, name, creatorID, time.Now().Unix(), groupKey, 1)
    if err != nil {
        return nil, err
    }
    
    // добавляем создателя как админа
    _, err = tx.ExecContext(ctx, `
        INSERT INTO group_members (group_id, user_id, joined_at, role, encrypted_key, key_version)
        VALUES (?, ?, ?, ?, ?, ?)
    `, groupID, creatorID, time.Now().Unix(), "admin", groupKey, 1) // тут должно быть шифрование ключа
    if err != nil {
        return nil, err
    }
    
    if err = tx.Commit(); err != nil {
        return nil, err
    }
    
    return &Group{
        ID:         groupID,
        Name:       name,
        CreatorID:  creatorID,
        CreatedAt:  time.Now(),
        GroupKey:   groupKey,
        KeyVersion: 1,
        MemberCount: 1,
    }, nil
}

// добавить участника
func (s *Service) AddMember(ctx context.Context, groupID, userID, adderID string, userPubKey []byte) error {
    // проверяем права добавляющего
    var role string
    err := s.db.QueryRowContext(ctx, `
        SELECT role FROM group_members WHERE group_id = ? AND user_id = ?
    `, groupID, adderID).Scan(&role)
    if err != nil {
        return errors.New("not a member or group not found")
    }
    if role != "admin" {
        return errors.New("only admins can add members")
    }
    
    // получаем ключ группы
    var groupKey []byte
    var keyVersion int
    err = s.db.QueryRowContext(ctx, `
        SELECT group_key, key_version FROM groups WHERE id = ?
    `, groupID).Scan(&groupKey, &keyVersion)
    if err != nil {
        return err
    }
    
    // тут должно быть шифрование groupKey публичным ключом userPubKey
    // пока просто копируем
    encryptedKey := make([]byte, len(groupKey))
    copy(encryptedKey, groupKey)
    
    // добавляем участника
    _, err = s.db.ExecContext(ctx, `
        INSERT INTO group_members (group_id, user_id, joined_at, role, encrypted_key, key_version)
        VALUES (?, ?, ?, ?, ?, ?)
    `, groupID, userID, time.Now().Unix(), "member", encryptedKey, keyVersion)
    
    return err
}

// удалить участника
func (s *Service) RemoveMember(ctx context.Context, groupID, userID, removerID string) error {
    // проверяем права
    var role string
    err := s.db.QueryRowContext(ctx, `
        SELECT role FROM group_members WHERE group_id = ? AND user_id = ?
    `, groupID, removerID).Scan(&role)
    if err != nil {
        return errors.New("not a member")
    }
    
    // только админ может удалять или сам участник может выйти
    if role != "admin" && removerID != userID {
        return errors.New("no permission")
    }
    
    _, err = s.db.ExecContext(ctx, `
        DELETE FROM group_members WHERE group_id = ? AND user_id = ?
    `, groupID, userID)
    
    // после удаления нужно ротировать ключ группы
    // пока пропускаем
    
    return err
}

// получить группы пользователя
func (s *Service) GetUserGroups(ctx context.Context, userID string) ([]*Group, error) {
    rows, err := s.db.QueryContext(ctx, `
        SELECT g.id, g.name, g.creator_id, g.created_at, gm.encrypted_key, g.key_version,
               (SELECT COUNT(*) FROM group_members WHERE group_id = g.id) as member_count
        FROM groups g
        JOIN group_members gm ON g.id = gm.group_id
        WHERE gm.user_id = ?
        ORDER BY g.created_at DESC
    `, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var groups []*Group
    for rows.Next() {
        var g Group
        var createdAt int64
        err := rows.Scan(&g.ID, &g.Name, &g.CreatorID, &createdAt, &g.GroupKey, &g.KeyVersion, &g.MemberCount)
        if err != nil {
            return nil, err
        }
        g.CreatedAt = time.Unix(createdAt, 0)
        groups = append(groups, &g)
    }
    
    return groups, nil
}

// получить участников группы
func (s *Service) GetGroupMembers(ctx context.Context, groupID, requesterID string) ([]*Member, error) {
    // проверяем что запрашивающий - участник
    var exists bool
    err := s.db.QueryRowContext(ctx, `
        SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?)
    `, groupID, requesterID).Scan(&exists)
    if err != nil || !exists {
        return nil, errors.New("not a member")
    }
    
    rows, err := s.db.QueryContext(ctx, `
        SELECT user_id, joined_at, role, key_version
        FROM group_members
        WHERE group_id = ?
        ORDER BY joined_at
    `, groupID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var members []*Member
    for rows.Next() {
        var m Member
        var joinedAt int64
        m.GroupID = groupID
        err := rows.Scan(&m.UserID, &joinedAt, &m.Role, &m.KeyVersion)
        if err != nil {
            return nil, err
        }
        m.JoinedAt = time.Unix(joinedAt, 0)
        members = append(members, &m)
    }
    
    return members, nil
}

// ротировать ключ группы
func (s *Service) RotateGroupKey(ctx context.Context, groupID, initiatorID string) error {
    // только админ может ротировать
    var role string
    err := s.db.QueryRowContext(ctx, `
        SELECT role FROM group_members WHERE group_id = ? AND user_id = ?
    `, groupID, initiatorID).Scan(&role)
    if err != nil || role != "admin" {
        return errors.New("only admins can rotate keys")
    }
    
    // генерим новый ключ
    newKey := make([]byte, 32)
    if _, err := rand.Read(newKey); err != nil {
        return err
    }
    
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // обновляем ключ группы
    _, err = tx.ExecContext(ctx, `
        UPDATE groups SET group_key = ?, key_version = key_version + 1
        WHERE id = ?
    `, newKey, groupID)
    if err != nil {
        return err
    }
    
    // тут нужно перешифровать ключ для всех участников
    // пока пропускаем
    
    return tx.Commit()
}
