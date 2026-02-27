package metrics

import (
	"testing"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector()
	if c == nil {
		t.Fatal("expected non-nil collector")
	}
}

func TestUpdateFromResponse(t *testing.T) {
	c := NewCollector()
	c.UpdateFromResponse(24.5, 100, 50)

	if c.LastTokensPerSec != 24.5 {
		t.Errorf("expected 24.5 t/s, got %f", c.LastTokensPerSec)
	}
	if c.TotalTokens != 50 {
		t.Errorf("expected 50 total tokens, got %d", c.TotalTokens)
	}
	if c.TotalMessages != 1 {
		t.Errorf("expected 1 message, got %d", c.TotalMessages)
	}

	// Second response
	c.UpdateFromResponse(30.0, 150, 75)
	if c.TotalTokens != 125 {
		t.Errorf("expected 125 total tokens, got %d", c.TotalTokens)
	}
	if c.TotalMessages != 2 {
		t.Errorf("expected 2 messages, got %d", c.TotalMessages)
	}
}

func TestUpdateContext(t *testing.T) {
	c := NewCollector()
	c.UpdateContext(2048, 4096)

	if c.ContextPercent != 50.0 {
		t.Errorf("expected 50%%, got %f", c.ContextPercent)
	}

	got := c.ContextString()
	if got != "CTX 50% (2048/4096)" {
		t.Errorf("unexpected context string: %s", got)
	}
}

func TestSpeedString(t *testing.T) {
	c := NewCollector()
	if c.SpeedString() != "— t/s" {
		t.Errorf("expected '— t/s' before any response, got %q", c.SpeedString())
	}

	c.UpdateFromResponse(24.5, 100, 50)
	if c.SpeedString() != "24.5 t/s" {
		t.Errorf("expected '24.5 t/s', got %q", c.SpeedString())
	}
}

func TestSnapshot(t *testing.T) {
	c := NewCollector()
	c.UpdateFromResponse(30.0, 100, 50)
	c.UpdateContext(1000, 4096)
	c.SetHardware(8, 16*1024*1024*1024, true, 999)

	snap := c.Snapshot()
	if snap.TokensPerSec != 30.0 {
		t.Errorf("expected 30.0, got %f", snap.TokensPerSec)
	}
	if !snap.GPUActive {
		t.Error("expected GPU active")
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := NewCollector()
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 1000; i++ {
			c.UpdateFromResponse(float64(i), i, i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			c.Snapshot()
			c.SpeedString()
			c.ContextString()
		}
		done <- true
	}()

	<-done
	<-done
}
