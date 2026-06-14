package fea

import (
	"fmt"
	"math"
)

type ParametricBridge struct {
	BridgeID    int
	BaseBridge  *BridgeModel
	MinSpan     float64
	MaxSpan     float64
	MinRise     float64
	MaxRise     float64
	MinWidth    float64
	MaxWidth    float64
	UseSurrogate bool
	Surrogate   *SurrogateModel
}

func NewParametricBridge(bridgeID int, defaultSpan, defaultRise, defaultWidth float64) *ParametricBridge {
	base := GenerateArchBridge(bridgeID, defaultSpan, defaultRise, defaultWidth)
	return &ParametricBridge{
		BridgeID:    bridgeID,
		BaseBridge:  base,
		MinSpan:     10.0,
		MaxSpan:     50.0,
		MinRise:     2.0,
		MaxRise:     15.0,
		MinWidth:    3.0,
		MaxWidth:    12.0,
		UseSurrogate: false,
	}
}

func (pb *ParametricBridge) SetConstraints(minSpan, maxSpan, minRise, maxRise, minWidth, maxWidth float64) {
	pb.MinSpan = minSpan
	pb.MaxSpan = maxSpan
	pb.MinRise = minRise
	pb.MaxRise = maxRise
	pb.MinWidth = minWidth
	pb.MaxWidth = maxWidth
}

func (pb *ParametricBridge) ValidateParams(span, rise, width float64) (valid bool, message string) {
	if span < pb.MinSpan || span > pb.MaxSpan {
		return false, "拱跨超出范围: " + formatFloat(pb.MinSpan) + "m - " + formatFloat(pb.MaxSpan) + "m"
	}
	if rise < pb.MinRise || rise > pb.MaxRise {
		return false, "矢高超出范围: " + formatFloat(pb.MinRise) + "m - " + formatFloat(pb.MaxRise) + "m"
	}
	if width < pb.MinWidth || width > pb.MaxWidth {
		return false, "桥面宽度超出范围: " + formatFloat(pb.MinWidth) + "m - " + formatFloat(pb.MaxWidth) + "m"
	}

	riseSpanRatio := rise / span
	if riseSpanRatio < 0.08 {
		return false, "矢跨比过小，建议大于1:12"
	}
	if riseSpanRatio > 0.35 {
		return false, "矢跨比过大，建议小于1:3"
	}

	return true, "参数合法"
}

func (pb *ParametricBridge) UpdateGeometry(span, rise, width float64) *BridgeModel {
	valid, _ := pb.ValidateParams(span, rise, width)
	if !valid {
		return pb.BaseBridge
	}

	newModel := GenerateArchBridge(pb.BridgeID, span, rise, width)

	if len(newModel.Structure.Members) > 0 {
		newModel.Structure.AssembleStiffness()
	}

	return newModel
}

func (pb *ParametricBridge) GetGeometryOptions() []GeometryOption {
	return []GeometryOption{
		{Name: "span_length", Label: "拱跨 (m)", Min: pb.MinSpan, Max: pb.MaxSpan, Step: 0.5, Default: pb.BaseBridge.SpanLength},
		{Name: "arch_rise", Label: "矢高 (m)", Min: pb.MinRise, Max: pb.MaxRise, Step: 0.1, Default: pb.BaseBridge.ArchRise},
		{Name: "deck_width", Label: "桥面宽度 (m)", Min: pb.MinWidth, Max: pb.MaxWidth, Step: 0.1, Default: pb.BaseBridge.DeckWidth},
	}
}

func (pb *ParametricBridge) TrainSurrogate(loadValue float64) {
	if pb.Surrogate == nil {
		pb.Surrogate = NewSurrogateModel()
	}
	pb.Surrogate.Train(pb.BridgeID, loadValue, pb.MinSpan, pb.MaxSpan, pb.MinRise, pb.MaxRise, pb.MinWidth, pb.MaxWidth)
	pb.UseSurrogate = true
}

func (pb *ParametricBridge) AnalyzeWithSurrogate(span, rise, width float64) *SurrogateResult {
	if pb.Surrogate == nil || !pb.UseSurrogate {
		return nil
	}
	return pb.Surrogate.Predict(span, rise, width)
}

type GeometryOption struct {
	Name    string  `json:"name"`
	Label   string  `json:"label"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Step    float64 `json:"step"`
	Default float64 `json:"default"`
}

type SurrogateModel struct {
	CoeffsStress     []float64
	CoeffsDisp       []float64
	CoeffsVolume     []float64
	TrainPoints      []SurrogateTrainPoint
	IsTrained        bool
	SpanMin          float64
	SpanMax          float64
	RiseMin          float64
	RiseMax          float64
	WidthMin         float64
	WidthMax         float64
}

type SurrogateTrainPoint struct {
	Span          float64
	Rise          float64
	Width         float64
	MaxStressRatio float64
	MaxDisp        float64
	TotalVolume    float64
}

type SurrogateResult struct {
	MaxStressRatio float64
	MaxDisplacement float64
	TotalVolume    float64
	IsApproximation bool
}

func NewSurrogateModel() *SurrogateModel {
	return &SurrogateModel{
		IsTrained: false,
	}
}

func (sm *SurrogateModel) Train(bridgeID int, loadValue, spanMin, spanMax, riseMin, riseMax, widthMin, widthMax float64) {
	sm.SpanMin = spanMin
	sm.SpanMax = spanMax
	sm.RiseMin = riseMin
	sm.RiseMax = riseMax
	sm.WidthMin = widthMin
	sm.WidthMax = widthMax

	spanLevels := []float64{spanMin, (spanMin + spanMax) / 2, spanMax}
	riseLevels := []float64{riseMin, (riseMin + riseMax) / 2, riseMax}
	widthLevels := []float64{widthMin, (widthMin + widthMax) / 2, widthMax}

	sm.TrainPoints = make([]SurrogateTrainPoint, 0)

	for _, s := range spanLevels {
		for _, r := range riseLevels {
			for _, w := range widthLevels {
				result := CalculateParametricAnalysis(bridgeID, s, r, w, loadValue)
				sm.TrainPoints = append(sm.TrainPoints, SurrogateTrainPoint{
					Span:           s,
					Rise:           r,
					Width:          w,
					MaxStressRatio: result.MaxStressRatio,
					MaxDisp:        result.MaxDisplacement,
					TotalVolume:    result.TotalVolume,
				})
			}
		}
	}

	sm.fitPolynomial()
	sm.IsTrained = true
}

func (sm *SurrogateModel) fitPolynomial() {
	n := len(sm.TrainPoints)
	if n < 10 {
		return
	}

	X := make([][]float64, n)
	yStress := make([]float64, n)
	yDisp := make([]float64, n)
	yVol := make([]float64, n)

	for i, p := range sm.TrainPoints {
		s := sm.normalize(p.Span, sm.SpanMin, sm.SpanMax)
		r := sm.normalize(p.Rise, sm.RiseMin, sm.RiseMax)
		w := sm.normalize(p.Width, sm.WidthMin, sm.WidthMax)

		X[i] = []float64{
			1.0,
			s, r, w,
			s * s, r * r, w * w,
			s * r, s * w, r * w,
		}
		yStress[i] = p.MaxStressRatio
		yDisp[i] = p.MaxDisp
		yVol[i] = p.TotalVolume
	}

	sm.CoeffsStress = leastSquares(X, yStress)
	sm.CoeffsDisp = leastSquares(X, yDisp)
	sm.CoeffsVolume = leastSquares(X, yVol)
}

func (sm *SurrogateModel) Predict(span, rise, width float64) *SurrogateResult {
	if !sm.IsTrained {
		return nil
	}

	s := sm.normalize(span, sm.SpanMin, sm.SpanMax)
	r := sm.normalize(rise, sm.RiseMin, sm.RiseMax)
	w := sm.normalize(width, sm.WidthMin, sm.WidthMax)

	x := []float64{
		1.0,
		s, r, w,
		s * s, r * r, w * w,
		s * r, s * w, r * w,
	}

	stress := polyEval(sm.CoeffsStress, x)
	disp := polyEval(sm.CoeffsDisp, x)
	vol := polyEval(sm.CoeffsVolume, x)

	if stress < 0 {
		stress = 0
	}
	if disp < 0 {
		disp = 0
	}
	if vol < 0 {
		vol = 0
	}

	return &SurrogateResult{
		MaxStressRatio:  stress,
		MaxDisplacement: disp,
		TotalVolume:     vol,
		IsApproximation: true,
	}
}

func (sm *SurrogateModel) normalize(val, min, max float64) float64 {
	if max <= min {
		return 0.5
	}
	return (val - min) / (max - min)
}

func polyEval(coeffs, x []float64) float64 {
	result := 0.0
	for i, c := range coeffs {
		if i < len(x) {
			result += c * x[i]
		}
	}
	return result
}

func leastSquares(X [][]float64, y []float64) []float64 {
	n := len(X[0])
	m := len(X)

	XTX := make([][]float64, n)
	for i := range XTX {
		XTX[i] = make([]float64, n)
	}
	XTy := make([]float64, n)

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			for k := 0; k < m; k++ {
				XTX[i][j] += X[k][i] * X[k][j]
			}
		}
		for k := 0; k < m; k++ {
			XTy[i] += X[k][i] * y[k]
		}
	}

	augmented := make([][]float64, n)
	for i := range augmented {
		augmented[i] = make([]float64, n+1)
		copy(augmented[i], XTX[i])
		augmented[i][n] = XTy[i]
	}

	for col := 0; col < n; col++ {
		maxRow := col
		for row := col + 1; row < n; row++ {
			if math.Abs(augmented[row][col]) > math.Abs(augmented[maxRow][col]) {
				maxRow = row
			}
		}
		augmented[col], augmented[maxRow] = augmented[maxRow], augmented[col]

		if math.Abs(augmented[col][col]) < 1e-12 {
			continue
		}

		for row := col + 1; row < n; row++ {
			factor := augmented[row][col] / augmented[col][col]
			for j := col; j <= n; j++ {
				augmented[row][j] -= factor * augmented[col][j]
			}
		}
	}

	coeffs := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		if math.Abs(augmented[i][i]) < 1e-12 {
			coeffs[i] = 0
			continue
		}
		coeffs[i] = augmented[i][n]
		for j := i + 1; j < n; j++ {
			coeffs[i] -= augmented[i][j] * coeffs[j]
		}
		coeffs[i] /= augmented[i][i]
	}

	return coeffs
}

type ParametricAnalysisResult struct {
	BridgeID            int
	SpanLength          float64
	ArchRise            float64
	DeckWidth           float64
	RiseSpanRatio       float64
	MemberCount         int
	NodeCount           int
	MaxStressRatio      float64
	MaxDisplacement     float64
	MaxDisplacementNode int
	TotalVolume         float64
	MaterialEfficiency  float64
	AnalysisDuration_ms float64
	MemberForces        []MemberForces
	Displacements       []NodeDisplacement
	YingzaoComparison   []YingzaoComparison
	Nodes               []NodeInfo
	Members             []MemberInfo
	IsSurrogateResult   bool
}

type NodeInfo struct {
	ID        int     `json:"id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	UX        float64 `json:"ux"`
	UY        float64 `json:"uy"`
	IsSupport bool    `json:"is_support"`
}

type MemberInfo struct {
	ID          int     `json:"id"`
	StartID     int     `json:"start_id"`
	EndID       int     `json:"end_id"`
	Type        string  `json:"type"`
	Length      float64 `json:"length"`
	Area        float64 `json:"area"`
	Stress      float64 `json:"stress"`
	StressRatio float64 `json:"stress_ratio"`
}

func CalculateParametricAnalysis(bridgeID int, span, rise, width float64, loadValue float64) ParametricAnalysisResult {
	model := GenerateArchBridge(bridgeID, span, rise, width)
	model.Structure.AssembleStiffness()

	loads := make(map[int]float64)
	numDeckNodes := (len(model.Structure.Nodes) / 2)
	startDeckID := numDeckNodes + 1
	for i := 0; i < numDeckNodes; i++ {
		nodeID := startDeckID + i
		loads[nodeID] = -loadValue / float64(numDeckNodes)
	}

	model.Structure.ApplyLoads(loads)
	model.Structure.Solve()

	forces := model.Structure.CalculateMemberForces()
	displacements := model.Structure.GetNodeDisplacements()
	comparison := model.Structure.CompareWithYingzaoFashi()

	maxStressRatio := 0.0
	maxDisp := 0.0
	maxDispNode := 0
	totalVolume := 0.0

	for _, f := range forces {
		if math.Abs(f.StressRatio) > maxStressRatio {
			maxStressRatio = math.Abs(f.StressRatio)
		}
		totalVolume += f.Length * model.Structure.Members[f.MemberID-1].A
	}

	for _, d := range displacements {
		disp := math.Sqrt(d.UX*d.UX + d.UY*d.UY)
		if disp > maxDisp {
			maxDisp = disp
			maxDispNode = d.NodeID
		}
	}

	riseSpanRatio := rise / span
	materialEfficiency := maxStressRatio / (totalVolume / (span * width))
	if materialEfficiency == 0 {
		materialEfficiency = 0
	}

	nodeInfos := make([]NodeInfo, 0, len(model.Structure.Nodes))
	for _, node := range model.Structure.Nodes {
		nodeInfos = append(nodeInfos, NodeInfo{
			ID:        node.ID,
			X:         node.X,
			Y:         node.Y,
			UX:        node.UX,
			UY:        node.UY,
			IsSupport: node.FixedX && node.FixedY,
		})
	}

	memberInfos := make([]MemberInfo, 0, len(forces))
	for _, f := range forces {
		member := model.Structure.Members[f.MemberID-1]
		memberInfos = append(memberInfos, MemberInfo{
			ID:          member.ID,
			StartID:     member.Start.ID,
			EndID:       member.End.ID,
			Type:        model.MemberTypes[member.ID],
			Length:      member.Length,
			Area:        member.A,
			Stress:      f.NormalStress + f.BendingStress,
			StressRatio: f.StressRatio,
		})
	}

	return ParametricAnalysisResult{
		BridgeID:            bridgeID,
		SpanLength:          span,
		ArchRise:            rise,
		DeckWidth:           width,
		RiseSpanRatio:       riseSpanRatio,
		MemberCount:         len(model.Structure.Members),
		NodeCount:           len(model.Structure.Nodes),
		MaxStressRatio:      maxStressRatio,
		MaxDisplacement:     maxDisp,
		MaxDisplacementNode: maxDispNode,
		TotalVolume:         totalVolume,
		MaterialEfficiency:  materialEfficiency,
		MemberForces:        forces,
		Displacements:       displacements,
		YingzaoComparison:   comparison,
		Nodes:               nodeInfos,
		Members:             memberInfos,
	}
}

func formatFloat(v float64) string {
	if v < 0 {
		v = -v
	}
	return fmt.Sprintf("%.1f", v)
}
