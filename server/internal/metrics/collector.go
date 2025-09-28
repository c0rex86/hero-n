package metrics

import (
    "context"
    "sync"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
)

var (
    MessagesProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "heroin_messages_processed_total",
            Help: "Total messages processed",
        },
        []string{"type", "status"},
    )
    
    FileOperations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "heroin_file_operations_total",
            Help: "Total file operations",
        },
        []string{"operation", "status"},
    )
    
    ActivePeers = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "heroin_active_peers",
            Help: "Number of active P2P peers",
        },
    )
    
    TransportLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "heroin_transport_latency_seconds",
            Help:    "Transport latency in seconds",
            Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
        },
        []string{"transport"},
    )
    
    GroupOperations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "heroin_group_operations_total",
            Help: "Total group operations",
        },
        []string{"operation"},
    )
    
    RelayHops = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "heroin_relay_hops",
            Help:    "Number of relay hops",
            Buckets: []float64{1, 2, 3, 4, 5},
        },
    )
    
    CARStreamBytes = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "heroin_car_stream_bytes_total",
            Help: "Total bytes streamed via CAR",
        },
    )
    
    DHTPeersFound = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "heroin_dht_peers_found",
            Help:    "Number of peers found via DHT",
            Buckets: []float64{0, 1, 5, 10, 20, 50},
        },
    )
)

func init() {
    prometheus.MustRegister(
        MessagesProcessed,
        FileOperations,
        ActivePeers,
        TransportLatency,
        GroupOperations,
        RelayHops,
        CARStreamBytes,
        DHTPeersFound,
    )
}

type Collector struct {
    mu      sync.RWMutex
    metrics map[string]interface{}
}

func NewCollector() *Collector {
    return &Collector{
        metrics: make(map[string]interface{}),
    }
}

func (c *Collector) RecordMessage(msgType, status string) {
    MessagesProcessed.WithLabelValues(msgType, status).Inc()
}

func (c *Collector) RecordFileOp(op, status string) {
    FileOperations.WithLabelValues(op, status).Inc()
}

func (c *Collector) SetActivePeers(count int) {
    ActivePeers.Set(float64(count))
}

func (c *Collector) RecordTransportLatency(transport string, latency time.Duration) {
    TransportLatency.WithLabelValues(transport).Observe(latency.Seconds())
}

func (c *Collector) RecordGroupOp(op string) {
    GroupOperations.WithLabelValues(op).Inc()
}

func (c *Collector) RecordRelayHops(hops int) {
    RelayHops.Observe(float64(hops))
}

func (c *Collector) AddCARBytes(bytes int64) {
    CARStreamBytes.Add(float64(bytes))
}

func (c *Collector) RecordDHTPeers(count int) {
    DHTPeersFound.Observe(float64(count))
}

func (c *Collector) StartPeriodicCollection(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for {
            select {
            case <-ticker.C:
                c.collect()
            case <-ctx.Done():
                ticker.Stop()
                return
            }
        }
    }()
}

func (c *Collector) collect() {
    // периодический сбор метрик
}
