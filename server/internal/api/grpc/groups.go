package grpcapi

import (
    "context"
    "errors"
    
    msgv1 "dev.c0rex64.heroin/internal/gen/shared/proto/messaging/v1"
    "dev.c0rex64.heroin/internal/groups"
)

// группы интерфейс
type GroupService interface {
    CreateGroup(ctx context.Context, creatorID, name string) (*groups.Group, error)
    AddMember(ctx context.Context, groupID, userID, adderID string, userPubKey []byte) error
    RemoveMember(ctx context.Context, groupID, userID, removerID string) error
    GetUserGroups(ctx context.Context, userID string) ([]*groups.Group, error)
    GetGroupMembers(ctx context.Context, groupID, requesterID string) ([]*groups.Member, error)
}

// создать группу
func (s *Server) CreateGroup(ctx context.Context, req *msgv1.CreateGroupRequest) (*msgv1.CreateGroupResponse, error) {
    if s.GroupSvc == nil {
        return nil, errors.New("groups not configured")
    }
    
    // получаем user id из контекста (должен быть после auth middleware)
    userID := getUserIDFromContext(ctx)
    if userID == "" {
        return nil, errors.New("unauthorized")
    }
    
    // создаем группу
    group, err := s.GroupSvc.CreateGroup(ctx, userID, req.Name)
    if err != nil {
        return nil, err
    }
    
    return &msgv1.CreateGroupResponse{
        GroupId:      group.ID,
        EncryptedKey: group.GroupKey, // тут должно быть шифрование
    }, nil
}

// добавить участника
func (s *Server) AddGroupMember(ctx context.Context, req *msgv1.AddGroupMemberRequest) (*msgv1.AddGroupMemberResponse, error) {
    if s.GroupSvc == nil {
        return nil, errors.New("groups not configured")
    }
    
    userID := getUserIDFromContext(ctx)
    if userID == "" {
        return nil, errors.New("unauthorized")
    }
    
    // получаем публичный ключ добавляемого юзера
    pubKey, err := s.AuthSvc.GetPublicKey(ctx, req.UserId)
    if err != nil {
        return nil, err
    }
    
    err = s.GroupSvc.AddMember(ctx, req.GroupId, req.UserId, userID, pubKey)
    if err != nil {
        return nil, err
    }
    
    return &msgv1.AddGroupMemberResponse{Success: true}, nil
}

// удалить участника
func (s *Server) RemoveGroupMember(ctx context.Context, req *msgv1.RemoveGroupMemberRequest) (*msgv1.RemoveGroupMemberResponse, error) {
    if s.GroupSvc == nil {
        return nil, errors.New("groups not configured")
    }
    
    userID := getUserIDFromContext(ctx)
    if userID == "" {
        return nil, errors.New("unauthorized")
    }
    
    err := s.GroupSvc.RemoveMember(ctx, req.GroupId, req.UserId, userID)
    if err != nil {
        return nil, err
    }
    
    return &msgv1.RemoveGroupMemberResponse{Success: true}, nil
}

// получить группы юзера
func (s *Server) GetGroups(ctx context.Context, req *msgv1.GetGroupsRequest) (*msgv1.GetGroupsResponse, error) {
    if s.GroupSvc == nil {
        return nil, errors.New("groups not configured")
    }
    
    userID := getUserIDFromContext(ctx)
    if userID == "" {
        return nil, errors.New("unauthorized")
    }
    
    groups, err := s.GroupSvc.GetUserGroups(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // конвертируем в proto
    resp := &msgv1.GetGroupsResponse{
        Groups: make([]*msgv1.Group, 0, len(groups)),
    }
    
    for _, g := range groups {
        resp.Groups = append(resp.Groups, &msgv1.Group{
            Id:            g.ID,
            Name:          g.Name,
            CreatorId:     g.CreatorID,
            CreatedAtUnix: g.CreatedAt.Unix(),
            MemberCount:   int32(g.MemberCount),
            EncryptedKey:  g.GroupKey, // уже зашифрован для юзера
            KeyVersion:    int32(g.KeyVersion),
        })
    }
    
    return resp, nil
}

// получить участников группы
func (s *Server) GetGroupMembers(ctx context.Context, req *msgv1.GetGroupMembersRequest) (*msgv1.GetGroupMembersResponse, error) {
    if s.GroupSvc == nil {
        return nil, errors.New("groups not configured")
    }
    
    userID := getUserIDFromContext(ctx)
    if userID == "" {
        return nil, errors.New("unauthorized")
    }
    
    members, err := s.GroupSvc.GetGroupMembers(ctx, req.GroupId, userID)
    if err != nil {
        return nil, err
    }
    
    // конвертируем в proto
    resp := &msgv1.GetGroupMembersResponse{
        Members: make([]*msgv1.GroupMember, 0, len(members)),
    }
    
    for _, m := range members {
        resp.Members = append(resp.Members, &msgv1.GroupMember{
            UserId:       m.UserID,
            Role:         m.Role,
            JoinedAtUnix: m.JoinedAt.Unix(),
        })
    }
    
    return resp, nil
}

// хелпер для получения user id из контекста
func getUserIDFromContext(ctx context.Context) string {
    // тут должна быть реальная логика извлечения из jwt/контекста
    // пока заглушка
    return "user123"
}
