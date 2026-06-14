package handlers

import (
	"net/http"

	"ancient-bridge-system/internal/database"
	"ancient-bridge-system/internal/reinforcement"

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
	OptimalSolutions []reinforcement.ReinforcementResult `json:"optimal_solutions"`
	Methods         []map[string]interface{}          `json:"reinforcement_methods"`
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

	optimizer := reinforcement.NewMultiObjectiveOptimizer()
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

	methods := reinforcement.GetReinforcementMethods()

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
	methods := reinforcement.GetReinforcementMethods()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    methods,
	})
}

func generateReinforcementRecommendations(results []reinforcement.ReinforcementResult) []string {
	recommendations := make([]string, 0)

	if len(results) == 0 {
		recommendations = append(recommendations, "未找到有效加固方案，请尝试调整参数")
		return recommendations
	}

	best := results[0]

	recommendations = append(recommendations,
		"最优综合方案：方案"+string(rune('0'+best.方案ID))+"，综合评分"+formatScore(best.综合评分))

	if best.刚度提升率 > 0.4 {
		recommendations = append(recommendations, "刚度提升显著，适合对变形控制要求高的场景")
	} else if best.刚度提升率 > 0.2 {
		recommendations = append(recommendations, "刚度提升适中，可满足一般加固需求")
	}

	if best.历史风貌影响度 < 0.3 {
		recommendations = append(recommendations, "历史风貌影响较小，符合文物保护原则")
	} else if best.历史风貌影响度 > 0.5 {
		recommendations = append(recommendations, "对历史风貌有一定影响，施工前需文物部门评估")
	}

	if best.成本IncreaseFactor > 2.5 {
		recommendations = append(recommendations, "成本较高，建议进行技术经济比较")
	}

	if best.施工复杂度 > 0.7 {
		recommendations = append(recommendations, "施工复杂，建议选择有经验的专业队伍")
	}

	if len(results) > 1 {
		recommendations = append(recommendations,
			"共获得"+string(rune('0'+len(results)))+"个Pareto最优方案，可根据具体需求权衡选择")
	}

	return recommendations
}

func formatScore(score float64) string {
	return string(rune(int(score*100+0.5))) + "分"
}
