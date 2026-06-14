package publicengagement

import "testing"

func TestNewParametricBridgeInEngagement(t *testing.T) {
	pb := NewParametricBridge(1, 25.6, 5.8, 6.5)
	if pb == nil {
		t.Fatal("NewParametricBridge returned nil")
	}
	if pb.BaseBridge == nil {
		t.Fatal("BaseBridge is nil")
	}
}

func TestValidateParamsInEngagement(t *testing.T) {
	pb := NewParametricBridge(1, 25.6, 5.8, 6.5)
	valid, _ := pb.ValidateParams(25.6, 5.8, 6.5)
	if !valid {
		t.Error("expected valid=true for span=25.6, rise=5.8, width=6.5")
	}
	valid, _ = pb.ValidateParams(5, 5.8, 6.5)
	if valid {
		t.Error("expected valid=false for span=5")
	}
}

func TestCalculateParametricAnalysisInEngagement(t *testing.T) {
	result := CalculateParametricAnalysis(1, 25.6, 5.8, 6.5, 50)
	if result.MaxStressRatio <= 0 {
		t.Error("expected MaxStressRatio > 0")
	}
	if result.MemberCount <= 0 {
		t.Error("expected MemberCount > 0")
	}
}

func TestSurrogateTrainAndPredict(t *testing.T) {
	pb := NewParametricBridge(1, 25.6, 5.8, 6.5)
	pb.SetConstraints(15, 40, 3, 10, 4, 8)
	pb.TrainSurrogate(50)
	if pb.Surrogate.IsTrained != true {
		t.Error("expected Surrogate.IsTrained == true")
	}
	result := pb.AnalyzeWithSurrogate(25, 6, 6)
	if result == nil {
		t.Error("expected non-nil result from AnalyzeWithSurrogate")
	}
}

func TestGeometryOptionsInEngagement(t *testing.T) {
	pb := NewParametricBridge(1, 25.6, 5.8, 6.5)
	opts := pb.GetGeometryOptions()
	if len(opts) != 3 {
		t.Errorf("expected len == 3, got %d", len(opts))
	}
}
