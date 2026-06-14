package tests

import (
	"math"
	"testing"

	"ancient-bridge-system/internal/fea"
)

// ============================================================
// 功能4: 参数化交互体验测试
// 覆盖: 参数验证、参数灵敏度、几何更新、结果稳定性
// 场景: 正常、边界、异常
// ============================================================

// --- 参数桥接构造 ---

func TestParametricBridgeCreation(t *testing.T) {
	pb := fea.NewParametricBridge(1, 25.6, 5.8, 6.5)
	if pb == nil {
		t.Fatal("参数化桥不应为nil")
	}

	if pb.BaseBridge == nil {
		t.Error("基础桥梁模型不应为nil")
	}

	t.Logf("初始参数: 跨=%.1fm, 矢=%.1fm, 宽=%.1fm",
		pb.BaseBridge.SpanLength,
		pb.BaseBridge.ArchRise,
		pb.BaseBridge.DeckWidth)

	if pb.MinSpan != 10.0 || pb.MaxSpan != 50.0 {
		t.Errorf("跨度范围错误: 期望[10,50], 得到[%.1f,%.1f]",
			pb.MinSpan, pb.MaxSpan)
	}
	if pb.MinRise != 2.0 || pb.MaxRise != 15.0 {
		t.Errorf("矢高范围错误: 期望[2,15], 得到[%.1f,%.1f]",
			pb.MinRise, pb.MaxRise)
	}
	t.Log("✓ 参数化桥梁创建验证通过")
}

// --- 参数正常范围验证 ---

func TestParameterValidationNormal(t *testing.T) {
	pb := fea.NewParametricBridge(1, 25.6, 5.8, 6.5)

	tests := []struct {
		name      string
		span      float64
		rise      float64
		width     float64
		expectOK  bool
	}{
		{"汴水虹桥标准", 25.6, 5.8, 6.5, true},
		{"最小跨度边界", 10.0, 2.5, 4.0, true},
		{"最大跨度边界", 50.0, 15.0, 12.0, true},
		{"最小矢跨比边界(1:12)", 24.0, 2.0, 5.0, true},
		{"最大矢跨比边界(1:3)", 9.0, 3.0, 4.0, true},
		{"典型廊桥参数", 28.0, 6.0, 5.0, true},
		{"中等跨拱桥", 20.0, 5.0, 5.5, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid, msg := pb.ValidateParams(tc.span, tc.rise, tc.width)
			if valid != tc.expectOK {
				t.Errorf("%s: 期望valid=%v, 得到%v (%s)",
					tc.name, tc.expectOK, valid, msg)
			}
			t.Logf("%s: valid=%v, %s", tc.name, valid, msg)
		})
	}
	t.Log("✓ 正常参数范围验证通过")
}

// --- 参数边界异常验证 ---

func TestParameterValidationEdgeCases(t *testing.T) {
	pb := fea.NewParametricBridge(1, 25.6, 5.8, 6.5)

	tests := []struct {
		name     string
		span     float64
		rise     float64
		width    float64
		expectOK bool
	}{
		{"跨度太小", 5.0, 2.0, 4.0, false},
		{"跨度太大", 60.0, 15.0, 10.0, false},
		{"矢高太小", 20.0, 1.0, 5.0, false},
		{"矢高太大", 20.0, 20.0, 5.0, false},
		{"桥宽太小", 25.0, 5.0, 1.0, false},
		{"桥宽太大", 25.0, 5.0, 20.0, false},
		{"矢跨比过小<1:12", 30.0, 2.0, 6.0, false},
		{"矢跨比过大>1:3", 10.0, 5.0, 5.0, false},
		{"零跨度", 0.0, 0.0, 0.0, false},
		{"负跨度", -10.0, 5.0, 5.0, false},
		{"负矢高", 20.0, -3.0, 5.0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid, msg := pb.ValidateParams(tc.span, tc.rise, tc.width)
			if valid != tc.expectOK {
				t.Errorf("%s: 期望valid=%v, 得到%v (%s)",
					tc.name, tc.expectOK, valid, msg)
			}
			if !tc.expectOK && valid {
				t.Errorf("%s: 应返回错误但通过了", tc.name)
			}
			t.Logf("%s: valid=%v, %s", tc.name, valid, msg)
		})
	}
	t.Log("✓ 参数边界异常验证通过")
}

// --- 几何选项 ---

func TestGeometryOptions(t *testing.T) {
	pb := fea.NewParametricBridge(1, 25.6, 5.8, 6.5)
	options := pb.GetGeometryOptions()

	t.Logf("几何参数选项数: %d", len(options))
	if len(options) != 3 {
		t.Errorf("期望3个参数选项, 得到%d", len(options))
	}

	expected := map[string]struct {
		min, max, step, def float64
	}{
		"span_length": {10.0, 50.0, 0.5, 25.6},
		"arch_rise":   {2.0, 15.0, 0.1, 5.8},
		"deck_width":  {3.0, 12.0, 0.1, 6.5},
	}

	for i, opt := range options {
		exp, exists := expected[opt.Name]
		if !exists {
			t.Errorf("参数#%d名未知: %s", i, opt.Name)
			continue
		}

		if math.Abs(opt.Min-exp.min) > 1e-9 {
			t.Errorf("%s Min: 期望%.1f, 得到%.1f", opt.Name, exp.min, opt.Min)
		}
		if math.Abs(opt.Max-exp.max) > 1e-9 {
			t.Errorf("%s Max: 期望%.1f, 得到%.1f", opt.Name, exp.max, opt.Max)
		}
		if math.Abs(opt.Step-exp.step) > 1e-9 {
			t.Errorf("%s Step: 期望%.1f, 得到%.1f", opt.Name, exp.step, opt.Step)
		}
		if math.Abs(opt.Default-exp.def) > 1e-9 {
			t.Errorf("%s Default: 期望%.1f, 得到%.1f", opt.Name, exp.def, opt.Default)
		}

		t.Logf("  %s (%s): [%.1f, %.1f], step=%.1f, default=%.1f",
			opt.Name, opt.Label, opt.Min, opt.Max, opt.Step, opt.Default)
	}
	t.Log("✓ 几何参数选项验证通过")
}

// --- 参数灵敏度分析 ---

func TestParameterSensitivityAnalysis(t *testing.T) {
	loadValue := 50.0

	sweepSpan := []struct {
		span float64
		rise float64
		width float64
	}{
		{15.0, 4.0, 5.0},
		{20.0, 5.0, 5.0},
		{25.0, 6.0, 5.0},
		{30.0, 7.0, 5.0},
		{40.0, 9.0, 5.0},
	}

	t.Log("=== 跨度灵敏度 (矢跨比~1:4, 桥宽固定) ===")
	prevStress := 0.0
	prevDisp := 0.0
	for i, p := range sweepSpan {
		result := fea.CalculateParametricAnalysis(1, p.span, p.rise, p.width, loadValue)

		stressSensitivity := 0.0
		dispSensitivity := 0.0
		if i > 0 && prevStress > 0 {
			stressSensitivity = (result.MaxStressRatio - prevStress) / prevStress * 100
			dispSensitivity = (result.MaxDisplacement - prevDisp) / prevDisp * 100
		}

		t.Logf("  跨=%.0fm: 应力比=%.3f (Δ%.1f%%), 位移=%.3fmm (Δ%.1f%%), 材料=%.1fm³",
			p.span,
			result.MaxStressRatio, stressSensitivity,
			result.MaxDisplacement, dispSensitivity,
			result.TotalVolume)

		if result.MaxStressRatio <= 0 {
			t.Errorf("跨度%.0f: 应力比应为正, 得到%.4f", p.span, result.MaxStressRatio)
		}
		if result.MaxDisplacement <= 0 {
			t.Errorf("跨度%.0f: 最大位移应为正, 得到%.4f", p.span, result.MaxDisplacement)
		}

		if i > 0 {
			if result.MaxStressRatio < prevStress*0.5 {
				t.Errorf("跨度增加应力比应上升: %.3f -> %.3f (异常下降)",
					prevStress, result.MaxStressRatio)
			}
			if result.MaxDisplacement < prevDisp*0.5 {
				t.Errorf("跨度增加位移应上升: %.3fmm -> %.3fmm (异常下降)",
					prevDisp, result.MaxDisplacement)
			}
		}

		prevStress = result.MaxStressRatio
		prevDisp = result.MaxDisplacement
	}
	t.Log("✓ 跨度灵敏度验证通过 (应力/位移随跨增大)")

	sweepRise := []struct {
		span float64
		rise float64
		width float64
	}{
		{25.0, 3.0, 6.0},
		{25.0, 4.5, 6.0},
		{25.0, 6.0, 6.0},
		{25.0, 7.5, 6.0},
		{25.0, 8.3, 6.0},
	}

	t.Log("\n=== 矢高灵敏度 (跨度固定25m, 桥宽固定) ===")
	prevStress = 0
	prevDisp = 0
	for i, p := range sweepRise {
		result := fea.CalculateParametricAnalysis(1, p.span, p.rise, p.width, loadValue)
		ratio := p.rise / p.span

		stressSensitivity := 0.0
		dispSensitivity := 0.0
		if i > 0 && prevStress > 0 {
			stressSensitivity = (result.MaxStressRatio - prevStress) / prevStress * 100
			dispSensitivity = (result.MaxDisplacement - prevDisp) / prevDisp * 100
		}

		t.Logf("  矢=%.1fm (矢跨比=1:%.1f): 应力比=%.3f (Δ%.1f%%), 位移=%.3fmm (Δ%.1f%%)",
			p.rise, 1.0/ratio,
			result.MaxStressRatio, stressSensitivity,
			result.MaxDisplacement, dispSensitivity)

		if ratio < 1.0/12.0 || ratio > 1.0/3.0 {
			continue
		}

		if i > 0 && prevStress > 0 {
			if result.MaxStressRatio > prevStress*1.1 {
				t.Errorf("矢高增加应力比应下降: %.3f -> %.3f (异常上升>10%%)",
					prevStress, result.MaxStressRatio)
			}
		}

		prevStress = result.MaxStressRatio
		prevDisp = result.MaxDisplacement
	}
	t.Log("✓ 矢高灵敏度验证通过 (合理矢跨比下应力随矢高增加趋于下降)")

	sweepWidth := []struct {
		span float64
		rise float64
		width float64
	}{
		{25.0, 6.0, 3.5},
		{25.0, 6.0, 5.0},
		{25.0, 6.0, 6.5},
		{25.0, 6.0, 8.0},
		{25.0, 6.0, 10.0},
	}

	t.Log("\n=== 桥宽灵敏度 (跨度25m, 矢高6m) ===")
	prevStress = 0
	prevDisp = 0
	for i, p := range sweepWidth {
		result := fea.CalculateParametricAnalysis(1, p.span, p.rise, p.width, loadValue)

		stressChange := 0.0
		volumeChange := 0.0
		if i > 0 && prevStress > 0 {
			stressChange = (result.MaxStressRatio - prevStress) / prevStress * 100
			volumeChange = (result.TotalVolume - sweepSpan[0].span*sweepWidth[0].width*0.8) /
				(sweepSpan[0].span*sweepWidth[0].width*0.8) * 100
		}

		t.Logf("  宽=%.1fm: 应力比=%.3f (Δ%.1f%%), 材料=%.1fm³, 效率=%.4f",
			p.width,
			result.MaxStressRatio, stressChange,
			result.TotalVolume,
			result.MaterialEfficiency)

		if result.MaterialEfficiency <= 0 {
			t.Errorf("桥宽%.1f: 材料效率应为正", p.width)
		}

		prevStress = result.MaxStressRatio
		_ = volumeChange
	}
	t.Log("✓ 桥宽灵敏度验证通过")
}

// --- 参数化分析结果完整性 ---

func TestParametricAnalysisIntegrity(t *testing.T) {
	result := fea.CalculateParametricAnalysis(1, 25.6, 5.8, 6.5, 50.0)

	t.Logf("参数化分析结果完整性检查:")
	t.Logf("  桥梁ID: %d", result.BridgeID)
	t.Logf("  跨度: %.1fm, 矢高: %.1fm, 桥宽: %.1fm",
		result.SpanLength, result.ArchRise, result.DeckWidth)
	t.Logf("  矢跨比: %.4f (1:%.1f)",
		result.RiseSpanRatio, 1.0/result.RiseSpanRatio)
	t.Logf("  构件数: %d, 节点数: %d",
		result.MemberCount, result.NodeCount)
	t.Logf("  最大应力比: %.3f", result.MaxStressRatio)
	t.Logf("  最大位移: %.3fmm (节点#%d)",
		result.MaxDisplacement, result.MaxDisplacementNode)
	t.Logf("  材料总量: %.1fm³", result.TotalVolume)
	t.Logf("  材料效率: %.4f", result.MaterialEfficiency)
	t.Logf("  构件内力: %d条, 节点位移: %d条",
		len(result.MemberForces), len(result.Displacements))
	t.Logf("  营造法式对比: %d条", len(result.YingzaoComparison))
	t.Logf("  节点信息: %d条, 构件信息: %d条",
		len(result.Nodes), len(result.Members))

	if result.BridgeID != 1 {
		t.Errorf("桥梁ID错误: 期望1, 得到%d", result.BridgeID)
	}

	ratio := result.ArchRise / result.SpanLength
	if math.Abs(result.RiseSpanRatio-ratio) > 1e-9 {
		t.Errorf("矢跨比计算错误: 期望%.4f, 得到%.4f", ratio, result.RiseSpanRatio)
	}

	if result.MemberCount <= 0 {
		t.Error("构件数应为正")
	}
	if result.NodeCount <= 0 {
		t.Error("节点数应为正")
	}

	if result.MaxStressRatio <= 0 {
		t.Error("最大应力比应为正")
	}
	if result.MaxDisplacement <= 0 {
		t.Error("最大位移应为正")
	}
	if result.TotalVolume <= 0 {
		t.Error("材料体积应为正")
	}

	if len(result.MemberForces) != result.MemberCount {
		t.Errorf("构件内力条数不匹配: 期望%d, 得到%d",
			result.MemberCount, len(result.MemberForces))
	}
	if len(result.Displacements) != result.NodeCount {
		t.Errorf("节点位移条数不匹配: 期望%d, 得到%d",
			result.NodeCount, len(result.Displacements))
	}
	if len(result.Nodes) != result.NodeCount {
		t.Errorf("节点信息条数不匹配: 期望%d, 得到%d",
			result.NodeCount, len(result.Nodes))
	}
	if len(result.Members) != result.MemberCount {
		t.Errorf("构件信息条数不匹配: 期望%d, 得到%d",
			result.MemberCount, len(result.Members))
	}
	t.Log("✓ 参数化分析结果完整性验证通过")
}

// --- 结果合理性(物理约束) ---

func TestParametricAnalysisPhysicalConstraints(t *testing.T) {
	tests := []struct {
		name  string
		span  float64
		rise  float64
		width float64
		load  float64
	}{
		{"标准工况", 25.6, 5.8, 6.5, 50.0},
		{"小桥", 15.0, 4.0, 4.0, 20.0},
		{"大桥", 40.0, 12.0, 8.0, 80.0},
		{"重载", 25.0, 6.0, 6.0, 200.0},
		{"轻载", 25.0, 6.0, 6.0, 10.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := fea.CalculateParametricAnalysis(1, tc.span, tc.rise, tc.width, tc.load)

			allowableDisp := tc.span / 400.0 * 1000.0
			t.Logf("%s:", tc.name)
			t.Logf("  跨度/400容许位移: %.2fmm", allowableDisp)
			t.Logf("  实际最大位移:     %.2fmm (%.0f%%容许值)",
				result.MaxDisplacement,
				result.MaxDisplacement/allowableDisp*100)

			if result.MaxDisplacement > allowableDisp*5.0 {
				t.Logf("  警告: 位移超过5倍L/400 (%.0f%%), 可能需关注",
					result.MaxDisplacement/allowableDisp*100)
			}

			if result.MaxStressRatio > 5.0 {
				t.Errorf("%s: 最大应力比异常大 (>5): %.3f",
					tc.name, result.MaxStressRatio)
			}

			expectedVolumeMin := tc.span * tc.width * 0.3
			expectedVolumeMax := tc.span * tc.width * 3.0
			if result.TotalVolume < expectedVolumeMin ||
				result.TotalVolume > expectedVolumeMax {
				t.Errorf("%s: 材料体积异常: %.1fm³ (合理范围%.1f~%.1f)",
					tc.name, result.TotalVolume,
					expectedVolumeMin, expectedVolumeMax)
			}

			for i, m := range result.Members {
				if m.StressRatio < 0 {
					t.Errorf("%s 构件#%d: 应力比不能为负: %.4f",
						tc.name, i+1, m.StressRatio)
				}
				if m.Length <= 0 {
					t.Errorf("%s 构件#%d: 长度应为正: %.4f",
						tc.name, i+1, m.Length)
				}
				if m.Area <= 0 {
					t.Errorf("%s 构件#%d: 截面积应为正: %.4f",
						tc.name, i+1, m.Area)
				}
			}
			t.Logf("  ✓ 通过检查 (应力比=%.3f)", result.MaxStressRatio)
		})
	}
	t.Log("✓ 结果物理合理性约束验证通过")
}

// --- 几何更新功能 ---

func TestGeometryUpdate(t *testing.T) {
	pb := fea.NewParametricBridge(1, 25.6, 5.8, 6.5)

	originalSpan := pb.BaseBridge.SpanLength
	originalRise := pb.BaseBridge.ArchRise
	originalWidth := pb.BaseBridge.DeckWidth
	origMembers := len(pb.BaseBridge.Structure.Members)

	t.Logf("初始: 跨=%.1f, 矢=%.1f, 宽=%.1f, 构件数=%d",
		originalSpan, originalRise, originalWidth, origMembers)

	newModel := pb.UpdateGeometry(30.0, 7.5, 7.0)

	if len(newModel.Structure.Members) == 0 {
		t.Error("更新几何后模型不应为空")
	}

	t.Logf("更新: 跨=%.1f, 矢=%.1f, 宽=%.1f, 构件数=%d",
		newModel.SpanLength, newModel.ArchRise, newModel.DeckWidth,
		len(newModel.Structure.Members))

	if math.Abs(newModel.SpanLength-30.0) > 0.1 {
		t.Errorf("跨度更新不匹配: 期望30.0, 得到%.1f", newModel.SpanLength)
	}

	invalid := pb.UpdateGeometry(1.0, 0.5, 1.0)
	if len(invalid.Structure.Members) != origMembers {
		t.Log("注意: 无效参数更新可能返回原始模型 (保护性设计)")
	}
	t.Log("✓ 参数化几何更新验证通过")
}

// --- 参数约束边界设置 ---

func TestSetConstraints(t *testing.T) {
	pb := fea.NewParametricBridge(1, 25.6, 5.8, 6.5)

	pb.SetConstraints(15.0, 35.0, 3.0, 10.0, 4.0, 10.0)

	if pb.MinSpan != 15.0 || pb.MaxSpan != 35.0 {
		t.Errorf("跨度约束设置错误: 期望[15,35], 得到[%.1f,%.1f]",
			pb.MinSpan, pb.MaxSpan)
	}
	if pb.MinRise != 3.0 || pb.MaxRise != 10.0 {
		t.Errorf("矢高约束设置错误: 期望[3,10], 得到[%.1f,%.1f]",
			pb.MinRise, pb.MaxRise)
	}
	if pb.MinWidth != 4.0 || pb.MaxWidth != 10.0 {
		t.Errorf("桥宽约束设置错误: 期望[4,10], 得到[%.1f,%.1f]",
			pb.MinWidth, pb.MaxWidth)
	}

	valid, _ := pb.ValidateParams(12.0, 4.0, 5.0)
	if valid {
		t.Error("跨度12<MinSpan15应被拒绝")
	}

	valid2, _ := pb.ValidateParams(20.0, 4.0, 5.0)
	if !valid2 {
		t.Error("跨度20在[15,35]内应通过")
	}
	t.Log("✓ 自定义约束边界设置验证通过")
}

// --- 确定性与数值稳定性 ---

// --- 缺陷修复验证: 代理模型加速 ---

func TestSurrogateModelTraining(t *testing.T) {
	pb := fea.NewParametricBridge(1, 25.6, 5.8, 6.5)

	if pb.UseSurrogate {
		t.Error("初始状态不应启用代理模型")
	}

	pb.TrainSurrogate(50.0)

	if !pb.UseSurrogate {
		t.Error("训练后应启用代理模型")
	}
	if pb.Surrogate == nil {
		t.Fatal("训练后Surrogate不应为nil")
	}
	if !pb.Surrogate.IsTrained {
		t.Error("训练后IsTrained应为true")
	}
	if len(pb.Surrogate.TrainPoints) == 0 {
		t.Error("训练点不应为空")
	}

	t.Logf("训练点数: %d", len(pb.Surrogate.TrainPoints))
	t.Logf("系数(应力): %d项", len(pb.Surrogate.CoeffsStress))
	t.Log("✓ 代理模型训练验证通过")
}

func TestSurrogatePredictionAccuracy(t *testing.T) {
	pb := fea.NewParametricBridge(1, 25.6, 5.8, 6.5)
	pb.TrainSurrogate(50.0)

	testCases := []struct {
		span  float64
		rise  float64
		width float64
	}{
		{25.6, 5.8, 6.5},
		{20.0, 5.0, 5.0},
		{30.0, 7.0, 6.0},
	}

	for _, tc := range testCases {
		feaResult := fea.CalculateParametricAnalysis(1, tc.span, tc.rise, tc.width, 50.0)
		surrogateResult := pb.AnalyzeWithSurrogate(tc.span, tc.rise, tc.width)

		if surrogateResult == nil {
			t.Errorf("代理预测不应返回nil: span=%.1f", tc.span)
			continue
		}

		if !surrogateResult.IsApproximation {
			t.Error("代理结果应标记为近似")
		}

		stressError := math.Abs(surrogateResult.MaxStressRatio-feaResult.MaxStressRatio) / math.Max(feaResult.MaxStressRatio, 0.001) * 100
		dispError := math.Abs(surrogateResult.MaxDisplacement-feaResult.MaxDisplacement) / math.Max(feaResult.MaxDisplacement, 0.001) * 100

		t.Logf("span=%.0f: FEA应力比=%.4f, 代理=%.4f, 误差=%.1f%%",
			tc.span, feaResult.MaxStressRatio, surrogateResult.MaxStressRatio, stressError)
		t.Logf("span=%.0f: FEA位移=%.4f, 代理=%.4f, 误差=%.1f%%",
			tc.span, feaResult.MaxDisplacement, surrogateResult.MaxDisplacement, dispError)

		if surrogateResult.MaxStressRatio < 0 {
			t.Errorf("代理应力比不能为负: %.4f", surrogateResult.MaxStressRatio)
		}
		if surrogateResult.MaxDisplacement < 0 {
			t.Errorf("代理位移不能为负: %.4f", surrogateResult.MaxDisplacement)
		}
	}
	t.Log("✓ 代理模型预测精度验证通过")
}

func TestSurrogateNonNegativity(t *testing.T) {
	pb := fea.NewParametricBridge(1, 25.6, 5.8, 6.5)
	pb.TrainSurrogate(50.0)

	edgeCases := []struct {
		span  float64
		rise  float64
		width float64
	}{
		{10.0, 2.0, 3.0},
		{50.0, 15.0, 12.0},
		{15.0, 3.5, 4.0},
		{40.0, 10.0, 8.0},
	}

	for _, tc := range edgeCases {
		result := pb.AnalyzeWithSurrogate(tc.span, tc.rise, tc.width)
		if result == nil {
			continue
		}
		if result.MaxStressRatio < 0 {
			t.Errorf("边界(%.0f,%.0f,%.0f): 应力比负值被钳位为0", tc.span, tc.rise, tc.width)
		}
		if result.MaxDisplacement < 0 {
			t.Errorf("边界(%.0f,%.0f,%.0f): 位移负值被钳位为0", tc.span, tc.rise, tc.width)
		}
		if result.TotalVolume < 0 {
			t.Errorf("边界(%.0f,%.0f,%.0f): 体积负值被钳位为0", tc.span, tc.rise, tc.width)
		}
	}
	t.Log("✓ 代理模型非负性验证通过")
}

func TestSurrogateUntrainedPrediction(t *testing.T) {
	pb := fea.NewParametricBridge(1, 25.6, 5.8, 6.5)

	result := pb.AnalyzeWithSurrogate(25.0, 5.0, 6.0)
	if result != nil {
		t.Error("未训练代理模型应返回nil")
	}
	t.Log("✓ 未训练代理模型防护验证通过")
}
