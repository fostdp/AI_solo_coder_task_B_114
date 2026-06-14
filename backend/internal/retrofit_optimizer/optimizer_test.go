package retrofitoptimizer

import (
	"testing"
)

func TestNewOptimizerInRetrofit(t *testing.T) {
	opt := NewMultiObjectiveOptimizer()
	if opt == nil {
		t.Fatal("NewMultiObjectiveOptimizer() returned nil")
	}
}

func TestGetReinforcementMethodsInRetrofit(t *testing.T) {
	methods := GetReinforcementMethods()
	if len(methods) != 5 {
		t.Fatalf("expected 5 methods, got %d", len(methods))
	}
	for _, m := range methods {
		if _, ok := m["method"]; !ok {
			t.Errorf("method missing 'method' key: %v", m)
		}
		if _, ok := m["bond_factor"]; !ok {
			t.Errorf("method missing 'bond_factor' key: %v", m)
		}
	}
}

func TestOptimizeInRetrofit(t *testing.T) {
	opt := NewMultiObjectiveOptimizer()
	opt.PopulationSize = 20
	opt.MaxGenerations = 5
	results := opt.Optimize(1.0, 1.0, 40)
	if len(results) < 1 {
		t.Fatal("expected at least 1 result")
	}
	for _, r := range results {
		if r.PlanID <= 0 {
			t.Errorf("expected PlanID > 0, got %d", r.PlanID)
		}
	}
}

func TestBondFactorsInRange(t *testing.T) {
	methods := GetReinforcementMethods()
	for _, m := range methods {
		bf, ok := m["bond_factor"].(float64)
		if !ok {
			t.Fatalf("bond_factor is not float64: %v", m["bond_factor"])
		}
		if bf <= 0 || bf > 1 {
			t.Errorf("bond_factor %f not in (0,1]", bf)
		}
	}
}

func TestEncodeDecodeParamsInRetrofit(t *testing.T) {
	opt := NewMultiObjectiveOptimizer()
	params := ReinforcementParams{
		Method:              MethodCFRP,
		CFRPThickness:       2.5,
		CFRPLayers:          5,
		IronHoopCount:       10,
		IronHoopWidth:       7.5,
		SteelPlateThickness: 10.0,
		WoodenSpliceLength:  1.5,
		TargetNodes:         []int{1, 2, 3},
	}
	genes := opt.EncodeParams(params)
	decoded := opt.DecodeParams(genes)
	if decoded.Method != params.Method {
		t.Errorf("Method: expected %v, got %v", params.Method, decoded.Method)
	}
	if decoded.CFRPThickness != params.CFRPThickness {
		t.Errorf("CFRPThickness: expected %v, got %v", params.CFRPThickness, decoded.CFRPThickness)
	}
	if decoded.CFRPLayers != params.CFRPLayers {
		t.Errorf("CFRPLayers: expected %v, got %v", params.CFRPLayers, decoded.CFRPLayers)
	}
	if decoded.IronHoopCount != params.IronHoopCount {
		t.Errorf("IronHoopCount: expected %v, got %v", params.IronHoopCount, decoded.IronHoopCount)
	}
	if decoded.IronHoopWidth != params.IronHoopWidth {
		t.Errorf("IronHoopWidth: expected %v, got %v", params.IronHoopWidth, decoded.IronHoopWidth)
	}
	if decoded.SteelPlateThickness != params.SteelPlateThickness {
		t.Errorf("SteelPlateThickness: expected %v, got %v", params.SteelPlateThickness, decoded.SteelPlateThickness)
	}
	if decoded.WoodenSpliceLength != params.WoodenSpliceLength {
		t.Errorf("WoodenSpliceLength: expected %v, got %v", params.WoodenSpliceLength, decoded.WoodenSpliceLength)
	}
}
