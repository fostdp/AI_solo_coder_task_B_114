package bridgecomparator

import "testing"

func TestNewHistoricalComparison(t *testing.T) {
	hc := NewHistoricalComparison()
	if hc == nil {
		t.Fatal("NewHistoricalComparison() returned nil")
	}
}

func TestGetAllBridgesCount(t *testing.T) {
	hc := NewHistoricalComparison()
	bridges := hc.GetAllBridges()
	if len(bridges) != 7 {
		t.Fatalf("expected 7 bridges, got %d", len(bridges))
	}
}

func TestArchaeologicalEstimateInComparator(t *testing.T) {
	hc := NewHistoricalComparison()
	bridge := HistoricalBridge{ID: 99, Name: "test", Dynasty: DynastyTang, SpanLength: 20, ArchRise: 0, DeckWidth: 0, TotalLength: 0, DataConfidence: 0.3}
	hc.ArchaeologicalEstimate(&bridge)
	if bridge.ArchRise <= 0 {
		t.Fatal("ArchRise should be > 0 after estimation")
	}
	if bridge.DeckWidth <= 0 {
		t.Fatal("DeckWidth should be > 0 after estimation")
	}
	if len(bridge.EstimatedFields) < 1 {
		t.Fatalf("expected at least 1 estimated field, got %d", len(bridge.EstimatedFields))
	}
}

func TestCompareBridgesInComparator(t *testing.T) {
	hc := NewHistoricalComparison()
	bridges := hc.GetAllBridges()
	result := hc.CompareBridges(bridges[0], bridges[1])
	if result.MetricsA == (EfficiencyMetrics{}) {
		t.Fatal("MetricsA should not be zero value")
	}
	if result.MetricsB == (EfficiencyMetrics{}) {
		t.Fatal("MetricsB should not be zero value")
	}
}

func TestDataReliabilityInComparator(t *testing.T) {
	hc := NewHistoricalComparison()
	bridges := hc.GetAllBridges()
	for _, b := range bridges {
		if b.DataConfidence < 0 || b.DataConfidence > 1 {
			t.Fatalf("bridge %s DataConfidence %f not in [0,1]", b.Name, b.DataConfidence)
		}
	}
}
