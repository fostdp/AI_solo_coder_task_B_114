package fatigue

import (
	"math"
)

type FatigueAnalysis struct {
	SpanLength    float64
	MemberCount   int
	FatigueLimit  float64
	EnduranceLimit float64
}

type MemberFatigueResult struct {
	MemberID          int
	MemberType        string
	StressHistory     []float64
	StressRange       []float64
	MeanStress        []float64
	CycleCount        int
	DamageCumulative  float64
	RemainingLife     float64
	FatigueSafetyFactor float64
	DamageContribution map[string]float64
}

type FatigueResult struct {
	MemberResults      []MemberFatigueResult
	TotalDamage        float64
	MaxDamageMemberID  int
	EstimatedLifeYears float64
	HotspotMembers     []int
	CriticalLocations  []string
}

func NewFatigueAnalysis(spanLength float64, memberCount int) *FatigueAnalysis {
	return &FatigueAnalysis{
		SpanLength:     spanLength,
		MemberCount:    memberCount,
		FatigueLimit:   8.5,
		EnduranceLimit: 0.45,
	}
}

func (fa *FatigueAnalysis) RunRainflowCounting(stressHistory []float64) (ranges []float64, means []float64, counts []int) {
	n := len(stressHistory)
	if n < 2 {
		return
	}

	stack := make([]float64, 0, n)
	points := make([]float64, 0, n)

	for _, s := range stressHistory {
		points = append(points, s)
	}

	for len(points) >= 3 {
		x := points[len(points)-3]
		y := points[len(points)-2]
		z := points[len(points)-1]

		s1 := math.Abs(y - x)
		s2 := math.Abs(z - y)

		if s2 >= s1 {
			stack = append(stack, y)
			points = points[:len(points)-2]
			points = append(points, z)
		} else {
			break
		}
	}

	m := len(stack) / 2
	ranges = make([]float64, 0, m)
	means = make([]float64, 0, m)
	counts = make([]int, 0, m)

	for i := 0; i+1 < len(stack); i += 2 {
		sRange := math.Abs(stack[i+1] - stack[i])
		sMean := (stack[i+1] + stack[i]) / 2.0
		ranges = append(ranges, sRange)
		means = append(means, sMean)
		counts = append(counts, 1)
	}

	return
}

func (fa *FatigueAnalysis) SNCurve(stressRange float64, meanStress float64) float64 {
	correctedRange := fa.correctForMeanStress(stressRange, meanStress)

	if correctedRange <= fa.EnduranceLimit*fa.FatigueLimit {
		return 1e10
	}

	m := 5.0
	logN := 13.5 - m*math.Log10(correctedRange)
	return math.Pow(10, logN)
}

func (fa *FatigueAnalysis) correctForMeanStress(stressRange float64, meanStress float64) float64 {
	ultimateStrength := 50.0
	if meanStress > 0 {
		return stressRange / (1.0 - meanStress/ultimateStrength)
	}
	return stressRange
}

func (fa *FatigueAnalysis) MinerRule(ranges []float64, means []float64, counts []int) float64 {
	damage := 0.0
	for i, r := range ranges {
		if r <= 0 {
			continue
		}
		n := float64(counts[i])
		ni := fa.SNCurve(r, means[i])
		if ni > 0 {
			damage += n / ni
		}
	}
	return damage
}

func (fa *FatigueAnalysis) CalculateMemberFatigue(memberID int, memberType string, stressHistory []float64, loadCyclesPerDay float64) MemberFatigueResult {
	ranges, means, counts := fa.RunRainflowCounting(stressHistory)

	damage := fa.MinerRule(ranges, means, counts)

	dailyDamage := damage * loadCyclesPerDay / float64(len(counts)+1)
	remainingLife := 1.0 / (dailyDamage * 365.0)

	maxStress := 0.0
	minStress := 0.0
	if len(stressHistory) > 0 {
		maxStress = stressHistory[0]
		minStress = stressHistory[0]
		for _, s := range stressHistory {
			if s > maxStress {
				maxStress = s
			}
			if s < minStress {
				minStress = s
			}
		}
	}

	allowableRange := fa.FatigueLimit * 0.6
	actualMaxRange := maxStress - minStress
	safetyFactor := allowableRange / actualMaxRange
	if actualMaxRange <= 0 {
		safetyFactor = 10.0
	}

	damageContrib := map[string]float64{
		"low_cycle":    0.0,
		"medium_cycle": 0.0,
		"high_cycle":   0.0,
	}

	for i, r := range ranges {
		if r > fa.FatigueLimit*0.8 {
			damageContrib["low_cycle"] += float64(counts[i]) / fa.SNCurve(r, means[i])
		} else if r > fa.FatigueLimit*0.4 {
			damageContrib["medium_cycle"] += float64(counts[i]) / fa.SNCurve(r, means[i])
		} else {
			damageContrib["high_cycle"] += float64(counts[i]) / fa.SNCurve(r, means[i])
		}
	}

	total := damageContrib["low_cycle"] + damageContrib["medium_cycle"] + damageContrib["high_cycle"]
	if total > 0 {
		for k := range damageContrib {
			damageContrib[k] /= total
		}
	}

	return MemberFatigueResult{
		MemberID:            memberID,
		MemberType:          memberType,
		StressHistory:       stressHistory,
		StressRange:         ranges,
		MeanStress:          means,
		CycleCount:          len(counts),
		DamageCumulative:    damage,
		RemainingLife:       remainingLife,
		FatigueSafetyFactor: safetyFactor,
		DamageContribution:  damageContrib,
	}
}

func (fa *FatigueAnalysis) AggregateResults(memberResults []MemberFatigueResult) FatigueResult {
	totalDamage := 0.0
	maxDamage := 0.0
	maxMemberID := 0
	hotspots := make([]int, 0)

	for _, r := range memberResults {
		totalDamage += r.DamageCumulative
		if r.DamageCumulative > maxDamage {
			maxDamage = r.DamageCumulative
			maxMemberID = r.MemberID
		}
		if r.FatigueSafetyFactor < 1.5 {
			hotspots = append(hotspots, r.MemberID)
		}
	}

	avgDailyDamage := totalDamage / float64(len(memberResults))
	estimatedLife := 1.0 / (avgDailyDamage * 365.0)
	if estimatedLife < 0 {
		estimatedLife = 0
	}

	criticalLocations := []string{"拱脚节点", "跨中拱肋", "斜撑与拱肋连接处", "立柱与桥面交接处"}

	return FatigueResult{
		MemberResults:       memberResults,
		TotalDamage:         totalDamage,
		MaxDamageMemberID:   maxMemberID,
		EstimatedLifeYears:  estimatedLife,
		HotspotMembers:      hotspots,
		CriticalLocations:   criticalLocations,
	}
}

func (fa *FatigueAnalysis) GenerateStressHistoryForSpectrum(spectrum []float64, memberIdx int, totalMembers int) []float64 {
	history := make([]float64, len(spectrum))

	positionEffect := math.Sin(math.Pi * float64(memberIdx+1) / float64(totalMembers+1))

	for t, load := range spectrum {
		baseStress := load * positionEffect * 0.15
		noise := (math.Sin(float64(t)*0.1) + math.Cos(float64(t)*0.07)) * 0.3
		history[t] = baseStress + noise
	}

	return history
}

func (fa *FatigueAnalysis) GetFatigueRecommendations(result FatigueResult) []string {
	recommendations := make([]string, 0)

	if result.EstimatedLifeYears < 50 {
		recommendations = append(recommendations, "结构剩余寿命不足50年，建议进行加固处理")
	} else if result.EstimatedLifeYears < 100 {
		recommendations = append(recommendations, "结构剩余寿命约50-100年，建议定期监测")
	}

	if len(result.HotspotMembers) > 0 {
		recommendations = append(recommendations, "存在疲劳热点构件，建议优先检查加固")
	}

	if result.MaxDamageMemberID > 0 {
		recommendations = append(recommendations, "最大损伤构件建议重点监测")
	}

	if result.TotalDamage > 0.7 {
		recommendations = append(recommendations, "总体损伤水平较高，建议限制重型车辆通行")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "结构疲劳状态良好，按常规周期维护即可")
	}

	return recommendations
}
