package movingloadsimulator

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"ancient-bridge-system/internal/fea"
)

type AgentType string

const (
	AgentPedestrian AgentType = "pedestrian"
	AgentOxCart     AgentType = "ox_cart"
	AgentHorseCart  AgentType = "horse_cart"
	AgentSedanChair AgentType = "sedan_chair"
	AgentMilitary   AgentType = "military_convoy"
	AgentPeddler    AgentType = "peddler"
)

type Agent struct {
	ID            int
	Type          AgentType
	X             float64
	Velocity      float64
	TargetX       float64
	Weight        float64
	DesiredVel    float64
	Radius        float64
	ArrivalTime   float64
	Active        bool
	PhaseOffset   float64
	StepFrequency float64
}

type SocialForceModel struct {
	Agents         []*Agent
	SpanLength     float64
	StartTime      time.Time
	CurrentTime    float64
	RandomSeed     int64
	TimeStep       float64
	SyncFactor     float64
}

var agentParameters = map[AgentType]struct {
	Weight        float64
	DesiredVel    float64
	Radius        float64
	Probability   float64
	StepFrequency float64
}{
	AgentPedestrian: {Weight: 0.7, DesiredVel: 1.2, Radius: 0.3, Probability: 0.45, StepFrequency: 2.0},
	AgentOxCart:     {Weight: 30.0, DesiredVel: 0.8, Radius: 1.2, Probability: 0.20, StepFrequency: 1.0},
	AgentHorseCart:  {Weight: 50.0, DesiredVel: 1.5, Radius: 1.5, Probability: 0.15, StepFrequency: 1.5},
	AgentSedanChair: {Weight: 8.0, DesiredVel: 1.0, Radius: 0.8, Probability: 0.08, StepFrequency: 1.8},
	AgentMilitary:   {Weight: 80.0, DesiredVel: 1.8, Radius: 2.0, Probability: 0.05, StepFrequency: 1.2},
	AgentPeddler:    {Weight: 3.0, DesiredVel: 0.6, Radius: 0.5, Probability: 0.07, StepFrequency: 1.6},
}

func NewSocialForceModel(spanLength float64) *SocialForceModel {
	return &SocialForceModel{
		SpanLength: spanLength,
		StartTime:  time.Now(),
		TimeStep:   0.1,
		SyncFactor: 0.1,
	}
}

func (sfm *SocialForceModel) SetRandomSeed(seed int64) {
	sfm.RandomSeed = seed
	rand.Seed(seed)
}

func (sfm *SocialForceModel) SetSyncFactor(factor float64) {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	sfm.SyncFactor = factor
}

func (sfm *SocialForceModel) SpawnAgents(durationSeconds float64, crowdDensity float64) {
	numAgents := int(crowdDensity * sfm.SpanLength / 2.0)
	if numAgents < 1 {
		numAgents = 1
	}

	idCounter := len(sfm.Agents) + 1

	for i := 0; i < numAgents; i++ {
		agentType := sfm.selectAgentType()
		params := agentParameters[agentType]

		startSide := rand.Float64() < 0.5
		startX := 0.0
		targetX := sfm.SpanLength
		if !startSide {
			startX = sfm.SpanLength
			targetX = 0.0
		}

		phaseOffset := rand.Float64() * 2.0 * math.Pi
		stepFreq := params.StepFrequency * (0.85 + rand.Float64()*0.3)

		agent := &Agent{
			ID:            idCounter,
			Type:          agentType,
			X:             startX + (rand.Float64()-0.5)*2.0,
			Velocity:      params.DesiredVel * (0.8 + rand.Float64()*0.4),
			TargetX:       targetX,
			Weight:        params.Weight * (0.9 + rand.Float64()*0.2),
			DesiredVel:    params.DesiredVel,
			Radius:        params.Radius,
			ArrivalTime:   sfm.CurrentTime + rand.Float64()*durationSeconds,
			Active:        true,
			PhaseOffset:   phaseOffset,
			StepFrequency: stepFreq,
		}
		sfm.Agents = append(sfm.Agents, agent)
		idCounter++
	}
}

func (sfm *SocialForceModel) selectAgentType() AgentType {
	r := rand.Float64()
	cumulative := 0.0
	for at, params := range agentParameters {
		cumulative += params.Probability
		if r <= cumulative {
			return at
		}
	}
	return AgentPedestrian
}

func (sfm *SocialForceModel) calculateSocialForce(agent *Agent) float64 {
	repulsiveForce := 0.0
	for _, other := range sfm.Agents {
		if other == agent || !other.Active {
			continue
		}
		dx := agent.X - other.X
		dist := math.Abs(dx)
		minDist := agent.Radius + other.Radius + 0.5
		if dist < minDist*3.0 {
			repulsiveForce += math.Exp(-(dist - minDist) / 0.5)
		}
	}
	return repulsiveForce
}

func (sfm *SocialForceModel) Step() {
	for _, agent := range sfm.Agents {
		if !agent.Active {
			continue
		}
		if sfm.CurrentTime < agent.ArrivalTime {
			continue
		}

		socialForce := sfm.calculateSocialForce(agent)

		dir := 1.0
		if agent.TargetX < agent.X {
			dir = -1.0
		}

		desiredDir := dir * (agent.DesiredVel - agent.Velocity) / 0.5

		agent.Velocity += (desiredDir - socialForce*dir) * sfm.TimeStep
		agent.Velocity = math.Max(0, math.Min(agent.DesiredVel*1.5, agent.Velocity))

		agent.X += agent.Velocity * sfm.TimeStep * dir

		if (dir > 0 && agent.X >= agent.TargetX) || (dir < 0 && agent.X <= agent.TargetX) {
			agent.Active = false
		}
	}
	sfm.CurrentTime += sfm.TimeStep
}

func (sfm *SocialForceModel) calculateDynamicLoadFactor(agent *Agent, timeSeconds float64) float64 {
	if agent.StepFrequency <= 0 {
		return 1.0
	}

	dynamicAmplitude := sfm.SyncFactor * 0.4

	oscillation := dynamicAmplitude * math.Sin(
		2.0*math.Pi*agent.StepFrequency*timeSeconds+agent.PhaseOffset)

	return 1.0 + oscillation
}

type LoadTimeStep struct {
	TimeStep         int
	TimeSeconds      float64
	TotalLoad        float64
	ActiveAgentCount int
	LoadDistribution map[int]float64
}

func (sfm *SocialForceModel) GetLoadSpectrum(numTimeSteps int, timeStep float64) []LoadTimeStep {
	spectrum := make([]LoadTimeStep, 0, numTimeSteps)

	for t := 0; t < numTimeSteps; t++ {
		sfm.Step()

		load := 0.0
		activeCount := 0
		positionalLoads := make(map[int]float64)

		timeSeconds := float64(t) * timeStep

		for _, agent := range sfm.Agents {
			if agent.Active && sfm.CurrentTime >= agent.ArrivalTime {
				dynamicFactor := sfm.calculateDynamicLoadFactor(agent, timeSeconds)

				effectiveWeight := agent.Weight * dynamicFactor

				positionIdx := int(agent.X / sfm.SpanLength * 10)
				if positionIdx < 0 {
					positionIdx = 0
				}
				if positionIdx > 9 {
					positionIdx = 9
				}
				positionalLoads[positionIdx] += effectiveWeight
				load += effectiveWeight
				activeCount++
			}
		}

		spectrum = append(spectrum, LoadTimeStep{
			TimeStep:         t,
			TimeSeconds:      timeSeconds,
			TotalLoad:        load,
			ActiveAgentCount: activeCount,
			LoadDistribution: positionalLoads,
		})
	}

	return spectrum
}

func (sfm *SocialForceModel) Reset() {
	sfm.Agents = sfm.Agents[:0]
	sfm.CurrentTime = 0
}

func (sfm *SocialForceModel) GetActiveAgents() []*Agent {
	active := make([]*Agent, 0)
	for _, a := range sfm.Agents {
		if a.Active && sfm.CurrentTime >= a.ArrivalTime {
			active = append(active, a)
		}
	}
	return active
}

type FatigueAnalysis struct {
	SpanLength     float64
	MemberCount    int
	FatigueLimit   float64
	EnduranceLimit float64
}

type MemberFatigueResult struct {
	MemberID             int
	MemberType           string
	StressHistory        []float64
	StressRange          []float64
	MeanStress           []float64
	CycleCount           int
	DamageCumulative     float64
	RemainingLife        float64
	FatigueSafetyFactor  float64
	DamageContribution   map[string]float64
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
		MemberResults:      memberResults,
		TotalDamage:        totalDamage,
		MaxDamageMemberID:  maxMemberID,
		EstimatedLifeYears: estimatedLife,
		HotspotMembers:     hotspots,
		CriticalLocations:  criticalLocations,
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

type AsyncInput struct {
	BridgeID         int
	SpanLength       float64
	ArchRise         float64
	DeckWidth        float64
	Duration         float64
	CrowdDensity     float64
	TimeStep         float64
	RandomSeed       int64
	LoadCyclesPerDay float64
}

type AsyncResult struct {
	LoadSpectrum  []LoadTimeStep
	MaxLoad       float64
	AvgLoad       float64
	TotalAgents   int
	Duration      float64
	FatigueResult FatigueResult
	PeakLocations []int
	Error         error
}

type AsyncSimulation struct {
	Input  AsyncInput
	Result AsyncResult
	mu     sync.Mutex
}

func SimulateAsync(input AsyncInput) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- AsyncResult{
					Error: fmt.Errorf("simulation panic: %v", r),
				}
			}
			close(resultChan)
		}()

		sfm := NewSocialForceModel(input.SpanLength)
		if input.RandomSeed != 0 {
			sfm.SetRandomSeed(input.RandomSeed)
		}
		if input.TimeStep > 0 {
			sfm.TimeStep = input.TimeStep
		}

		sfm.SpawnAgents(input.Duration, input.CrowdDensity)
		totalAgents := len(sfm.Agents)

		numSteps := int(input.Duration / sfm.TimeStep)
		if numSteps < 1 {
			numSteps = 1
		}

		spectrum := sfm.GetLoadSpectrum(numSteps, sfm.TimeStep)

		maxLoad := 0.0
		totalLoad := 0.0
		peakLocations := make([]int, 0)
		loadValues := make([]float64, len(spectrum))

		for i, ts := range spectrum {
			loadValues[i] = ts.TotalLoad
			if ts.TotalLoad > maxLoad {
				maxLoad = ts.TotalLoad
			}
			totalLoad += ts.TotalLoad
		}

		avgLoad := 0.0
		if len(spectrum) > 0 {
			avgLoad = totalLoad / float64(len(spectrum))
		}

		peakThreshold := maxLoad * 0.8
		for i, ts := range spectrum {
			if ts.TotalLoad >= peakThreshold {
				peakLocations = append(peakLocations, i)
			}
		}

		_ = fea.GenerateArchBridge(input.BridgeID, input.SpanLength, input.ArchRise, input.DeckWidth)

		totalMembers := 31
		fa := NewFatigueAnalysis(input.SpanLength, totalMembers)

		memberTypes := map[int]string{}
		for i := 1; i <= 10; i++ {
			memberTypes[i] = "arch_rib"
		}
		for i := 11; i <= 20; i++ {
			memberTypes[i] = "deck_beam"
		}
		for i := 21; i <= 31; i++ {
			memberTypes[i] = "vertical_post"
		}
		if len(memberTypes) < totalMembers {
			for i := len(memberTypes) + 1; i <= totalMembers; i++ {
				memberTypes[i] = "diagonal_brace"
			}
		}

		memberResults := make([]MemberFatigueResult, 0, totalMembers)
		for i := 0; i < totalMembers; i++ {
			memberID := i + 1
			memberType := memberTypes[memberID]
			stressHistory := fa.GenerateStressHistoryForSpectrum(loadValues, i, totalMembers)
			memberResult := fa.CalculateMemberFatigue(memberID, memberType, stressHistory, input.LoadCyclesPerDay)
			memberResults = append(memberResults, memberResult)
		}

		fatigueResult := fa.AggregateResults(memberResults)

		resultChan <- AsyncResult{
			LoadSpectrum:  spectrum,
			MaxLoad:       maxLoad,
			AvgLoad:       avgLoad,
			TotalAgents:   totalAgents,
			Duration:      input.Duration,
			FatigueResult: fatigueResult,
			PeakLocations: peakLocations,
			Error:         nil,
		}
	}()

	return resultChan
}
