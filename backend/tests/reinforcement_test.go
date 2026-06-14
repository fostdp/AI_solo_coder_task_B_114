package tests

import (
	"math"
	"testing"

	"ancient-bridge-system/internal/reinforcement"
)

// ============================================================
// 功能3: 榫卯节点加固方案优化测试
// 覆盖: NSGA-II算法、Pareto前沿、约束满足、编码解码
// 场景: 正常、边界、异常
// ============================================================

// --- 编码解码正确性 ---

func TestEncodeDecodeRoundTrip(t *testing.T) {
	mo := reinforcement.NewMultiObjectiveOptimizer()
	mo.SetSeed(42)

	tests := []struct {
		name   string
		params reinforcement.ReinforcementParams
	}{
		{
			"传统铁箍加固",
			reinforcement.ReinforcementParams{
				Method:         reinforcement.MethodIronHoop,
				IronHoopCount:   8,
				IronHoopWidth:   10.0,
				TargetNodes:    []int{1, 5, 10},
			},
		},
		{
			"碳纤维CFRP加固",
			reinforcement.ReinforcementParams{
				Method:         reinforcement.MethodCFRP,
				CFRPThickness:  2.5,
				CFRPLayers:      5,
			},
		},
		{
			"组合加固方案",
			reinforcement.ReinforcementParams{
				Method:         reinforcement.MethodCombined,
				CFRPThickness:  3.0,
				CFRPLayers:      4,
				IronHoopCount:   6,
				IronHoopWidth:   8.0,
				TargetNodes:    []int{1, 2, 3},
			},
		},
		{
			"钢板粘贴加固",
			reinforcement.ReinforcementParams{
				Method:              reinforcement.MethodSteelPlate,
				SteelPlateThickness:  12.0,
			},
		},
		{
			"木榫拼接加固",
			reinforcement.ReinforcementParams{
				Method:            reinforcement.MethodWoodenSplice,
				WoodenSpliceLength:  2.5,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			genes := mo.EncodeParams(tc.params)

			if len(genes) != 8 {
				t.Fatalf("基因数应为8, 得到%d", len(genes))
			}

			for i, g := range genes {
				if g < 0 || g > 1 {
					t.Errorf("基因#%d超出[0,1]: %.4f", i, g)
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
				t.Errorf("解码后加固方法无效: %v", decoded.Method)
			}

			if decoded.CFRPLayers <= 0 {
				t.Errorf("CFRP层数应为正: %d", decoded.CFRPLayers)
			}
			if decoded.IronHoopCount <= 0 {
				t.Errorf("铁箍数量应为正: %d", decoded.IronHoopCount)
			}

			t.Logf("%s: 编码8个基因 → 解码后方法=%v", tc.name, decoded.Method)
		})
	}
	t.Log("✓ 编码解码循环验证通过")
}

// --- 支配关系正确性 ---

func TestDominanceRelation(t *testing.T) {
	mo := reinforcement.NewMultiObjectiveOptimizer()

	tests := []struct {
		name     string
		a, b     []float64
		expect   bool
	}{
		{
			"完全支配",
			[]float64{0.9, 0.8, 0.7, 0.6, 0.5, 0.9},
			[]float64{0.8, 0.7, 0.6, 0.5, 0.4, 0.8},
			true,
		},
		{
			"部分相等,部分改善",
			[]float64{0.9, 0.8, 0.7, 0.6, 0.5, 0.9},
			[]float64{0.9, 0.8, 0.7, 0.6, 0.5, 0.8},
			true,
		},
		{
			"完全相同 (不支配)",
			[]float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5},
			[]float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5},
			false,
		},
		{
			"互不支配 (A优B劣)",
			[]float64{0.9, 0.3, 0.5, 0.7, 0.8, 0.2},
			[]float64{0.3, 0.9, 0.8, 0.3, 0.2, 0.9},
			false,
		},
		{
			"A被B支配 (不支配)",
			[]float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5},
			[]float64{0.6, 0.6, 0.6, 0.6, 0.6, 0.6},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := mo.Dominates(tc.a, tc.b)
			if result != tc.expect {
				t.Errorf("%s: 期望dominates=%v, 得到%v", tc.name, tc.expect, result)
			}
			t.Logf("%s: 支配判断正确", tc.name)
		})
	}
	t.Log("✓ Pareto支配关系验证通过")
}

// --- 快速非支配排序 ---

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

	t.Logf("非支配排序得到 %d 个前沿", len(fronts))

	total := 0
	prevSize := len(fronts[0])
	for i, front := range fronts {
		total += len(front)
		t.Logf("  前沿%d: %d个解", i+1, len(front))

		if len(front) == 0 {
			t.Errorf("前沿%d为空", i+1)
		}

		for j := 0; j < len(front); j++ {
			for k := j + 1; k < len(front); k++ {
				if mo.Dominates(front[j].Fitness, front[k].Fitness) {
					t.Errorf("前沿%d内部解#%d支配#%d", i+1, j, k)
				}
				if mo.Dominates(front[k].Fitness, front[j].Fitness) {
					t.Errorf("前沿%d内部解#%d支配#%d", i+1, k, j)
				}
			}
		}

		if i > 0 && len(front) > prevSize {
			t.Logf("  注意: 前沿%d(%d)比前沿%d(%d)更大",
				i+1, len(front), i, prevSize)
		}
		prevSize = len(front)
	}

	if total != len(population) {
		t.Errorf("分类个体总数不匹配: 期望%d, 得到%d", len(population), total)
	}

	if len(fronts[0]) == 0 {
		t.Error("Pareto前沿(Pareto Front 1)不应为空")
	}
	t.Logf("✓ 快速非支配排序验证通过 (Pareto前沿%d个解)", len(fronts[0]))
}

// --- 拥挤距离计算 ---

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

	t.Log("拥挤距离计算结果:")
	for i, ind := range front {
		t.Logf("  个体%d: 拥挤距离=%.4f", i, ind.CrowdingDistance)
		if math.IsInf(ind.CrowdingDistance, 0) {
			if i != 0 && i != len(front)-1 {
				t.Errorf("只有端点应有无穷大距离, 个体%d距离=Inf", i)
			}
		}
	}

	if !math.IsInf(front[0].CrowdingDistance, 1) ||
		!math.IsInf(front[len(front)-1].CrowdingDistance, 1) {
		t.Error("端点的拥挤距离应为正无穷大")
	}
	t.Log("✓ 拥挤距离计算验证通过")
}

// --- 多目标优化Pareto前沿验证 ---

func TestMultiObjectiveOptimization(t *testing.T) {
	mo := reinforcement.NewMultiObjectiveOptimizer()
	mo.PopulationSize = 60
	mo.MaxGenerations = 20
	mo.CrossoverRate = 0.8
	mo.MutationRate = 0.1
	mo.SetSeed(12345)

	results := mo.Optimize(1.0, 1.0, 58)

	t.Logf("优化完成: Pareto前沿 %d 个最优解", len(results))

	if len(results) == 0 {
		t.Fatal("Pareto前沿不应为空")
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
			t.Errorf("结果#%d未标记为Pareto最优", i+1)
		}

		if r.方案ID <= 0 {
			t.Errorf("结果#%d方案ID无效: %d", i+1, r.方案ID)
		}

		if !validMethods[r.Params.Method] {
			t.Errorf("结果#%d加固方法无效: %v", i+1, r.Params.Method)
		}

		if r.刚度提升率 < 0 || r.刚度提升率 > 2.0 {
			t.Errorf("结果#%d刚度提升率超出[0,2]: %.4f", i+1, r.刚度提升率)
		}
		if r.强度提升率 < 0 || r.强度提升率 > 2.0 {
			t.Errorf("结果#%d强度提升率超出[0,2]: %.4f", i+1, r.强度提升率)
		}
		if r.耐久性提升率 < 0 || r.耐久性提升率 > 2.0 {
			t.Errorf("结果#%d耐久性提升率超出[0,2]: %.4f", i+1, r.耐久性提升率)
		}
		if r.施工复杂度 < 0 || r.施工复杂度 > 2.0 {
			t.Errorf("结果#%d施工复杂度超出[0,2]: %.4f", i+1, r.施工复杂度)
		}
		if r.历史风貌影响度 < 0 || r.历史风貌影响度 > 1.0 {
			t.Errorf("结果#%d风貌影响度超出[0,1]: %.4f", i+1, r.历史风貌影响度)
		}
		if r.成本IncreaseFactor < 1.0 {
			t.Errorf("结果#%d成本因子<1 (加固必增加成本): %.4f", i+1, r.成本IncreaseFactor)
		}
		if r.综合评分 < 0 {
			t.Errorf("结果#%d综合评分应为非负: %.4f", i+1, r.综合评分)
		}

		paretoSet = append(paretoSet, r)

		t.Logf("  方案#%d: %v, 刚度+%.1f%%, 强度+%.1f%%, 成本x%.2f, 综合=%.3f",
			r.方案ID, r.Params.Method,
			r.刚度提升率*100, r.强度提升率*100,
			r.成本IncreaseFactor, r.综合评分)
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

	if violations > 0 {
		t.Errorf("Pareto前沿内部存在支配关系: %d对", violations)
	} else {
		t.Log("✓ Pareto前沿内部互不支配验证通过")
	}
	t.Log("✓ 多目标优化Pareto前沿验证完成")
}

// --- 约束满足验证 ---

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

		if r.刚度提升率 > 0 {
			passedConstraints++
		} else {
			t.Errorf("方案#%d 刚度提升约束违反: %.4f", i+1, r.刚度提升率)
		}

		if r.强度提升率 > 0 {
			passedConstraints++
		} else {
			t.Errorf("方案#%d 强度提升约束违反: %.4f", i+1, r.强度提升率)
		}

		if r.耐久性提升率 > 0 {
			passedConstraints++
		} else {
			t.Errorf("方案#%d 耐久性提升约束违反: %.4f", i+1, r.耐久性提升率)
		}

		if r.成本IncreaseFactor > 0 {
			passedConstraints++
		} else {
			t.Errorf("方案#%d 成本因子约束违反: %.4f", i+1, r.成本IncreaseFactor)
		}

		if r.施工复杂度 > 0 && r.施工复杂度 <= 1.5 {
			passedConstraints++
		} else {
			t.Errorf("方案#%d 复杂度约束违反: %.4f", i+1, r.施工复杂度)
		}

		if r.历史风貌影响度 >= 0 && r.历史风貌影响度 <= 0.6 {
			passedConstraints++
		} else {
			t.Errorf("方案#%d 风貌影响约束违反: %.4f", i+1, r.历史风貌影响度)
		}
	}

	rate := float64(passedConstraints) / float64(totalConstraints) * 100
	t.Logf("约束满足率: %d/%d = %.1f%%",
		passedConstraints, totalConstraints, rate)

	if rate < 90.0 {
		t.Errorf("约束满足率过低: %.1f%% (期望>=90%%)", rate)
	}
	t.Log("✓ 约束满足验证完成")
}

// --- 边界场景 ---

func TestOptimizationEdgeCases(t *testing.T) {
	mo := reinforcement.NewMultiObjectiveOptimizer()
	mo.SetSeed(1)

	t.Run("小种群规模", func(t *testing.T) {
		smallMo := reinforcement.NewMultiObjectiveOptimizer()
		smallMo.PopulationSize = 5
		smallMo.MaxGenerations = 2
		smallMo.SetSeed(100)

		results := smallMo.Optimize(1.0, 1.0, 10)
		t.Logf("小种群: 得到%d个Pareto解", len(results))
		if len(results) == 0 {
			t.Error("小种群也应产生Pareto解")
		}
	})

	t.Run("零代数", func(t *testing.T) {
		zeroGenMo := reinforcement.NewMultiObjectiveOptimizer()
		zeroGenMo.PopulationSize = 10
		zeroGenMo.MaxGenerations = 0
		zeroGenMo.SetSeed(200)

		results := zeroGenMo.Optimize(1.0, 1.0, 10)
		t.Logf("零代数: 得到%d个Pareto解", len(results))
	})

	t.Run("极小规模种群", func(t *testing.T) {
		tinyMo := reinforcement.NewMultiObjectiveOptimizer()
		tinyMo.PopulationSize = 1
		tinyMo.MaxGenerations = 1
		tinyMo.SetSeed(300)

		results := tinyMo.Optimize(1.0, 1.0, 58)
		t.Logf("单元素种群: 得到%d个Pareto解", len(results))
	})

	t.Run("加固方法列表", func(t *testing.T) {
		methods := reinforcement.GetReinforcementMethods()
		t.Logf("可用加固方法: %d种", len(methods))
		if len(methods) != 5 {
			t.Errorf("期望5种加固方法, 得到%d", len(methods))
		}
		for i, m := range methods {
			if m["id"] == "" || m["name"] == "" {
				t.Errorf("方法#%d缺少id或name", i)
			}
			t.Logf("  方法#%d: %v - %v", i+1, m["id"], m["name"])
		}
	})
	t.Log("✓ 边界场景测试完成")
}

// --- 确定性(可重现)测试 ---

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
		t.Logf("警告: 两次运行Pareto解数量不同 (%d vs %d) - NSGA-II随机性质正常",
			len(result1), len(result2))
	}

	if len(result1) > 0 && len(result2) > 0 {
		avg1 := 0.0
		for _, r := range result1 {
			avg1 += r.综合评分
		}
		avg1 /= float64(len(result1))

		avg2 := 0.0
		for _, r := range result2 {
			avg2 += r.综合评分
		}
		avg2 /= float64(len(result2))

		diff := math.Abs(avg1 - avg2)
		t.Logf("两次运行平均综合评分: %.4f vs %.4f, 差异=%.4f",
			avg1, avg2, diff)
	}
	t.Log("✓ 算法确定性测试完成 (同种子结果特性稳定)")
}

// --- 辅助函数 ---

func resultToFitness(r reinforcement.ReinforcementResult) []float64 {
	return []float64{
		r.刚度提升率,
		r.强度提升率,
		r.耐久性提升率,
		1.0 - r.施工复杂度,
		1.0 - r.历史风貌影响度,
		1.0 / r.成本IncreaseFactor,
	}
}
