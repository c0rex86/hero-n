package routing

import "time"

type Candidate struct {
	ID          string
	LatencyMs   float64
	LossPct     float64
	JitterMs    float64
	Stability   float64
	BlockRisk   float64
	Load        float64
	Priority    int
}

type Engine struct {
	window time.Duration
	thresholdMs int
	weights Weights
}

type Weights struct {
	Latency float64
	Loss    float64
	Jitter  float64
	Stab    float64
	Block   float64
	Load    float64
}

func New(windowSec int, thresholdMs int) *Engine {
	return &Engine{window: time.Duration(windowSec) * time.Second, thresholdMs: thresholdMs, weights: Weights{Latency: 0.4, Loss: 0.2, Jitter: 0.1, Stab: 0.2, Block: 0.05, Load: 0.05}}
}

func (e *Engine) Score(c Candidate) float64 {
	// lower score is better
	return e.weights.Latency*(c.LatencyMs/1000.0) + e.weights.Loss*(c.LossPct/100.0) + e.weights.Jitter*(c.JitterMs/1000.0) + e.weights.Stab*(1.0-c.Stability) + e.weights.Block*c.BlockRisk + e.weights.Load*c.Load - float64(c.Priority)*0.1
}
