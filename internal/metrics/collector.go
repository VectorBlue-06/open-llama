package metrics

import (
	"fmt"
	"sync"
)

// Collector gathers and exposes runtime metrics.
type Collector struct {
	mu sync.RWMutex

	// Per-response metrics (updated on StreamDoneMsg)
	LastTokensPerSec    float64
	LastPromptTokens    int
	LastPredictedTokens int

	// Cumulative metrics
	TotalTokens   int
	TotalMessages int

	// Context metrics (updated on each prompt build)
	ContextUsed    int
	ContextMax     int
	ContextPercent float64

	// Hardware (set at startup)
	CPUCores  int
	RAMTotal  uint64
	RAMUsed   uint64
	GPUActive bool
	GPULayers int
}

// NewCollector creates a new metrics collector.
func NewCollector() *Collector {
	return &Collector{}
}

// UpdateFromResponse updates metrics from a completion response.
func (c *Collector) UpdateFromResponse(tokensPerSec float64, promptTokens, predictedTokens int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.LastTokensPerSec = tokensPerSec
	c.LastPromptTokens = promptTokens
	c.LastPredictedTokens = predictedTokens
	c.TotalTokens += predictedTokens
	c.TotalMessages++
}

// UpdateContext updates context usage metrics.
func (c *Collector) UpdateContext(used, max int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ContextUsed = used
	c.ContextMax = max
	if max > 0 {
		c.ContextPercent = float64(used) / float64(max) * 100
	}
}

// SetHardware sets hardware info at startup.
func (c *Collector) SetHardware(cpuCores int, ramTotal uint64, gpuActive bool, gpuLayers int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.CPUCores = cpuCores
	c.RAMTotal = ramTotal
	c.GPUActive = gpuActive
	c.GPULayers = gpuLayers
}

// SpeedString returns a formatted tokens/sec string.
func (c *Collector) SpeedString() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.LastTokensPerSec <= 0 {
		return "— t/s"
	}
	return fmt.Sprintf("%.1f t/s", c.LastTokensPerSec)
}

// ContextString returns a formatted context usage string.
func (c *Collector) ContextString() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.ContextMax <= 0 {
		return "CTX —"
	}
	return fmt.Sprintf("CTX %d%% (%d/%d)", int(c.ContextPercent), c.ContextUsed, c.ContextMax)
}

// Snapshot returns a copy of current metrics (thread-safe).
func (c *Collector) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Snapshot{
		TokensPerSec:   c.LastTokensPerSec,
		TotalTokens:    c.TotalTokens,
		TotalMessages:  c.TotalMessages,
		ContextUsed:    c.ContextUsed,
		ContextMax:     c.ContextMax,
		ContextPercent: c.ContextPercent,
		GPUActive:      c.GPUActive,
	}
}

// Snapshot is an immutable copy of metrics at a point in time.
type Snapshot struct {
	TokensPerSec   float64
	TotalTokens    int
	TotalMessages  int
	ContextUsed    int
	ContextMax     int
	ContextPercent float64
	GPUActive      bool
}
