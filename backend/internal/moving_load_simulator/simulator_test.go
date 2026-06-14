package movingloadsimulator

import (
	"testing"
)

func TestSimulateAsyncCompletes(t *testing.T) {
	ch := SimulateAsync(AsyncInput{
		BridgeID:         1,
		SpanLength:       25.6,
		ArchRise:         5.8,
		DeckWidth:        6.5,
		Duration:         10,
		CrowdDensity:     0.5,
		TimeStep:         0.2,
		RandomSeed:       42,
		LoadCyclesPerDay: 500,
	})
	result := <-ch
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.MaxLoad <= 0 {
		t.Fatalf("expected MaxLoad > 0, got %f", result.MaxLoad)
	}
	if result.TotalAgents <= 0 {
		t.Fatalf("expected TotalAgents > 0, got %d", result.TotalAgents)
	}
	if result.FatigueResult.TotalDamage < 0 {
		t.Fatalf("expected FatigueResult.TotalDamage >= 0, got %f", result.FatigueResult.TotalDamage)
	}
}

func TestSimulateAsyncPanicRecovery(t *testing.T) {
	ch := SimulateAsync(AsyncInput{
		SpanLength: 0,
	})
	result := <-ch
	_ = result
}

func TestAsyncChannelClosed(t *testing.T) {
	ch := SimulateAsync(AsyncInput{
		BridgeID:         1,
		SpanLength:       25.6,
		ArchRise:         5.8,
		DeckWidth:        6.5,
		Duration:         10,
		CrowdDensity:     0.5,
		TimeStep:         0.2,
		RandomSeed:       42,
		LoadCyclesPerDay: 500,
	})
	count := 0
	for range ch {
		count++
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 value from channel, got %d", count)
	}
}

func TestNewSocialForceModelFromSimulator(t *testing.T) {
	sfm := NewSocialForceModel(25.6)
	if sfm == nil {
		t.Fatal("expected non-nil SocialForceModel")
	}
	if sfm.SyncFactor != 0.1 {
		t.Fatalf("expected SyncFactor 0.1, got %f", sfm.SyncFactor)
	}
}

func TestNewFatigueAnalysisFromSimulator(t *testing.T) {
	fa := NewFatigueAnalysis(25.6, 31)
	if fa == nil {
		t.Fatal("expected non-nil FatigueAnalysis")
	}
	if fa.FatigueLimit != 8.5 {
		t.Fatalf("expected FatigueLimit 8.5, got %f", fa.FatigueLimit)
	}
}
