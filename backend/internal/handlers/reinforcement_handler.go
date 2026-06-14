package handlers

import (
	"net/http"

	"ancient-bridge-system/internal/database"
	retrofitoptimizer "ancient-bridge-system/internal/retrofit_optimizer"

	"github.com/gin-gonic/gin"
)

type ReinforcementHandler struct{}

func NewReinforcementHandler() *ReinforcementHandler {
	return &ReinforcementHandler{}
}

type ReinforceRequest struct {
	BridgeID       int     `json:"bridge_id" binding:"required"`
	TargetNodes    []int   `json:"target_nodes"`
	PopulationSize int     `json:"population_size"`
	MaxGenerations int     `json:"max_generations"`
	RandomSeed     int64   `json:"random_seed"`
	WeightPreference struct {
		Stiffness  float64 `json:"stiffness" binding:"min=0,max=1"`
		Strength   float64 `json:"strength" binding:"min=0,max=1"`
		Durability float64 `json:"durability" binding:"min=0,max=1"`
		Cost       float64 `json:"cost" binding:"min=0,max=1"`
		Heritage   float64 `json:"heritage" binding:"min=0,max=1"`
	} `json:"weight_preference"`
}

type ReinforceResponse struct {
	AnalysisID      int                               `json:"analysis_id"`
	BridgeID        int                               `json:"bridge_id"`
	OptimalSolutions []retrofitoptimizer.ReinforcementResult `json:"optimal_solutions"`
	Methods         []map[string]interface{}          `json:"retrofitoptimizer_methods"`
	Recommendations []string                          `json:"recommendations"`
}

func (h *ReinforcementHandler) RunOptimization(c *gin.Context) {
	var req ReinforceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var memberCount int
	err := database.DB.Get(&memberCount,
		"SELECT COUNT(*) FROM bridge_members WHERE bridge_id = $1", req.BridgeID)
	if err != nil {
		memberCount = 40
	}

	if req.PopulationSize <= 0 {
		req.PopulationSize = 80
	}
	if req.MaxGenerations <= 0 {
		req.MaxGenerations = 40
	}

	optimizer := retrofitoptimizer.NewMultiObjectiveOptimizer()
	optimizer.PopulationSize = req.PopulationSize
	optimizer.MaxGenerations = req.MaxGenerations

	if req.RandomSeed > 0 {
		optimizer.SetSeed(req.RandomSeed)
	}

	originalStiffness := 1.0
	originalStrength := 1.0

	results := optimizer.Optimize(originalStiffness, originalStrength, memberCount)

	if len(req.TargetNodes) > 0 {
		for i := range results {
			results[i].Params.TargetNodes = req.TargetNodes
		}
	}

	recommendations := generateReinforcementRecommendations(results)

	methods := retrofitoptimizer.GetReinforcementMethods()

	response := ReinforceResponse{
		AnalysisID:       int(generateAnalysisID()),
		BridgeID:         req.BridgeID,
		OptimalSolutions: results,
		Methods:          methods,
		Recommendations:  recommendations,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

func (h *ReinforcementHandler) GetMethods(c *gin.Context) {
	methods := retrofitoptimizer.GetReinforcementMethods()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    methods,
	})
}

func generateReinforcementRecommendations(results []retrofitoptimizer.ReinforcementResult) []string {
	recommendations := make([]string, 0)

	if len(results) == 0 {
		recommendations = append(recommendations, "未找到有效加固方案，请尝试调整参数")
		return recommendations
	}

	best := results[0]

	recommendations = append(recommendations,
		"optimal plan: plan#"+fmt.Sprintf("%d", best.PlanID)+", score="+formatScore(best.OverallScore))

	if best.StiffnessGainRate > 0.4 {
		recommendations = append(recommendations, "stiffness gain significant, suitable for deformation control scenarios")
	} else if best.StiffnessGainRate > 0.2 {
		recommendations = append(recommendations, "moderate stiffness gain, meets general reinforcement needs")
	}

	if best.HeritageImpactRate < 0.3 {
		recommendations = append(recommendations, "low heritage impact, conforms to cultural preservation principles")
	} else if best.HeritageImpactRate > 0.5 {
		recommendations = append(recommendations, "significant heritage impact, heritage authority assessment needed before construction")
	}

	if best.CostIncreaseFactor > 2.5 {
		recommendations = append(recommendations, "high cost, technical-economic comparison recommended")
	}

	if best.ConstructionComplexity > 0.7 {
		recommendations = append(recommendations, "complex construction, experienced professional team recommended")
	}

	if len(results) > 1 {
		recommendations = append(recommendations,
			fmt.Sprintf("%d Pareto optimal plans obtained, choose based on specific requirements", len(results)))
	}

	return recommendations
}

func formatScore(score float64) string {
	return string(rune(int(score*100+0.5))) + "分"
}
