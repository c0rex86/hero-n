package transport

import (
    "context"
    "crypto/tls"
    "errors"
    "net"
    "net/http"
    "sync"
    "time"
    
    "github.com/quic-go/quic-go"
    "github.com/quic-go/quic-go/http3"
    "golang.org/x/net/websocket"
)

// типы транспорта
const (
    TypeTCP  = "tcp"
    TypeQUIC = "quic"
    TypeWSS  = "wss"
    TypeP2P  = "p2p"
)

// транспорт интерфейс
type Transport interface {
    Type() string
    Dial(ctx context.Context, addr string) (net.Conn, error)
    Listen(addr string) (net.Listener, error)
    Close() error
}

// менеджер транспортов
type Manager struct {
    mu         sync.RWMutex
    transports map[string]Transport
    active     string // активный транспорт
    fallbacks  []string // порядок fallback
}

func NewManager() *Manager {
    return &Manager{
        transports: make(map[string]Transport),
        fallbacks:  []string{TypeP2P, TypeQUIC, TypeWSS, TypeTCP},
    }
}

// добавить транспорт
func (m *Manager) AddTransport(t Transport) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.transports[t.Type()] = t
    
    // первый добавленный становится активным
    if m.active == "" {
        m.active = t.Type()
    }
}

// установить активный транспорт
func (m *Manager) SetActive(typ string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if _, ok := m.transports[typ]; !ok {
        return errors.New("transport not found")
    }
    
    m.active = typ
    return nil
}

// подключиться с fallback
func (m *Manager) DialWithFallback(ctx context.Context, addr string) (net.Conn, error) {
    m.mu.RLock()
    order := append([]string{m.active}, m.fallbacks...)
    m.mu.RUnlock()
    
    var lastErr error
    
    for _, typ := range order {
        m.mu.RLock()
        t, ok := m.transports[typ]
        m.mu.RUnlock()
        
        if !ok {
            continue
        }
        
        conn, err := t.Dial(ctx, addr)
        if err == nil {
            return conn, nil
        }
        
        lastErr = err
    }
    
    if lastErr != nil {
        return nil, lastErr
    }
    
    return nil, errors.New("no transports available")
}

// tcp транспорт
type TCPTransport struct {
    dialer *net.Dialer
}

func NewTCPTransport() *TCPTransport {
    return &TCPTransport{
        dialer: &net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        },
    }
}

func (t *TCPTransport) Type() string {
    return TypeTCP
}

func (t *TCPTransport) Dial(ctx context.Context, addr string) (net.Conn, error) {
    return t.dialer.DialContext(ctx, "tcp", addr)
}

func (t *TCPTransport) Listen(addr string) (net.Listener, error) {
    return net.Listen("tcp", addr)
}

func (t *TCPTransport) Close() error {
    return nil
}

// quic транспорт
type QUICTransport struct {
    tlsConfig *tls.Config
}

func NewQUICTransport(tlsConfig *tls.Config) *QUICTransport {
    if tlsConfig == nil {
        tlsConfig = &tls.Config{
            InsecureSkipVerify: true, // для разработки
        }
    }
    
    return &QUICTransport{
        tlsConfig: tlsConfig,
    }
}

func (q *QUICTransport) Type() string {
    return TypeQUIC
}

func (q *QUICTransport) Dial(ctx context.Context, addr string) (net.Conn, error) {
    session, err := quic.DialAddr(ctx, addr, q.tlsConfig, &quic.Config{
        MaxIdleTimeout:  30 * time.Second,
        KeepAlivePeriod: 10 * time.Second,
    })
    if err != nil {
        return nil, err
    }
    
    stream, err := session.OpenStreamSync(ctx)
    if err != nil {
        return nil, err
    }
    
    return &quicConn{stream: stream, session: session}, nil
}

func (q *QUICTransport) Listen(addr string) (net.Listener, error) {
    listener, err := quic.ListenAddr(addr, q.tlsConfig, &quic.Config{
        MaxIdleTimeout:  30 * time.Second,
        KeepAlivePeriod: 10 * time.Second,
    })
    if err != nil {
        return nil, err
    }
    
    return &quicListener{listener: listener}, nil
}

func (q *QUICTransport) Close() error {
    return nil
}

// wss транспорт
type WSSTransport struct {
    client *http.Client
}

func NewWSSTransport() *WSSTransport {
    return &WSSTransport{
        client: &http.Client{
            Transport: &http3.RoundTripper{
                TLSClientConfig: &tls.Config{
                    InsecureSkipVerify: true, // для разработки
                },
            },
        },
    }
}

func (w *WSSTransport) Type() string {
    return TypeWSS
}

func (w *WSSTransport) Dial(ctx context.Context, addr string) (net.Conn, error) {
    wsConfig, err := websocket.NewConfig("wss://"+addr+"/ws", "https://"+addr)
    if err != nil {
        return nil, err
    }
    
    ws, err := websocket.DialConfig(wsConfig)
    if err != nil {
        return nil, err
    }
    
    return ws, nil
}

func (w *WSSTransport) Listen(addr string) (net.Listener, error) {
    // wss требует http сервера
    return nil, errors.New("wss listen not implemented")
}

func (w *WSSTransport) Close() error {
    return nil
}

// обертки для quic
type quicConn struct {
    stream  quic.Stream
    session quic.Connection
}

func (c *quicConn) Read(b []byte) (int, error) {
    return c.stream.Read(b)
}

func (c *quicConn) Write(b []byte) (int, error) {
    return c.stream.Write(b)
}

func (c *quicConn) Close() error {
    c.stream.Close()
    return c.session.CloseWithError(0, "")
}

func (c *quicConn) LocalAddr() net.Addr {
    return c.session.LocalAddr()
}

func (c *quicConn) RemoteAddr() net.Addr {
    return c.session.RemoteAddr()
}

func (c *quicConn) SetDeadline(t time.Time) error {
    return c.stream.SetDeadline(t)
}

func (c *quicConn) SetReadDeadline(t time.Time) error {
    return c.stream.SetReadDeadline(t)
}

func (c *quicConn) SetWriteDeadline(t time.Time) error {
    return c.stream.SetWriteDeadline(t)
}

type quicListener struct {
    listener *quic.Listener
}

func (l *quicListener) Accept() (net.Conn, error) {
    session, err := l.listener.Accept(context.Background())
    if err != nil {
        return nil, err
    }
    
    stream, err := session.AcceptStream(context.Background())
    if err != nil {
        return nil, err
    }
    
    return &quicConn{stream: stream, session: session}, nil
}

func (l *quicListener) Close() error {
    return l.listener.Close()
}

func (l *quicListener) Addr() net.Addr {
    return l.listener.Addr()
}
