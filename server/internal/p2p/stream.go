package p2p

import (
    "context"
    "encoding/binary"
    "errors"
    "io"
    "sync"
    
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/core/protocol"
)

const (
    ProtocolID = "/heroin/msg/1.0.0"
    MaxMessageSize = 10 * 1024 * 1024 // 10mb
)

type StreamHandler func(peer.ID, []byte) error

type StreamManager struct {
    host    host.Host
    handler StreamHandler
    
    mu      sync.RWMutex
    streams map[peer.ID]network.Stream
}

func NewStreamManager(h host.Host) *StreamManager {
    sm := &StreamManager{
        host:    h,
        streams: make(map[peer.ID]network.Stream),
    }
    
    h.SetStreamHandler(protocol.ID(ProtocolID), sm.handleStream)
    return sm
}

func (sm *StreamManager) SetHandler(h StreamHandler) {
    sm.handler = h
}

func (sm *StreamManager) SendMessage(ctx context.Context, peerID peer.ID, data []byte) error {
    if len(data) > MaxMessageSize {
        return errors.New("message too large")
    }
    
    stream, err := sm.getOrCreateStream(ctx, peerID)
    if err != nil {
        return err
    }
    
    sizeBuf := make([]byte, 4)
    binary.BigEndian.PutUint32(sizeBuf, uint32(len(data)))
    
    if _, err := stream.Write(sizeBuf); err != nil {
        sm.removeStream(peerID)
        return err
    }
    
    if _, err := stream.Write(data); err != nil {
        sm.removeStream(peerID)
        return err
    }
    
    return nil
}

func (sm *StreamManager) getOrCreateStream(ctx context.Context, peerID peer.ID) (network.Stream, error) {
    sm.mu.RLock()
    stream, exists := sm.streams[peerID]
    sm.mu.RUnlock()
    
    if exists && stream.Conn().IsClosed() == false {
        return stream, nil
    }
    
    newStream, err := sm.host.NewStream(ctx, peerID, protocol.ID(ProtocolID))
    if err != nil {
        return nil, err
    }
    
    sm.mu.Lock()
    sm.streams[peerID] = newStream
    sm.mu.Unlock()
    
    return newStream, nil
}

func (sm *StreamManager) removeStream(peerID peer.ID) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    if stream, exists := sm.streams[peerID]; exists {
        stream.Close()
        delete(sm.streams, peerID)
    }
}

func (sm *StreamManager) handleStream(stream network.Stream) {
    defer stream.Close()
    
    peerID := stream.Conn().RemotePeer()
    
    for {
        sizeBuf := make([]byte, 4)
        if _, err := io.ReadFull(stream, sizeBuf); err != nil {
            return
        }
        
        size := binary.BigEndian.Uint32(sizeBuf)
        if size > MaxMessageSize {
            return
        }
        
        data := make([]byte, size)
        if _, err := io.ReadFull(stream, data); err != nil {
            return
        }
        
        if sm.handler != nil {
            if err := sm.handler(peerID, data); err != nil {
                return
            }
        }
    }
}

func (sm *StreamManager) Close() {
    sm.mu.Lock()
    defer sm.mu.Unlock()

    for _, stream := range sm.streams {
        stream.Close()
    }
    sm.streams = make(map[peer.ID]network.Stream)
}

func (sm *StreamManager) GetActivePeers() []string {
    sm.mu.RLock()
    defer sm.mu.RUnlock()

    peers := make([]string, 0, len(sm.streams))
    for peerID := range sm.streams {
        peers = append(peers, peerID.String())
    }
    return peers
}
