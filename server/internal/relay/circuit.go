package relay

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
    "github.com/multiformats/go-multiaddr"
)

type CircuitChain struct {
    relays []host.Host
    mu     sync.RWMutex
}

func NewCircuitChain(ctx context.Context, count int) (*CircuitChain, error) {
    chain := &CircuitChain{
        relays: make([]host.Host, 0, count),
    }
    
    for i := 0; i < count; i++ {
        port := 4100 + i
        addr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)
        
        h, err := createRelayHost(ctx, addr)
        if err != nil {
            chain.Close()
            return nil, err
        }
        
        chain.relays = append(chain.relays, h)
    }
    
    return chain, nil
}

func createRelayHost(ctx context.Context, addr string) (host.Host, error) {
    ma, err := multiaddr.NewMultiaddr(addr)
    if err != nil {
        return nil, err
    }
    
    h, err := libp2p.New(
        libp2p.ListenAddrs(ma),
        libp2p.EnableRelay(),
        libp2p.EnableAutoRelay(),
    )
    if err != nil {
        return nil, err
    }
    
    _, err = relay.New(h,
        relay.WithLimit(nil),
        relay.WithInfiniteLimits(),
    )
    if err != nil {
        h.Close()
        return nil, err
    }
    
    return h, nil
}

func (c *CircuitChain) GetRelayAddrs() []peer.AddrInfo {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    addrs := make([]peer.AddrInfo, 0, len(c.relays))
    for _, r := range c.relays {
        addrs = append(addrs, peer.AddrInfo{
            ID:    r.ID(),
            Addrs: r.Addrs(),
        })
    }
    
    return addrs
}

func (c *CircuitChain) BuildPath(ctx context.Context, target peer.ID) ([]peer.ID, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if len(c.relays) < 2 {
        return nil, fmt.Errorf("need at least 2 relays")
    }
    
    path := make([]peer.ID, 0, len(c.relays)+1)
    for _, r := range c.relays {
        path = append(path, r.ID())
    }
    path = append(path, target)
    
    return path, nil
}

func (c *CircuitChain) RotateRelays(ctx context.Context) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if len(c.relays) == 0 {
        return nil
    }
    
    first := c.relays[0]
    c.relays = append(c.relays[1:], first)
    
    return nil
}

func (c *CircuitChain) AddRelay(ctx context.Context, addr string) error {
    h, err := createRelayHost(ctx, addr)
    if err != nil {
        return err
    }
    
    c.mu.Lock()
    c.relays = append(c.relays, h)
    c.mu.Unlock()
    
    return nil
}

func (c *CircuitChain) RemoveRelay(peerID peer.ID) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    for i, r := range c.relays {
        if r.ID() == peerID {
            r.Close()
            c.relays = append(c.relays[:i], c.relays[i+1:]...)
            return
        }
    }
}

func (c *CircuitChain) Close() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    for _, r := range c.relays {
        r.Close()
    }
    c.relays = nil
}

type RelayManager struct {
    chains map[string]*CircuitChain
    mu     sync.RWMutex
}

func NewRelayManager() *RelayManager {
    return &RelayManager{
        chains: make(map[string]*CircuitChain),
    }
}

func (rm *RelayManager) CreateChain(ctx context.Context, id string, relayCount int) error {
    chain, err := NewCircuitChain(ctx, relayCount)
    if err != nil {
        return err
    }
    
    rm.mu.Lock()
    rm.chains[id] = chain
    rm.mu.Unlock()
    
    return nil
}

func (rm *RelayManager) GetChain(id string) *CircuitChain {
    rm.mu.RLock()
    defer rm.mu.RUnlock()
    return rm.chains[id]
}

func (rm *RelayManager) RotateAll(ctx context.Context) {
    rm.mu.RLock()
    chains := make([]*CircuitChain, 0, len(rm.chains))
    for _, c := range rm.chains {
        chains = append(chains, c)
    }
    rm.mu.RUnlock()
    
    for _, c := range chains {
        c.RotateRelays(ctx)
    }
}

func (rm *RelayManager) StartRotation(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for {
            select {
            case <-ticker.C:
                rm.RotateAll(ctx)
            case <-ctx.Done():
                ticker.Stop()
                return
            }
        }
    }()
}

func (rm *RelayManager) Close() {
    rm.mu.Lock()
    defer rm.mu.Unlock()

    for _, c := range rm.chains {
        c.Close()
    }
    rm.chains = nil
}

func (rm *RelayManager) GetChains() map[string]*CircuitChain {
    rm.mu.RLock()
    defer rm.mu.RUnlock()

    chains := make(map[string]*CircuitChain)
    for id, chain := range rm.chains {
        chains[id] = chain
    }
    return chains
}
