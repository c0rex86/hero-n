package grpcapi

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	authv1 "dev.c0rex64.heroin/internal/gen/shared/proto/auth/v1"
	msgv1 "dev.c0rex64.heroin/internal/gen/shared/proto/messaging/v1"
	stgv1 "dev.c0rex64.heroin/internal/gen/shared/proto/storage/v1"
	"google.golang.org/grpc"
)

type Server struct {
	gs  *grpc.Server
	lis net.Listener

	AuthSvc      AuthService
	MessagingSvc MessagingService
	StorageSvc   StorageService

	authv1.UnimplementedAuthServiceServer
	msgv1.UnimplementedMessagingServiceServer
	stgv1.UnimplementedStorageServiceServer
}

type AuthService interface {
	Register(ctx context.Context, username string, passwordProof []byte, clientPub []byte) (string, error)
	Login(ctx context.Context, username string, passwordProof []byte, deviceID string, secondCode string, now time.Time) (string, string, time.Time, error)
	Refresh(ctx context.Context, refreshToken string, deviceID string, now time.Time) (string, time.Time, error)
}

type MessagingService interface {
	Send(ctx context.Context, envelope []byte) error
	Pull(ctx context.Context, conversationID string, since int64) ([][]byte, error)
}

type StorageService interface {
	PutCAR(ctx context.Context, name, mime string, size int64, carChunks [][]byte, totalBlake3 []byte) (fileID string, cid string, err error)
	GetCAR(ctx context.Context, cid string) ([][]byte, error)
}

func New(addr string) (*Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil { return nil, fmt.Errorf("listen: %w", err) }
	s := &Server{gs: grpc.NewServer(), lis: lis}
	authv1.RegisterAuthServiceServer(s.gs, s)
	msgv1.RegisterMessagingServiceServer(s.gs, s)
	stgv1.RegisterStorageServiceServer(s.gs, s)
	return s, nil
}

func (s *Server) Start() error {
	log.Printf("grpc listening on %s", s.lis.Addr())
	return s.gs.Serve(s.lis)
}

// Auth

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	id, err := s.AuthSvc.Register(ctx, req.Username, req.PasswordProof, req.ClientPubkey)
	if err != nil { return nil, err }
	return &authv1.RegisterResponse{UserId: id}, nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	now := time.Now()
	access, refresh, exp, err := s.AuthSvc.Login(ctx, req.Username, req.PasswordProof, req.DeviceId, req.SecondaryCode, now)
	if err != nil { return nil, err }
	return &authv1.LoginResponse{AccessToken: access, RefreshToken: refresh, ExpiresAtUnix: exp.Unix()}, nil
}

func (s *Server) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	now := time.Now()
	tok, exp, err := s.AuthSvc.Refresh(ctx, req.RefreshToken, req.DeviceId, now)
	if err != nil { return nil, err }
	return &authv1.RefreshResponse{AccessToken: tok, ExpiresAtUnix: exp.Unix()}, nil
}

// Messaging

func (s *Server) Send(ctx context.Context, req *msgv1.SendRequest) (*msgv1.SendResponse, error) {
	b := []byte{}
	b = append(b, req.Envelope.Ciphertext...)
	if err := s.MessagingSvc.Send(ctx, b); err != nil { return nil, err }
	return &msgv1.SendResponse{Accepted: true}, nil
}

func (s *Server) Pull(ctx context.Context, req *msgv1.PullRequest) (*msgv1.PullResponse, error) {
	envs, err := s.MessagingSvc.Pull(ctx, req.ConversationId, req.SinceUnix)
	if err != nil { return nil, err }
	out := make([]*msgv1.Envelope, 0, len(envs))
	for _, b := range envs {
		out = append(out, &msgv1.Envelope{Ciphertext: b})
	}
	return &msgv1.PullResponse{Envelopes: out}, nil
}

// Storage

func (s *Server) PutFile(stream stgv1.StorageService_PutFileServer) error {
	var chunks [][]byte
	var name, mime string
	var size int64
	var b3 []byte
	for {
		req, err := stream.Recv()
		if err != nil { break }
		name, mime, size = req.Name, req.Mime, req.SizeBytes
		if len(b3) == 0 && len(req.TotalBlake3) > 0 { b3 = append([]byte(nil), req.TotalBlake3...) }
		chunks = append(chunks, req.EncryptedCarChunk)
		if req.LastChunk { break }
	}
	fileID, cid, err := s.StorageSvc.PutCAR(stream.Context(), name, mime, size, chunks, b3)
	if err != nil { return err }
	return stream.SendAndClose(&stgv1.PutFileResponse{Accepted: true, FileId: fileID, Cid: cid})
}

func (s *Server) GetFile(req *stgv1.GetFileRequest, stream stgv1.StorageService_GetFileServer) error {
	chunks, err := s.StorageSvc.GetCAR(stream.Context(), req.Cid)
	if err != nil { return err }
	for i, ch := range chunks {
		if err := stream.Send(&stgv1.GetFileResponse{EncryptedCarChunk: ch, LastChunk: i == len(chunks)-1}); err != nil { return err }
	}
	return nil
}
