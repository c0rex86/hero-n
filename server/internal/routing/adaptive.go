package routing

import (
    "context"
    "sync"
    "time"
)

// метрики маршрута
type RouteMetrics struct {
    Latency      time.Duration // задержка rtt
    PacketLoss   float64       // потери пакетов 0-1
    Jitter       time.Duration // джиттер
    Stability    float64       // стабильность 0-1
    BlockingRisk float64       // риск блокировки 0-1
    Load         float64       // загрузка 0-1
    LastUpdate   time.Time
}

// транспорт
type Transport struct {
    ID       string
    Type     string // "p2p", "relay", "wss", "quic"
    Endpoint string
    Priority int
}

// адаптивный роутер
type AdaptiveRouter struct {
    mu        sync.RWMutex
    routes    map[string]*RouteMetrics // transport id -> metrics
    transports []Transport
    window    time.Duration
    threshold time.Duration
    
    // веса для скоринга
    weights struct {
        latency      float64
        loss         float64
        jitter       float64
        stability    float64
        blockingRisk float64
        load         float64
    }
}

func NewAdaptiveRouter(windowSec int, thresholdMs int) *AdaptiveRouter {
    return &AdaptiveRouter{
        routes:    make(map[string]*RouteMetrics),
        window:    time.Duration(windowSec) * time.Second,
        threshold: time.Duration(thresholdMs) * time.Millisecond,
        weights: struct {
            latency      float64
            loss         float64
            jitter       float64
            stability    float64
            blockingRisk float64
            load         float64
        }{
            latency:      0.3,
            loss:         0.25,
            jitter:       0.15,
            stability:    0.15,
            blockingRisk: 0.1,
            load:         0.05,
        },
    }
}

// обновить метрики транспорта
func (r *AdaptiveRouter) UpdateMetrics(transportID string, m RouteMetrics) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    m.LastUpdate = time.Now()
    r.routes[transportID] = &m
}

// получить лучший транспорт
func (r *AdaptiveRouter) SelectBestTransport() *Transport {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    if len(r.transports) == 0 {
        return nil
    }
    
    var best *Transport
    bestScore := -1.0
    
    for _, t := range r.transports {
        score := r.scoreTransport(t.ID)
        if score > bestScore {
            bestScore = score
            best = &t
        }
    }
    
    return best
}

// расчет скора транспорта
func (r *AdaptiveRouter) scoreTransport(id string) float64 {
    m, ok := r.routes[id]
    if !ok || time.Since(m.LastUpdate) > r.window {
        return 0.0 // нет свежих метрик
    }
    
    // нормализация метрик
    latencyScore := 1.0 - min(m.Latency.Seconds()/r.threshold.Seconds(), 1.0)
    lossScore := 1.0 - m.PacketLoss
    jitterScore := 1.0 - min(m.Jitter.Seconds()/(r.threshold.Seconds()*0.5), 1.0)
    stabilityScore := m.Stability
    blockingScore := 1.0 - m.BlockingRisk
    loadScore := 1.0 - m.Load
    
    // взвешенная сумма
    total := r.weights.latency * latencyScore +
        r.weights.loss * lossScore +
        r.weights.jitter * jitterScore +
        r.weights.stability * stabilityScore +
        r.weights.blockingRisk * blockingScore +
        r.weights.load * loadScore
    
    return total
}

// добавить транспорт
func (r *AdaptiveRouter) AddTransport(t Transport) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // проверка дублей
    for i, existing := range r.transports {
        if existing.ID == t.ID {
            r.transports[i] = t
            return
        }
    }
    
    r.transports = append(r.transports, t)
}

// удалить транспорт
func (r *AdaptiveRouter) RemoveTransport(id string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    for i, t := range r.transports {
        if t.ID == id {
            r.transports = append(r.transports[:i], r.transports[i+1:]...)
            delete(r.routes, id)
            return
        }
    }
}

// активный пробинг транспорта
func (r *AdaptiveRouter) ProbeTransport(ctx context.Context, id string) error {
    // тут должен быть реальный пробинг
    // пока заглушка
    metrics := RouteMetrics{
        Latency:      50 * time.Millisecond,
        PacketLoss:   0.01,
        Jitter:       5 * time.Millisecond,
        Stability:    0.95,
        BlockingRisk: 0.1,
        Load:         0.3,
    }
    
    r.UpdateMetrics(id, metrics)
    return nil
}

// миграция на новый транспорт
func (r *AdaptiveRouter) MigrateTransport(from, to string) error {
    // логика миграции активных сессий
    // пока заглушка
    return nil
}

func min(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}
