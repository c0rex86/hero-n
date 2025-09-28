package grpcapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	authv1 "dev.c0rex64.heroin/internal/gen/shared/proto/auth/v1"
	msgv1 "dev.c0rex64.heroin/internal/gen/shared/proto/messaging/v1"
	stgv1 "dev.c0rex64.heroin/internal/gen/shared/proto/storage/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"

	"dev.c0rex64.heroin/internal/metrics"
	"dev.c0rex64.heroin/internal/p2p"
	"dev.c0rex64.heroin/internal/relay"
	"dev.c0rex64.heroin/internal/routing"
)

type Server struct {
	gs  *grpc.Server
	lis net.Listener

	AuthSvc      AuthService
	MessagingSvc MessagingService
	StorageSvc   StorageService
	GroupSvc     GroupService
	Collector    *metrics.Collector
	StreamMgr    *p2p.StreamManager
	RelayMgr     *relay.RelayManager
	Router       *routing.AdaptiveRouter

	authv1.UnimplementedAuthServiceServer
	msgv1.UnimplementedMessagingServiceServer
	stgv1.UnimplementedStorageServiceServer
}

type AuthService interface {
	Register(ctx context.Context, username string, passwordProof []byte, clientPub []byte) (string, error)
	Login(ctx context.Context, username string, passwordProof []byte, deviceID string, secondCode string, now time.Time) (string, string, time.Time, error)
	Refresh(ctx context.Context, refreshToken string, deviceID string, now time.Time) (string, time.Time, error)
	GetPublicKey(ctx context.Context, userID string) ([]byte, error)
}

type MessagingService interface {
	Send(ctx context.Context, envelope []byte) error
	Pull(ctx context.Context, conversationID string, since int64) ([][]byte, error)
	PullPage(ctx context.Context, conversationID string, since int64, limit int) ([][]byte, int64, bool, error)
}

type StorageService interface {
	PutCAR(ctx context.Context, name, mime string, size int64, carChunks [][]byte, totalBlake3 []byte) (fileID string, cid string, err error)
	GetCAR(ctx context.Context, cid string) ([][]byte, error)
}

var (
	rateMu sync.Mutex
	rateMap = map[string][]time.Time{}
)

func rateLimitUnaryInterceptor(maxPerMinute int) grpc.UnaryServerInterceptor {
	window := time.Minute
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		p, _ := peer.FromContext(ctx)
		key := "unknown"
		if p != nil && p.Addr != nil { key = p.Addr.String() }
		if isAuthMethod(info.FullMethod) {
			rateMu.Lock()
			t := time.Now()
			v := rateMap[key]
			var vv []time.Time
			for _, ts := range v { if t.Sub(ts) < window { vv = append(vv, ts) } }
			if len(vv) >= maxPerMinute {
				rateMu.Unlock()
				return nil, status.Error(codes.ResourceExhausted, "rate limit")
			}
			vv = append(vv, t)
			rateMap[key] = vv
			rateMu.Unlock()
		}
		return handler(ctx, req)
	}
}

func isAuthMethod(method string) bool {
	return method == "/auth.v1.AuthService/Login" || method == "/auth.v1.AuthService/Register" || method == "/auth.v1.AuthService/Refresh" || method == "/auth.v1.AuthService/GetPublicKey"
}

func New(addr string) (*Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil { return nil, fmt.Errorf("listen: %w", err) }
	s := &Server{gs: grpc.NewServer(grpc.UnaryInterceptor(rateLimitUnaryInterceptor(60))), lis: lis}
	authv1.RegisterAuthServiceServer(s.gs, s)
	msgv1.RegisterMessagingServiceServer(s.gs, s)
	stgv1.RegisterStorageServiceServer(s.gs, s)
	return s, nil
}

func (s *Server) Start() error {
	log.Printf("grpc listening on %s", s.lis.Addr())
	return s.gs.Serve(s.lis)
}

func (s *Server) Stop() {
	s.gs.GracefulStop()
}

// Auth

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if s.Collector != nil { s.Collector.RecordMessage("auth", "register") }
	id, err := s.AuthSvc.Register(ctx, req.Username, req.PasswordProof, req.ClientPubkey)
	if err != nil { if s.Collector != nil { s.Collector.RecordMessage("auth", "register_failed") }; return nil, err }
	return &authv1.RegisterResponse{UserId: id}, nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if s.Collector != nil { s.Collector.RecordMessage("auth", "login") }
	now := time.Now()
	access, refresh, exp, err := s.AuthSvc.Login(ctx, req.Username, req.PasswordProof, req.DeviceId, req.SecondaryCode, now)
	if err != nil { if s.Collector != nil { s.Collector.RecordMessage("auth", "login_failed") }; return nil, err }
	return &authv1.LoginResponse{AccessToken: access, RefreshToken: refresh, ExpiresAtUnix: exp.Unix()}, nil
}

func (s *Server) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	now := time.Now()
	tok, exp, err := s.AuthSvc.Refresh(ctx, req.RefreshToken, req.DeviceId, now)
	if err != nil { return nil, err }
	return &authv1.RefreshResponse{AccessToken: tok, ExpiresAtUnix: exp.Unix()}, nil
}

func (s *Server) GetPublicKey(ctx context.Context, req *authv1.GetPublicKeyRequest) (*authv1.GetPublicKeyResponse, error) {
	pk, err := s.AuthSvc.GetPublicKey(ctx, req.UserId)
	if err != nil { return nil, err }
	return &authv1.GetPublicKeyResponse{PublicKey: pk}, nil
}

// Messaging

func (s *Server) Send(ctx context.Context, req *msgv1.SendRequest) (*msgv1.SendResponse, error) {
	if s.Collector != nil { s.Collector.RecordMessage("msg", "send") }
	b, err := json.Marshal(req.Envelope)
	if err != nil { if s.Collector != nil { s.Collector.RecordMessage("msg", "send_failed") }; return nil, err }
	if err := s.MessagingSvc.Send(ctx, b); err != nil { if s.Collector != nil { s.Collector.RecordMessage("msg", "send_failed") }; return nil, err }
	return &msgv1.SendResponse{Success: true}, nil
}

func (s *Server) Pull(ctx context.Context, req *msgv1.PullRequest) (*msgv1.PullResponse, error) {
	limit := int(req.PageSize)
	if limit <= 0 { limit = 100 }
	envs, next, more, err := s.MessagingSvc.PullPage(ctx, req.ConversationId, req.SinceUnix, limit)
	if err != nil { return nil, err }
	out := make([]*msgv1.Envelope, 0, len(envs))
	for _, b := range envs {
		var e msgv1.Envelope
		if json.Unmarshal(b, &e) == nil {
			out = append(out, &e)
		}
	}
	return &msgv1.PullResponse{Envelopes: out, NextSinceUnix: next, HasMore: more}, nil
}

// Storage

func (s *Server) PutFile(stream stgv1.StorageService_PutFileServer) error {
	if s.Collector != nil { s.Collector.RecordFileOp("upload", "start") }
	var chunks [][]byte
	var name, mime string
	var size int64
	var b3 []byte
	var totalBytes int64
	for {
		req, err := stream.Recv()
		if err != nil { break }
		name, mime, size = req.Name, req.Mime, req.SizeBytes
		if len(b3) == 0 && len(req.TotalBlake3) > 0 { b3 = append([]byte(nil), req.TotalBlake3...) }
		chunks = append(chunks, req.EncryptedCarChunk)
		totalBytes += int64(len(req.EncryptedCarChunk))
		if req.LastChunk { break }
	}
	fileID, cid, err := s.StorageSvc.PutCAR(stream.Context(), name, mime, size, chunks, b3)
	if err != nil { if s.Collector != nil { s.Collector.RecordFileOp("upload", "failed") }; return err }
	if s.Collector != nil { s.Collector.RecordFileOp("upload", "success"); s.Collector.AddCARBytes(totalBytes) }
	return stream.SendAndClose(&stgv1.PutFileResponse{Accepted: true, FileId: fileID, Cid: cid})
}

func (s *Server) GetFile(req *stgv1.GetFileRequest, stream stgv1.StorageService_GetFileServer) error {
	if s.Collector != nil { s.Collector.RecordFileOp("download", "start") }
	chunks, err := s.StorageSvc.GetCAR(stream.Context(), req.Cid)
	if err != nil { if s.Collector != nil { s.Collector.RecordFileOp("download", "failed") }; return err }
	var totalBytes int64
	for i, ch := range chunks {
		totalBytes += int64(len(ch))
		if err := stream.Send(&stgv1.GetFileResponse{EncryptedCarChunk: ch, LastChunk: i == len(chunks)-1}); err != nil {
			if s.Collector != nil { s.Collector.RecordFileOp("download", "failed") }
			return err
		}
	}
	if s.Collector != nil { s.Collector.RecordFileOp("download", "success"); s.Collector.AddCARBytes(totalBytes) }
	return nil
}

func (s *Server) GetActivePeers(ctx context.Context, req *msgv1.GetActivePeersRequest) (*msgv1.GetActivePeersResponse, error) {
	if s.StreamMgr == nil { return nil, status.Error(codes.Unimplemented, "p2p not configured") }
	peers := s.StreamMgr.GetActivePeers()
	return &msgv1.GetActivePeersResponse{PeerIds: peers}, nil
}

func (s *Server) GetRelayChains(ctx context.Context, req *msgv1.GetRelayChainsRequest) (*msgv1.GetRelayChainsResponse, error) {
	if s.RelayMgr == nil { return nil, status.Error(codes.Unimplemented, "relay not configured") }
	chains := s.RelayMgr.GetChains()
	resp := &msgv1.GetRelayChainsResponse{Chains: make([]*msgv1.RelayChain, 0, len(chains))}
	for id, chain := range chains {
		resp.Chains = append(resp.Chains, &msgv1.RelayChain{Id: id, RelayCount: int32(len(chain.GetRelayAddrs()))})
	}
	return resp, nil
}

func (s *Server) GetRoutingMetrics(ctx context.Context, req *msgv1.GetRoutingMetricsRequest) (*msgv1.GetRoutingMetricsResponse, error) {
	if s.Router == nil { return nil, status.Error(codes.Unimplemented, "routing not configured") }
	transports := s.Router.GetTransports()
	resp := &msgv1.GetRoutingMetricsResponse{Transports: make([]*msgv1.TransportMetrics, 0, len(transports))}
	for _, t := range transports {
		metrics := s.Router.GetMetrics(t.ID)
		resp.Transports = append(resp.Transports, &msgv1.TransportMetrics{
			TransportId:  t.ID,
			LatencyMs:    int32(metrics.Latency.Milliseconds()),
			PacketLoss:   float32(metrics.PacketLoss),
			JitterMs:     int32(metrics.Jitter.Milliseconds()),
			Stability:    float32(metrics.Stability),
			BlockingRisk: float32(metrics.BlockingRisk),
			Load:         float32(metrics.Load),
		})
	}
	return resp, nil
}
