package discovery

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/libp2p/go-libp2p"
    dht "github.com/libp2p/go-libp2p-kad-dht"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
    drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
    "github.com/multiformats/go-multiaddr"
)

// dht discovery сервис
type DHTDiscovery struct {
    Host      host.Host
    dht       *dht.IpfsDHT
    routingDiscovery *drouting.RoutingDiscovery

    mu        sync.RWMutex
    peers     map[peer.ID]peer.AddrInfo
    userPeers map[string][]peer.ID // user id -> peer ids
}

// конфиг discovery
type Config struct {
    ListenAddrs    []string
    BootstrapPeers []peer.AddrInfo
    Namespace      string // namespace для rendezvous
    EnableMDNS     bool
}

// создать discovery сервис
func NewDHTDiscovery(ctx context.Context, cfg Config) (*DHTDiscovery, error) {
    // парсим адреса
    listenAddrs := make([]multiaddr.Multiaddr, 0, len(cfg.ListenAddrs))
    for _, addr := range cfg.ListenAddrs {
        ma, err := multiaddr.NewMultiaddr(addr)
        if err != nil {
            return nil, fmt.Errorf("invalid listen addr %s: %w", addr, err)
        }
        listenAddrs = append(listenAddrs, ma)
    }
    
    // создаем хост
    h, err := libp2p.New(
        libp2p.ListenAddrs(listenAddrs...),
        libp2p.EnableAutoRelay(),
        libp2p.EnableNATService(),
    )
    if err != nil {
        return nil, err
    }

    // создаем dht
    kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeAutoServer))
    if err != nil {
        h.Close()
        return nil, err
    }

    // bootstrap dht
    if err = kadDHT.Bootstrap(ctx); err != nil {
        h.Close()
        return nil, err
    }

    d := &DHTDiscovery{
        Host:      h,
        dht:       kadDHT,
        routingDiscovery: drouting.NewRoutingDiscovery(kadDHT),
        peers:     make(map[peer.ID]peer.AddrInfo),
        userPeers: make(map[string][]peer.ID),
    }
    
    // подключаемся к bootstrap пирам
    for _, pinfo := range cfg.BootstrapPeers {
        go func(pi peer.AddrInfo) {
            if err := d.Host.Connect(ctx, pi); err != nil {
                // логируем ошибку
            }
        }(pinfo)
    }

    // mdns discovery для локальной сети
    if cfg.EnableMDNS {
        mdnsService := mdns.NewMdnsService(d.Host, cfg.Namespace, &mdnsNotifee{d: d})
        if err := mdnsService.Start(); err != nil {
            // не критично
        }
    }
    
    return d, nil
}

// найти пиров для юзера
func (d *DHTDiscovery) FindPeersForUser(ctx context.Context, userID string) ([]peer.AddrInfo, error) {
    d.mu.RLock()
    peerIDs, exists := d.userPeers[userID]
    d.mu.RUnlock()
    
    if exists {
        // возвращаем известных пиров
        peers := make([]peer.AddrInfo, 0, len(peerIDs))
        for _, pid := range peerIDs {
            if info, ok := d.peers[pid]; ok {
                peers = append(peers, info)
            }
        }
        return peers, nil
    }
    
    // ищем через rendezvous
    namespace := fmt.Sprintf("heroin:user:%s", userID)
    peerChan, err := d.routingDiscovery.FindPeers(ctx, namespace)
    if err != nil {
        return nil, err
    }
    
    var peers []peer.AddrInfo
    for p := range peerChan {
        if p.ID == d.Host.ID() {
            continue // пропускаем себя
        }
        peers = append(peers, p)
        
        // сохраняем для кеша
        d.mu.Lock()
        d.peers[p.ID] = p
        d.userPeers[userID] = append(d.userPeers[userID], p.ID)
        d.mu.Unlock()
    }
    
    return peers, nil
}

// анонсировать себя для юзера
func (d *DHTDiscovery) AnnounceUser(ctx context.Context, userID string) error {
    namespace := fmt.Sprintf("heroin:user:%s", userID)
    
    // анонсируем без ttl опции (не поддерживается в новой версии)
    _, err := d.routingDiscovery.Advertise(ctx, namespace)
    return err
}

// найти провайдеров контента
func (d *DHTDiscovery) FindProviders(ctx context.Context, contentID string) ([]peer.AddrInfo, error) {
    // пока заглушка, нужна конверсия в cid
    // providers := d.dht.FindProvidersAsync(ctx, cid, 10)
    
    var peers []peer.AddrInfo
    // for p := range providers {
    //     peers = append(peers, p)
    // }
    
    return peers, nil
}

// анонсировать контент
func (d *DHTDiscovery) ProvideContent(ctx context.Context, contentID string) error {
    // пока заглушка, нужна конверсия в cid
    // return d.dht.Provide(ctx, cid, true)
    return nil
}

// получить свой peer id
func (d *DHTDiscovery) PeerID() peer.ID {
    return d.Host.ID()
}

// получить адреса для подключения
func (d *DHTDiscovery) ListenAddrs() []multiaddr.Multiaddr {
    return d.Host.Addrs()
}

// закрыть discovery
func (d *DHTDiscovery) Close() error {
    if err := d.dht.Close(); err != nil {
        return err
    }
    return d.Host.Close()
}

// mdns notifee для локального discovery
type mdnsNotifee struct {
    d *DHTDiscovery
}

func (n *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
    n.d.mu.Lock()
    n.d.peers[pi.ID] = pi
    n.d.mu.Unlock()

    // пробуем подключиться
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _ = n.d.Host.Connect(ctx, pi)
}
