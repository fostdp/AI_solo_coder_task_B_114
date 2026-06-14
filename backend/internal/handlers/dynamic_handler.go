package handlers

import (
	"net/http"

	"ancient-bridge-system/internal/database"
	"ancient-bridge-system/internal/fatigue"
	"ancient-bridge-system/internal/fea"
	socialforce "ancient-bridge-system/internal/social_force"

	"github.com/gin-gonic/gin"
)

type DynamicLoadHandler struct{}

func NewDynamicLoadHandler() *DynamicLoadHandler {
	return &DynamicLoadHandler{}
}

type SocialForceRequest struct {
	BridgeID      int     `json:"bridge_id" binding:"required"`
	Duration      float64 `json:"duration" binding:"min=10,max=600"`
	CrowdDensity  float64 `json:"crowd_density" binding:"min=0.1,max=5.0"`
	TimeStep      float64 `json:"time_step" binding:"min=0.05,max=1.0"`
	RandomSeed    int64   `json:"random_seed"`
	LoadCyclesPerDay float64 `json:"load_cycles_per_day"`
}

type SocialForceResponse struct {
	AnalysisID     int                          `json:"analysis_id"`
	BridgeID       int                          `json:"bridge_id"`
	LoadSpectrum   []socialforce.LoadTimeStep   `json:"load_spectrum"`
	MaxLoad        float64                      `json:"max_load"`
	AvgLoad        float64                      `json:"avg_load"`
	TotalAgents    int                          `json:"total_agents"`
	Duration       float64                      `json:"duration"`
	FatigueResult  *FatigueResultSummary        `json:"fatigue_result,omitempty"`
	PeakLocations  []int                        `json:"peak_load_locations"`
}

type FatigueResultSummary struct {
	TotalDamage        float64   `json:"total_damage"`
	MaxDamageMemberID  int       `json:"max_damage_member_id"`
	EstimatedLifeYears float64   `json:"estimated_life_years"`
	HotspotMembers     []int     `json:"hotspot_members"`
	CriticalLocations  []string  `json:"critical_locations"`
	Recommendations    []string  `json:"recommendations"`
}

func (h *DynamicLoadHandler) SocialForceAnalysis(c *gin.Context) {
	var req SocialForceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var bridge struct {
		SpanLength float64 `db:"span_length"`
		ArchRise   float64 `db:"arch_rise"`
		DeckWidth  float64 `db:"deck_width"`
	}
	err := database.DB.Get(&bridge, "SELECT span_length, arch_rise, deck_width FROM bridges WHERE bridge_id = $1", req.BridgeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bridge not found"})
		return
	}

	if req.Duration <= 0 {
		req.Duration = 60
	}
	if req.CrowdDensity <= 0 {
		req.CrowdDensity = 1.0
	}
	if req.TimeStep <= 0 {
		req.TimeStep = 0.2
	}
	if req.LoadCyclesPerDay <= 0 {
		req.LoadCyclesPerDay = 500
	}

	sfm := socialforce.NewSocialForceModel(bridge.SpanLength)
	if req.RandomSeed > 0 {
		sfm.SetRandomSeed(req.RandomSeed)
	}

	numAgents := int(req.CrowdDensity * bridge.SpanLength * 2)
	sfm.SpawnAgents(req.Duration, req.CrowdDensity)

	numSteps := int(req.Duration / req.TimeStep)
	spectrum := sfm.GetLoadSpectrum(numSteps, req.TimeStep)

	maxLoad := 0.0
	avgLoad := 0.0
	peakLocations := make([]int, 0)

	loadByPosition := make(map[int]float64)
	for _, ts := range spectrum {
		if ts.TotalLoad > maxLoad {
			maxLoad = ts.TotalLoad
		}
		avgLoad += ts.TotalLoad
		for pos, load := range ts.LoadDistribution {
			loadByPosition[pos] += load
		}
	}
	avgLoad /= float64(len(spectrum))

	for pos := 0; pos < 10; pos++ {
		if loadByPosition[pos] > avgLoad*1.5 {
			peakLocations = append(peakLocations, pos)
		}
	}

	var fatigueSummary *FatigueResultSummary

	bridgeModel := fea.GenerateArchBridge(req.BridgeID, bridge.SpanLength, bridge.ArchRise, bridge.DeckWidth)
	memberCount := len(bridgeModel.Structure.Members)

	fa := fatigue.NewFatigueAnalysis(bridge.SpanLength, memberCount)
	loadSpectrumValues := make([]float64, len(spectrum))
	for i, ts := range spectrum {
		loadSpectrumValues[i] = ts.TotalLoad
	}

	memberResults := make([]fatigue.MemberFatigueResult, 0, memberCount)
	for i := 0; i < memberCount; i++ {
		memberType := bridgeModel.MemberTypes[i+1]
		stressHistory := fa.GenerateStressHistoryForSpectrum(loadSpectrumValues, i, memberCount)
		result := fa.CalculateMemberFatigue(i+1, memberType, stressHistory, req.LoadCyclesPerDay)
		memberResults = append(memberResults, result)
	}

	fatigueResult := fa.AggregateResults(memberResults)
	recommendations := fa.GetFatigueRecommendations(fatigueResult)

	fatigueSummary = &FatigueResultSummary{
		TotalDamage:        fatigueResult.TotalDamage,
		MaxDamageMemberID:  fatigueResult.MaxDamageMemberID,
		EstimatedLifeYears: fatigueResult.EstimatedLifeYears,
		HotspotMembers:     fatigueResult.HotspotMembers,
		CriticalLocations:  fatigueResult.CriticalLocations,
		Recommendations:    recommendations,
	}

	response := SocialForceResponse{
		AnalysisID:    int(generateAnalysisID()),
		BridgeID:      req.BridgeID,
		LoadSpectrum:  spectrum,
		MaxLoad:       maxLoad,
		AvgLoad:       avgLoad,
		TotalAgents:   numAgents,
		Duration:      req.Duration,
		FatigueResult: fatigueSummary,
		PeakLocations: peakLocations,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

func (h *DynamicLoadHandler) GetAgentTypes(c *gin.Context) {
	types := []map[string]interface{}{
		{"type": "pedestrian", "name": "行人", "weight": 0.7, "velocity": 1.2},
		{"type": "ox_cart", "name": "牛车", "weight": 30.0, "velocity": 0.8},
		{"type": "horse_cart", "name": "马车", "weight": 50.0, "velocity": 1.5},
		{"type": "sedan_chair", "name": "轿子", "weight": 8.0, "velocity": 1.0},
		{"type": "military_convoy", "name": "军车车队", "weight": 80.0, "velocity": 1.8},
		{"type": "peddler", "name": "挑担商贩", "weight": 3.0, "velocity": 0.6},
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    types,
	})
}
