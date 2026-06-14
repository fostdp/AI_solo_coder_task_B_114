package socialforce

import (
	"math"
	"math/rand"
	"time"
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

type LoadTimeStep struct {
	TimeStep         int
	TimeSeconds      float64
	TotalLoad        float64
	ActiveAgentCount int
	LoadDistribution map[int]float64
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
