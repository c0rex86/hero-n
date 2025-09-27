package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	host      host.Host
	discovery mdns.Service
	relay     *relay.Relay
}

type Config struct {
	ListenAddrs    []string
	BootstrapPeers []string
	EnableRelay    bool
	EnableMDNS     bool
}

func NewNode(ctx context.Context, cfg Config) (*Node, error) {
	var opts []libp2p.Option
	
	for _, addr := range cfg.ListenAddrs {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil { return nil, fmt.Errorf("invalid listen addr %s: %w", addr, err) }
		opts = append(opts, libp2p.ListenAddrs(ma))
	}
	
	h, err := libp2p.New(opts...)
	if err != nil { return nil, fmt.Errorf("create libp2p host: %w", err) }
	
	n := &Node{host: h}
	
	if cfg.EnableRelay {
		r, err := relay.New(h)
		if err != nil { return nil, fmt.Errorf("create relay: %w", err) }
		n.relay = r
	}
	
	if cfg.EnableMDNS {
		disc := mdns.NewMdnsService(h, "heroin", &discoveryNotifee{})
		if err := disc.Start(); err != nil { return nil, fmt.Errorf("start mdns: %w", err) }
		n.discovery = disc
	}
	
	if err := n.connectBootstrapPeers(ctx, cfg.BootstrapPeers); err != nil {
		log.Printf("bootstrap connect failed: %v", err)
	}
	
	log.Printf("libp2p node started, peer id: %s", h.ID())
	for _, addr := range h.Addrs() {
		log.Printf("listening on: %s/p2p/%s", addr, h.ID())
	}
	
	return n, nil
}

func (n *Node) connectBootstrapPeers(ctx context.Context, peers []string) error {
	for _, peerStr := range peers {
		ma, err := multiaddr.NewMultiaddr(peerStr)
		if err != nil { continue }
		info, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil { continue }
		if err := n.host.Connect(ctx, *info); err != nil {
			log.Printf("failed to connect to bootstrap peer %s: %v", peerStr, err)
		} else {
			log.Printf("connected to bootstrap peer: %s", info.ID)
		}
	}
	return nil
}

func (n *Node) Close() error {
	if n.discovery != nil { n.discovery.Close() }
	if n.relay != nil { n.relay.Close() }
	return n.host.Close()
}

type discoveryNotifee struct{}

func (d *discoveryNotifee) HandlePeerFound(info peer.AddrInfo) {
	log.Printf("discovered peer via mdns: %s", info.ID)
}
