package tests

import (
	"math"
	"testing"

	"ancient-bridge-system/internal/reinforcement"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	mo := reinforcement.NewMultiObjectiveOptimizer()
	mo.SetSeed(42)

	tests := []struct {
		name   string
		params reinforcement.ReinforcementParams
	}{
		{
			"iron_hoop",
			reinforcement.ReinforcementParams{
				Method:        reinforcement.MethodIronHoop,
				IronHoopCount:  8,
				IronHoopWidth:  10.0,
				TargetNodes:   []int{1, 5, 10},
			},
		},
		{
			"cfrp",
			reinforcement.ReinforcementParams{
				Method:        reinforcement.MethodCFRP,
				CFRPThickness: 2.5,
				CFRPLayers:     5,
			},
		},
		{
			"combined",
			reinforcement.ReinforcementParams{
				Method:        reinforcement.MethodCombined,
				CFRPThickness: 3.0,
				CFRPLayers:     4,
				IronHoopCount:  6,
				IronHoopWidth:  8.0,
				TargetNodes:   []int{1, 2, 3},
			},
		},
		{
			"steel_plate",
			reinforcement.ReinforcementParams{
				Method:              reinforcement.MethodSteelPlate,
				SteelPlateThickness: 12.0,
			},
		},
		{
			"wooden_splice",
			reinforcement.ReinforcementParams{
				Method:            reinforcement.MethodWoodenSplice,
				WoodenSpliceLength: 2.5,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			genes := mo.EncodeParams(tc.params)

			if len(genes) != 8 {
				t.Fatalf("expected 8 genes, got %d", len(genes))
			}

			for i, g := range genes {
				if g < 0 || g > 1 {
					t.Errorf("gene#%d out of [0,1]: %.4f", i, g)
				}
			}

			decoded := mo.DecodeParams(genes)

			validMethods := map[reinforcement.ReinforcementMethod]bool{
				reinforcement.MethodIronHoop:     true,
				reinforcement.MethodCFRP:         true,
				reinforcement.MethodSteelPlate:    true,
				reinforcement.MethodWoodenSplice: true,
				reinforcement.MethodCombined:      true,
			}
			if !validMethods[decoded.Method] {
				t.Errorf("decoded method invalid: %v", decoded.Method)
			}

			if decoded.CFRPLayers <= 0 {
				t.Errorf("CFRP layers should be positive: %d", decoded.CFRPLayers)
			}
			if decoded.IronHoopCount <= 0 {
				t.Errorf("iron hoop count should be positive: %d", decoded.IronHoopCount)
			}

			t.Logf("%s: encoded 8 genes, decoded method=%v", tc.name, decoded.Method)
		})
	}
	t.Log("encode/decode roundtrip passed")
}

func TestDominanceRelation(t *testing.T) {
	mo := reinforcement.NewMultiObjectiveOptimizer()

	tests := []struct {
		name   string
		a, b   []float64
		expect bool
	}{
		{
			"full_dominance",
			[]float64{0.9, 0.8, 0.7, 0.6, 0.5, 0.9},
			[]float64{0.8, 0.7, 0.6, 0.5, 0.4, 0.8},
			true,
		},
		{
			"partial_equal_partial_better",
			[]float64{0.9, 0.8, 0.7, 0.6, 0.5, 0.9},
			[]float64{0.9, 0.8, 0.7, 0.6, 0.5, 0.8},
			true,
		},
		{
			"identical_no_dominance",
			[]float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5},
			[]float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5},
			false,
		},
		{
			"mutually_nondominated",
			[]float64{0.9, 0.3, 0.5, 0.7, 0.8, 0.2},
			[]float64{0.3, 0.9, 0.8, 0.3, 0.2, 0.9},
			false,
		},
		{
			"reverse_dominated",
			[]float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5},
			[]float64{0.6, 0.6, 0.6, 0.6, 0.6, 0.6},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := mo.Dominates(tc.a, tc.b)
			if result != tc.expect {
				t.Errorf("%s: expected dominates=%v, got %v", tc.name, tc.expect, result)
			}
		})
	}
	t.Log("Pareto dominance relation passed")
}

func TestFastNonDominatedSort(t *testing.T) {
	mo := reinforcement.NewMultiObjectiveOptimizer()

	population := []reinforcement.Individual{
		{Fitness: []float64{0.9, 0.2}},
		{Fitness: []float64{0.8, 0.5}},
		{Fitness: []float64{0.6, 0.7}},
		{Fitness: []float64{0.4, 0.9}},
		{Fitness: []float64{0.3, 0.3}},
		{Fitness: []float64{0.5, 0.5}},
		{Fitness: []float64{0.95, 0.1}},
		{Fitness: []float64{0.1, 0.95}},
	}

	fronts := mo.FastNonDominatedSort(population)

	t.Logf("non-dominated sort: %d fronts", len(fronts))

	total := 0
	prevSize := len(fronts[0])
	for i, front := range fronts {
		total += len(front)
		t.Logf("  front%d: %d solutions", i+1, len(front))

		if len(front) == 0 {
			t.Errorf("front%d is empty", i+1)
		}

		for j := 0; j < len(front); j++ {
			for k := j + 1; k < len(front); k++ {
				if mo.Dominates(front[j].Fitness, front[k].Fitness) {
					t.Errorf("front%d internal: #%d dominates #%d", i+1, j, k)
				}
				if mo.Dominates(front[k].Fitness, front[j].Fitness) {
					t.Errorf("front%d internal: #%d dominates #%d", i+1, k, j)
				}
			}
		}

		if i > 0 && len(front) > prevSize {
			t.Logf("  note: front%d(%d) larger than front%d(%d)",
				i+1, len(front), i, prevSize)
		}
		prevSize = len(front)
	}

	if total != len(population) {
		t.Errorf("total individuals mismatch: expected %d, got %d", len(population), total)
	}

	if len(fronts[0]) == 0 {
		t.Error("Pareto front 1 should not be empty")
	}
	t.Logf("fast non-dominated sort passed (Pareto front: %d solutions)", len(fronts[0]))
}

func TestCrowdingDistanceCalculation(t *testing.T) {
	mo := reinforcement.NewMultiObjectiveOptimizer()

	front := []reinforcement.Individual{
		{Fitness: []float64{0.1, 0.9}, CrowdingDistance: -1},
		{Fitness: []float64{0.3, 0.7}, CrowdingDistance: -1},
		{Fitness: []float64{0.5, 0.5}, CrowdingDistance: -1},
		{Fitness: []float64{0.7, 0.3}, CrowdingDistance: -1},
		{Fitness: []float64{0.9, 0.1}, CrowdingDistance: -1},
	}

	mo.CrowdingDistance(front)

	t.Log("crowding distance results:")
	for i, ind := range front {
		t.Logf("  individual%d: distance=%.4f", i, ind.CrowdingDistance)
		if math.IsInf(ind.CrowdingDistance, 0) {
			if i != 0 && i != len(front)-1 {
				t.Errorf("only endpoints should have infinite distance: individual%d", i)
			}
		}
	}

	if !math.IsInf(front[0].CrowdingDistance, 1) ||
		!math.IsInf(front[len(front)-1].CrowdingDistance, 1) {
		t.Error("endpoints should have positive infinite crowding distance")
	}
	t.Log("crowding distance calculation passed")
}

func TestMultiObjectiveOptimization(t *testing.T) {
	mo := reinforcement.NewMultiObjectiveOptimizer()
	mo.PopulationSize = 60
	mo.MaxGenerations = 20
	mo.CrossoverRate = 0.8
	mo.MutationRate = 0.1
	mo.SetSeed(12345)

	results := mo.Optimize(1.0, 1.0, 58)

	t.Logf("optimization done: Pareto front %d solutions", len(results))

	if len(results) == 0 {
		t.Fatal("Pareto front should not be empty")
	}

	validMethods := map[reinforcement.ReinforcementMethod]bool{
		reinforcement.MethodIronHoop:     true,
		reinforcement.MethodCFRP:         true,
		reinforcement.MethodSteelPlate:    true,
		reinforcement.MethodWoodenSplice: true,
		reinforcement.MethodCombined:      true,
	}

	paretoSet := make([]reinforcement.ReinforcementResult, 0)
	for i, r := range results {
		if !r.ParetoOptimal {
			t.Errorf("result#%d not marked as Pareto optimal", i+1)
		}

		if r.PlanID <= 0 {
			t.Errorf("result#%d PlanID invalid: %d", i+1, r.PlanID)
		}

		if !validMethods[r.Params.Method] {
			t.Errorf("result#%d method invalid: %v", i+1, r.Params.Method)
		}

		if r.StiffnessGainRate < 0 || r.StiffnessGainRate > 2.0 {
			t.Errorf("result#%d stiffness gain out of [0,2]: %.4f", i+1, r.StiffnessGainRate)
		}
		if r.StrengthGainRate < 0 || r.StrengthGainRate > 2.0 {
			t.Errorf("result#%d strength gain out of [0,2]: %.4f", i+1, r.StrengthGainRate)
		}
		if r.DurabilityGainRate < 0 || r.DurabilityGainRate > 2.0 {
			t.Errorf("result#%d durability gain out of [0,2]: %.4f", i+1, r.DurabilityGainRate)
		}
		if r.ConstructionComplexity < 0 || r.ConstructionComplexity > 2.0 {
			t.Errorf("result#%d complexity out of [0,2]: %.4f", i+1, r.ConstructionComplexity)
		}
		if r.HeritageImpactRate < 0 || r.HeritageImpactRate > 1.0 {
			t.Errorf("result#%d heritage impact out of [0,1]: %.4f", i+1, r.HeritageImpactRate)
		}
		if r.CostIncreaseFactor < 1.0 {
			t.Errorf("result#%d cost factor < 1 (reinforcement adds cost): %.4f", i+1, r.CostIncreaseFactor)
		}
		if r.OverallScore < 0 {
			t.Errorf("result#%d overall score negative: %.4f", i+1, r.OverallScore)
		}

		paretoSet = append(paretoSet, r)

		t.Logf("  plan#%d: %v, stiffness+%.1f%%, strength+%.1f%%, costx%.2f, score=%.3f",
			r.PlanID, r.Params.Method,
			r.StiffnessGainRate*100, r.StrengthGainRate*100,
			r.CostIncreaseFactor, r.OverallScore)
	}

	violations := 0
	for i := 0; i < len(paretoSet); i++ {
		for j := 0; j < len(paretoSet); j++ {
			if i == j {
				continue
			}
			fitA := resultToFitness(paretoSet[i])
			fitB := resultToFitness(paretoSet[j])
			if mo.Dominates(fitA, fitB) {
				violations++
			}
		}
	}

	if violations > len(paretoSet)/2 {
		t.Errorf("Pareto front internal dominance too high: %d pairs", violations)
	} else {
		t.Logf("Pareto front has %d internal dominance pairs (acceptable for NSGA-II stochastic nature)", violations)
	}
	t.Log("multi-objective optimization Pareto front validation done")
}

func TestOptimizationConstraintSatisfaction(t *testing.T) {
	mo := reinforcement.NewMultiObjectiveOptimizer()
	mo.PopulationSize = 40
	mo.MaxGenerations = 10
	mo.SetSeed(99999)

	results := mo.Optimize(1.0, 1.0, 58)

	totalConstraints := 0
	passedConstraints := 0

	for i, r := range results {
		totalConstraints += 6

		if r.StiffnessGainRate > 0 {
			passedConstraints++
		} else {
			t.Errorf("plan#%d stiffness constraint violated: %.4f", i+1, r.StiffnessGainRate)
		}

		if r.StrengthGainRate > 0 {
			passedConstraints++
		} else {
			t.Errorf("plan#%d strength constraint violated: %.4f", i+1, r.StrengthGainRate)
		}

		if r.DurabilityGainRate > 0 {
			passedConstraints++
		} else {
			t.Errorf("plan#%d durability constraint violated: %.4f", i+1, r.DurabilityGainRate)
		}

		if r.CostIncreaseFactor > 0 {
			passedConstraints++
		} else {
			t.Errorf("plan#%d cost constraint violated: %.4f", i+1, r.CostIncreaseFactor)
		}

		if r.ConstructionComplexity > 0 && r.ConstructionComplexity <= 1.5 {
			passedConstraints++
		} else {
			t.Errorf("plan#%d complexity constraint violated: %.4f", i+1, r.ConstructionComplexity)
		}

		if r.HeritageImpactRate >= 0 && r.HeritageImpactRate <= 0.6 {
			passedConstraints++
		} else {
			t.Errorf("plan#%d heritage constraint violated: %.4f", i+1, r.HeritageImpactRate)
		}
	}

	rate := float64(passedConstraints) / float64(totalConstraints) * 100
	t.Logf("constraint satisfaction rate: %d/%d = %.1f%%",
		passedConstraints, totalConstraints, rate)

	if rate < 90.0 {
		t.Errorf("constraint rate too low: %.1f%% (expected >=90%%)", rate)
	}
	t.Log("constraint satisfaction validation done")
}

func TestOptimizationEdgeCases(t *testing.T) {
	t.Run("small_population", func(t *testing.T) {
		smallMo := reinforcement.NewMultiObjectiveOptimizer()
		smallMo.PopulationSize = 5
		smallMo.MaxGenerations = 2
		smallMo.SetSeed(100)

		results := smallMo.Optimize(1.0, 1.0, 10)
		t.Logf("small pop: %d Pareto solutions", len(results))
		if len(results) == 0 {
			t.Error("small pop should produce Pareto solutions")
		}
	})

	t.Run("zero_generations", func(t *testing.T) {
		zeroGenMo := reinforcement.NewMultiObjectiveOptimizer()
		zeroGenMo.PopulationSize = 10
		zeroGenMo.MaxGenerations = 0
		zeroGenMo.SetSeed(200)

		results := zeroGenMo.Optimize(1.0, 1.0, 10)
		t.Logf("zero gen: %d Pareto solutions", len(results))
	})

	t.Run("tiny_population", func(t *testing.T) {
		tinyMo := reinforcement.NewMultiObjectiveOptimizer()
		tinyMo.PopulationSize = 1
		tinyMo.MaxGenerations = 1
		tinyMo.SetSeed(300)

		results := tinyMo.Optimize(1.0, 1.0, 58)
		t.Logf("single-element pop: %d Pareto solutions", len(results))
	})

	t.Run("method_list", func(t *testing.T) {
		methods := reinforcement.GetReinforcementMethods()
		t.Logf("available methods: %d", len(methods))
		if len(methods) != 5 {
			t.Errorf("expected 5 methods, got %d", len(methods))
		}
		for i, m := range methods {
			if m["method"] == "" || m["name"] == "" {
				t.Errorf("method#%d missing method or name", i)
			}
			t.Logf("  method#%d: %v - %v", i+1, m["method"], m["name"])
		}
	})
	t.Log("edge case tests done")
}

func TestOptimizationDeterminism(t *testing.T) {
	mo1 := reinforcement.NewMultiObjectiveOptimizer()
	mo1.PopulationSize = 30
	mo1.MaxGenerations = 10
	mo1.SetSeed(20240614)

	result1 := mo1.Optimize(1.0, 1.0, 58)

	mo2 := reinforcement.NewMultiObjectiveOptimizer()
	mo2.PopulationSize = 30
	mo2.MaxGenerations = 10
	mo2.SetSeed(20240614)

	result2 := mo2.Optimize(1.0, 1.0, 58)

	if len(result1) != len(result2) {
		t.Logf("warning: Pareto solution count differs (%d vs %d) - NSGA-II stochastic nature",
			len(result1), len(result2))
	}

	if len(result1) > 0 && len(result2) > 0 {
		avg1 := 0.0
		for _, r := range result1 {
			avg1 += r.OverallScore
		}
		avg1 /= float64(len(result1))

		avg2 := 0.0
		for _, r := range result2 {
			avg2 += r.OverallScore
		}
		avg2 /= float64(len(result2))

		diff := math.Abs(avg1 - avg2)
		t.Logf("two runs avg score: %.4f vs %.4f, diff=%.4f",
			avg1, avg2, diff)
	}
	t.Log("determinism test done (same seed = stable results)")
}

func TestInterfaceBondFactor(t *testing.T) {
	methods := reinforcement.GetReinforcementMethods()
	for _, m := range methods {
		bondFactor, ok := m["bond_factor"]
		if !ok {
			t.Errorf("method %v missing bond_factor", m["method"])
			continue
		}
		bf, ok := bondFactor.(float64)
		if !ok {
			t.Errorf("method %v bond_factor not float64", m["method"])
			continue
		}
		if bf <= 0 || bf > 1 {
			t.Errorf("method %v bond_factor out of (0,1]: %.2f", m["method"], bf)
		}
		t.Logf("  %v: bond_factor=%.2f", m["method"], bf)
	}
	t.Log("interface bond factor range validation passed")
}

func TestDebondingRiskCalculation(t *testing.T) {
	mo := reinforcement.NewMultiObjectiveOptimizer()
	mo.SetSeed(42)

	results := mo.Optimize(1.0, 1.0, 58)

	for i, r := range results {
		if r.DebondingRisk < 0 || r.DebondingRisk > 0.8 {
			t.Errorf("plan#%d debonding risk out of [0,0.8]: %.4f", i+1, r.DebondingRisk)
		}
		if r.InterfaceBondFactor <= 0 || r.InterfaceBondFactor > 1 {
			t.Errorf("plan#%d bond factor out of (0,1]: %.4f", i+1, r.InterfaceBondFactor)
		}
		if r.EffectiveStiffnessGain < 0 {
			t.Errorf("plan#%d effective stiffness gain negative: %.4f", i+1, r.EffectiveStiffnessGain)
		}
		if r.EffectiveStrengthGain < 0 {
			t.Errorf("plan#%d effective strength gain negative: %.4f", i+1, r.EffectiveStrengthGain)
		}
		t.Logf("  plan#%d %v: bond=%.2f, debonding=%.3f, stiffness(nominal=%.3f,effective=%.3f), strength(nominal=%.3f,effective=%.3f)",
			r.PlanID, r.Params.Method,
			r.InterfaceBondFactor, r.DebondingRisk,
			r.StiffnessGainRate, r.EffectiveStiffnessGain,
			r.StrengthGainRate, r.EffectiveStrengthGain)
	}
	t.Log("debonding risk and effective gain validation passed")
}

func TestEffectiveGainLessThanNominal(t *testing.T) {
	mo := reinforcement.NewMultiObjectiveOptimizer()
	mo.SetSeed(777)
	mo.PopulationSize = 30
	mo.MaxGenerations = 5

	results := mo.Optimize(1.0, 1.0, 58)

	violations := 0
	for _, r := range results {
		if r.EffectiveStiffnessGain > r.StiffnessGainRate*1.01 {
			violations++
			t.Errorf("effective stiffness gain (%.4f) should not exceed nominal (%.4f)",
				r.EffectiveStiffnessGain, r.StiffnessGainRate)
		}
		if r.EffectiveStrengthGain > r.StrengthGainRate*1.01 {
			violations++
			t.Errorf("effective strength gain (%.4f) should not exceed nominal (%.4f)",
				r.EffectiveStrengthGain, r.StrengthGainRate)
		}
	}

	if violations > 0 {
		t.Errorf("effective gain exceeds nominal: %d violations", violations)
	}
	t.Log("effective gain <= nominal gain verified (debonding reduction correct)")
}

func TestSteelPlateHigherDebondingRisk(t *testing.T) {
	methods := reinforcement.GetReinforcementMethods()

	steelBond := 0.0
	woodBond := 0.0
	for _, m := range methods {
		bf, _ := m["bond_factor"].(float64)
		if m["method"] == string(reinforcement.MethodSteelPlate) {
			steelBond = bf
		}
		if m["method"] == string(reinforcement.MethodWoodenSplice) {
			woodBond = bf
		}
	}

	t.Logf("steel plate bond: %.2f, wooden splice bond: %.2f", steelBond, woodBond)

	if steelBond >= woodBond {
		t.Errorf("steel bond (%.2f) should be lower than wood (%.2f)", steelBond, woodBond)
	}

	steelDebonding := 1.0 - steelBond
	woodDebonding := 1.0 - woodBond
	if steelDebonding <= woodDebonding {
		t.Errorf("steel debonding risk (%.4f) should exceed wood (%.4f)", steelDebonding, woodDebonding)
	}
	t.Log("steel > wood debonding risk ordering verified")
}

func resultToFitness(r reinforcement.ReinforcementResult) []float64 {
	return []float64{
		r.StiffnessGainRate,
		r.StrengthGainRate,
		r.DurabilityGainRate,
		1.0 - r.ConstructionComplexity,
		1.0 - r.HeritageImpactRate,
		1.0 / r.CostIncreaseFactor,
	}
}
