package historicalcomparison

import (
	"fmt"
	"math"
)

type Dynasty string

const (
	DynastyHanJin   Dynasty = "han_jin"
	DynastyTang    Dynasty = "tang"
	DynastySong    Dynasty = "song"
	DynastyMing    Dynasty = "ming"
	DynastyQing    Dynasty = "qing"
)

type BridgeTypology string

const (
	TypologyBeamBridge     BridgeTypology = "beam_bridge"
	TypologyTimberArch     BridgeTypology = "timber_arch"
	TypologyThroughArch    BridgeTypology = "through_arch"
	TypologyGalleryBridge  BridgeTypology = "gallery_bridge"
)

type HistoricalBridge struct {
	ID                int
	Name              string
	Dynasty           Dynasty
	Typology          BridgeTypology
	SpanLength        float64
	ArchRise          float64
	DeckWidth         float64
	TotalLength       float64
	MaterialType      string
	ConstructionMethod string
	HistoricalEra     string
	KeyInnovation     string
}

type EfficiencyMetrics struct {
	SpanToDepthRatio        float64
	MaterialEfficiency      float64
	LoadCarryingCapacity    float64
	ConstructionComplexity  float64
	DurabilityScore         float64
	WeightToSpanRatio       float64
}

type ComparisonResult struct {
	BridgeA          HistoricalBridge
	BridgeB          HistoricalBridge
	MetricsA         EfficiencyMetrics
	MetricsB         EfficiencyMetrics
	NormalizedScores map[string]map[Dynasty]float64
	RadarData        []RadarPoint
	AdvantagesA      []string
	AdvantagesB      []string
	HistoricalNotes  []string
	TechEvolution    []TechEvolutionPoint
}

type RadarPoint struct {
	Metric string  `json:"metric"`
	ValueA float64 `json:"value_a"`
	ValueB float64 `json:"value_b"`
}

type TechEvolutionPoint struct {
	Period    string  `json:"period"`
	Innovation string `json:"innovation"`
	Impact    float64 `json:"impact"`
	Year      int     `json:"year"`
}

var HistoricalBridgeDatabase = []HistoricalBridge{
	{
		ID:                101,
		Name:              "灞桥",
		Dynasty:           DynastyHanJin,
		Typology:          TypologyBeamBridge,
		SpanLength:        18.0,
		ArchRise:          2.5,
		DeckWidth:         4.5,
		TotalLength:       400.0,
		MaterialType:      "木梁石墩",
		ConstructionMethod: "石墩木梁",
		HistoricalEra:     "汉晋时期(公元前206年-公元420年)",
		KeyInnovation:     "多跨简支木梁桥，石砌桥墩",
	},
	{
		ID:                102,
		Name:              "枫桥",
		Dynasty:           DynastyTang,
		Typology:          TypologyBeamBridge,
		SpanLength:        18.5,
		ArchRise:          3.8,
		DeckWidth:         4.2,
		TotalLength:       24.0,
		MaterialType:      "木材",
		ConstructionMethod: "单孔木拱",
		HistoricalEra:     "唐代(公元618年-907年)",
		KeyInnovation:     "木拱技术萌芽，向拱结构过渡",
	},
	{
		ID:                103,
		Name:              "汴水虹桥",
		Dynasty:           DynastySong,
		Typology:          TypologyThroughArch,
		SpanLength:        25.6,
		ArchRise:          5.8,
		DeckWidth:         6.5,
		TotalLength:       32.0,
		MaterialType:      "木材",
		ConstructionMethod: "叠梁拱/贯木拱",
		HistoricalEra:     "北宋(公元960年-1127年)",
		KeyInnovation:     "贯木拱技术成熟，无柱大跨木拱桥",
	},
	{
		ID:                104,
		Name:              "龙津桥",
		Dynasty:           DynastyMing,
		Typology:          TypologyGalleryBridge,
		SpanLength:        28.5,
		ArchRise:          6.2,
		DeckWidth:         5.2,
		TotalLength:       35.0,
		MaterialType:      "木材",
		ConstructionMethod: "木拱廊桥",
		HistoricalEra:     "明代(公元1368年-1644年)",
		KeyInnovation:     "木拱廊桥，廊屋保护结构",
	},
	{
		ID:                105,
		Name:              "千乘桥",
		Dynasty:           DynastyMing,
		Typology:          TypologyTimberArch,
		SpanLength:        27.3,
		ArchRise:          5.9,
		DeckWidth:         5.0,
		TotalLength:       34.0,
		MaterialType:      "木材",
		ConstructionMethod: "贯木拱",
		HistoricalEra:     "明代",
		KeyInnovation:     "三节拱五节拱组合技术",
	},
	{
		ID:                106,
		Name:              "飞虹桥",
		Dynasty:           DynastyQing,
		Typology:          TypologyGalleryBridge,
		SpanLength:        19.5,
		ArchRise:          4.2,
		DeckWidth:         4.5,
		TotalLength:       26.0,
		MaterialType:      "木材",
		ConstructionMethod: "木拱廊桥",
		HistoricalEra:     "清代(公元1644年-1912年)",
		KeyInnovation:     "工艺精细化，装饰艺术发展",
	},
	{
		ID:                107,
		Name:              "安澜桥",
		Dynasty:           DynastySong,
		Typology:          TypologyBeamBridge,
		SpanLength:        24.0,
		ArchRise:          5.2,
		DeckWidth:         4.8,
		TotalLength:       31.0,
		MaterialType:      "竹木",
		ConstructionMethod: "竹索木桥",
		HistoricalEra:     "宋代",
		KeyInnovation:     "竹索加固木梁技术",
	},
}

func NewHistoricalComparison() *HistoricalComparison {
	return &HistoricalComparison{}
}

type HistoricalComparison struct{}

func (hc *HistoricalComparison) CalculateEfficiency(bridge HistoricalBridge) EfficiencyMetrics {
	spanToDepthRatio := bridge.SpanLength / math.Max(bridge.ArchRise, 0.1)

	materialEfficiency := 0.0
	switch bridge.Typology {
	case TypologyBeamBridge:
		materialEfficiency = 60.0
	case TypologyTimberArch:
		materialEfficiency = 78.0
	case TypologyThroughArch:
		materialEfficiency = 90.0
	case TypologyGalleryBridge:
		materialEfficiency = 85.0
	}

	loadCapacity := 0.0
	switch bridge.Dynasty {
	case DynastyHanJin:
		loadCapacity = 50.0
	case DynastyTang:
		loadCapacity = 65.0
	case DynastySong:
		loadCapacity = 85.0
	case DynastyMing:
		loadCapacity = 90.0
	case DynastyQing:
		loadCapacity = 88.0
	}

	complexity := 0.0
	switch bridge.ConstructionMethod {
	case "石墩木梁":
		complexity = 40.0
	case "单孔木拱":
		complexity = 55.0
	case "叠梁拱/贯木拱", "贯木拱":
		complexity = 78.0
	case "木拱廊桥":
		complexity = 90.0
	case "竹索木桥":
		complexity = 50.0
	}

	durability := 0.0
	switch bridge.Typology {
	case TypologyBeamBridge:
		durability = 50.0
	case TypologyTimberArch:
		durability = 70.0
	case TypologyThroughArch:
		durability = 65.0
	case TypologyGalleryBridge:
		durability = 85.0
	}

	weightToSpanRatio := 0.0
	if bridge.Typology == TypologyBeamBridge {
		weightToSpanRatio = 2.5
	} else if bridge.Typology == TypologyThroughArch {
		weightToSpanRatio = 1.2
	} else {
		weightToSpanRatio = 1.5
	}

	return EfficiencyMetrics{
		SpanToDepthRatio:       spanToDepthRatio,
		MaterialEfficiency:     materialEfficiency,
		LoadCarryingCapacity:   loadCapacity,
		ConstructionComplexity: complexity,
		DurabilityScore:        durability,
		WeightToSpanRatio:      weightToSpanRatio,
	}
}

func (hc *HistoricalComparison) CompareBridges(bridgeA HistoricalBridge, bridgeB HistoricalBridge) ComparisonResult {
	metricsA := hc.CalculateEfficiency(bridgeA)
	metricsB := hc.CalculateEfficiency(bridgeB)

	allMetrics := []struct {
		name string
		a    float64
		b    float64
	}{
		{"跨高比", metricsA.SpanToDepthRatio, metricsB.SpanToDepthRatio},
		{"材料效率", metricsA.MaterialEfficiency, metricsB.MaterialEfficiency},
		{"承载能力", metricsA.LoadCarryingCapacity, metricsB.LoadCarryingCapacity},
		{"施工复杂度", metricsA.ConstructionComplexity, metricsB.ConstructionComplexity},
		{"耐久性评分", metricsA.DurabilityScore, metricsB.DurabilityScore},
		{"自重跨度比", 100 / metricsA.WeightToSpanRatio, 100 / metricsB.WeightToSpanRatio},
	}

	normalized := make(map[string]map[Dynasty]float64)
	for _, m := range allMetrics {
		maxVal := math.Max(m.a, m.b)
		if maxVal == 0 {
			maxVal = 1
		}
		normalized[m.name] = map[Dynasty]float64{
			bridgeA.Dynasty: m.a / maxVal * 100,
			bridgeB.Dynasty: m.b / maxVal * 100,
		}
	}

	radarData := make([]RadarPoint, len(allMetrics))
	for i, m := range allMetrics {
		radarData[i] = RadarPoint{
			Metric: m.name,
			ValueA: normalized[m.name][bridgeA.Dynasty],
			ValueB: normalized[m.name][bridgeB.Dynasty],
		}
	}

	advantagesA := hc.analyzeAdvantages(metricsA, metricsB, bridgeA, bridgeB)
	advantagesB := hc.analyzeAdvantages(metricsB, metricsA, bridgeB, bridgeA)

	notes := hc.generateHistoricalNotes(bridgeA, bridgeB)
	evolution := hc.getTechEvolution()

	return ComparisonResult{
		BridgeA:          bridgeA,
		BridgeB:          bridgeB,
		MetricsA:         metricsA,
		MetricsB:         metricsB,
		NormalizedScores: normalized,
		RadarData:        radarData,
		AdvantagesA:      advantagesA,
		AdvantagesB:      advantagesB,
		HistoricalNotes:  notes,
		TechEvolution:    evolution,
	}
}

func (hc *HistoricalComparison) analyzeAdvantages(a, b EfficiencyMetrics, bridgeA, bridgeB HistoricalBridge) []string {
	advantages := make([]string, 0)

	if a.MaterialEfficiency > b.MaterialEfficiency {
		diff := a.MaterialEfficiency - b.MaterialEfficiency
		advantages = append(advantages,
			"材料效率更高 "+bridgeA.Name+" 比 "+bridgeB.Name+" 高出 "+formatFloat(diff)+"%")
	}

	if a.LoadCarryingCapacity > b.LoadCarryingCapacity {
		diff := a.LoadCarryingCapacity - b.LoadCarryingCapacity
		advantages = append(advantages,
			"承载能力更强 高出 "+formatFloat(diff)+"%")
	}

	if a.DurabilityScore > b.DurabilityScore {
		diff := a.DurabilityScore - b.DurabilityScore
		advantages = append(advantages,
			"耐久性更优 评分高出 "+formatFloat(diff)+"分")
	}

	if a.SpanToDepthRatio > b.SpanToDepthRatio*1.2 {
		advantages = append(advantages,
			"结构造型更优美 跨高比更大，视觉更轻盈")
	}

	if a.ConstructionComplexity < b.ConstructionComplexity*0.8 {
		advantages = append(advantages,
			"施工更简便 复杂度降低 "+formatFloat((1-a.ConstructionComplexity/b.ConstructionComplexity)*100)+"%")
	}

	if a.WeightToSpanRatio < b.WeightToSpanRatio*0.8 {
		advantages = append(advantages,
			"结构自重更轻 每米跨度自重降低 "+formatFloat((1-a.WeightToSpanRatio/b.WeightToSpanRatio)*100)+"%")
	}

	return advantages
}

func (hc *HistoricalComparison) generateHistoricalNotes(a, b HistoricalBridge) []string {
	notes := make([]string, 0)

	notes = append(notes, a.HistoricalEra+"："+a.KeyInnovation)
	notes = append(notes, b.HistoricalEra+"："+b.KeyInnovation)

	if a.Dynasty == DynastyHanJin && b.Dynasty == DynastySong {
		notes = append(notes, "技术跃迁：从汉晋简支木梁到宋代贯木拱，跨度能力提升约40%")
		notes = append(notes, "材料效率提升：木拱结构相比简支梁可节省约25%的木材用量")
		notes = append(notes, "美学进步：从厚重墩柱到无柱飞拱，体现宋代审美追求")
	}

	if a.Typology == TypologyBeamBridge && b.Typology == TypologyThroughArch {
		notes = append(notes, "力学原理演变：从受弯为主的梁结构到受压为主的拱结构")
		notes = append(notes, "跨度突破：木拱技术使单跨能力从20米级提升至30米级")
	}

	return notes
}

func (hc *HistoricalComparison) getTechEvolution() []TechEvolutionPoint {
	return []TechEvolutionPoint{
		{Period: "西周", Innovation: "简支木梁桥出现", Impact: 30.0, Year: -1000},
		{Period: "秦汉", Innovation: "石墩木梁桥普及", Impact: 45.0, Year: -200},
		{Period: "南北朝", Innovation: "木拱技术萌芽", Impact: 60.0, Year: 500},
		{Period: "唐代", Innovation: "单孔木拱桥出现", Impact: 70.0, Year: 700},
		{Period: "北宋", Innovation: "贯木拱技术成熟", Impact: 95.0, Year: 1050},
		{Period: "南宋", Innovation: "木拱廊桥发展", Impact: 85.0, Year: 1200},
		{Period: "明代", Innovation: "三节拱五节拱组合", Impact: 88.0, Year: 1450},
		{Period: "清代", Innovation: "工艺精细化与装饰", Impact: 80.0, Year: 1700},
	}
}

func (hc *HistoricalComparison) GetBridgesByDynasty(dynasty Dynasty) []HistoricalBridge {
	result := make([]HistoricalBridge, 0)
	for _, b := range HistoricalBridgeDatabase {
		if b.Dynasty == dynasty {
			result = append(result, b)
		}
	}
	return result
}

func (hc *HistoricalComparison) GetAllBridges() []HistoricalBridge {
	return HistoricalBridgeDatabase
}

func formatFloat(v float64) string {
	if v < 0 {
		v = -v
	}
	return fmt.Sprintf("%.1f", v)
}
