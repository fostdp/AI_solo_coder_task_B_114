package fea

import (
	"fmt"
	"math"
)

type ParametricBridge struct {
	BaseBridge *BridgeModel
	MinSpan    float64
	MaxSpan    float64
	MinRise    float64
	MaxRise    float64
	MinWidth   float64
	MaxWidth   float64
}

func NewParametricBridge(bridgeID int, defaultSpan, defaultRise, defaultWidth float64) *ParametricBridge {
	base := GenerateArchBridge(bridgeID, defaultSpan, defaultRise, defaultWidth)
	return &ParametricBridge{
		BaseBridge: base,
		MinSpan:    10.0,
		MaxSpan:    50.0,
		MinRise:    2.0,
		MaxRise:    15.0,
		MinWidth:   3.0,
		MaxWidth:   12.0,
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

	newModel := GenerateArchBridge(pb.BaseBridge.BridgeName, span, rise, width)

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

type GeometryOption struct {
	Name    string  `json:"name"`
	Label   string  `json:"label"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Step    float64 `json:"step"`
	Default float64 `json:"default"`
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
	ID       int     `json:"id"`
	StartID  int     `json:"start_id"`
	EndID    int     `json:"end_id"`
	Type     string  `json:"type"`
	Length   float64 `json:"length"`
	Area     float64 `json:"area"`
	Stress   float64 `json:"stress"`
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
