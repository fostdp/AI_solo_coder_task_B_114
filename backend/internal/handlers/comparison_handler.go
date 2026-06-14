package handlers

import (
	"net/http"
	"strconv"

	historicalcomparison "ancient-bridge-system/internal/historical_comparison"

	"github.com/gin-gonic/gin"
)

type ComparisonHandler struct{}

func NewComparisonHandler() *ComparisonHandler {
	return &ComparisonHandler{}
}

type CompareRequest struct {
	BridgeAID int `json:"bridge_a_id" binding:"required"`
	BridgeBID int `json:"bridge_b_id" binding:"required"`
}

type CompareResponse struct {
	AnalysisID       int                                         `json:"analysis_id"`
	BridgeA          historicalcomparison.HistoricalBridge       `json:"bridge_a"`
	BridgeB          historicalcomparison.HistoricalBridge       `json:"bridge_b"`
	MetricsA         historicalcomparison.EfficiencyMetrics      `json:"metrics_a"`
	MetricsB         historicalcomparison.EfficiencyMetrics      `json:"metrics_b"`
	RadarData        []historicalcomparison.RadarPoint           `json:"radar_data"`
	AdvantagesA      []string                                    `json:"advantages_a"`
	AdvantagesB      []string                                    `json:"advantages_b"`
	HistoricalNotes  []string                                    `json:"historical_notes"`
	TechEvolution    []historicalcomparison.TechEvolutionPoint   `json:"tech_evolution"`
	NormalizedScores map[string]map[string]float64               `json:"normalized_scores"`
}

func (h *ComparisonHandler) CompareBridges(c *gin.Context) {
	var req CompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hc := historicalcomparison.NewHistoricalComparison()
	allBridges := hc.GetAllBridges()

	var bridgeA, bridgeB historicalcomparison.HistoricalBridge
	foundA, foundB := false, false

	for _, b := range allBridges {
		if b.ID == req.BridgeAID {
			bridgeA = b
			foundA = true
		}
		if b.ID == req.BridgeBID {
			bridgeB = b
			foundB = true
		}
	}

	if !foundA || !foundB {
		c.JSON(http.StatusNotFound, gin.H{"error": "One or both bridges not found"})
		return
	}

	result := hc.CompareBridges(bridgeA, bridgeB)

	normalizedScores := make(map[string]map[string]float64)
	for metric, scores := range result.NormalizedScores {
		normalizedScores[metric] = map[string]float64{
			string(bridgeA.Dynasty): scores[bridgeA.Dynasty],
			string(bridgeB.Dynasty): scores[bridgeB.Dynasty],
		}
	}

	response := CompareResponse{
		AnalysisID:       int(generateAnalysisID()),
		BridgeA:          result.BridgeA,
		BridgeB:          result.BridgeB,
		MetricsA:         result.MetricsA,
		MetricsB:         result.MetricsB,
		RadarData:        result.RadarData,
		AdvantagesA:      result.AdvantagesA,
		AdvantagesB:      result.AdvantagesB,
		HistoricalNotes:  result.HistoricalNotes,
		TechEvolution:    result.TechEvolution,
		NormalizedScores: normalizedScores,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

func (h *ComparisonHandler) GetHistoricalBridges(c *gin.Context) {
	dynastyParam := c.Query("dynasty")

	hc := historicalcomparison.NewHistoricalComparison()
	var bridges []historicalcomparison.HistoricalBridge

	if dynastyParam != "" {
		bridges = hc.GetBridgesByDynasty(historicalcomparison.Dynasty(dynastyParam))
	} else {
		bridges = hc.GetAllBridges()
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    bridges,
		"count":   len(bridges),
	})
}

func (h *ComparisonHandler) GetDynasties(c *gin.Context) {
	dynasties := []map[string]interface{}{
		{"id": "han_jin", "name": "汉晋", "period": "公元前206年-公元420年", "description": "简支木梁桥成熟时期"},
		{"id": "tang", "name": "唐代", "period": "公元618年-907年", "description": "木拱技术萌芽时期"},
		{"id": "song", "name": "宋代", "period": "公元960年-1279年", "description": "贯木拱技术鼎盛时期"},
		{"id": "ming", "name": "明代", "period": "公元1368年-1644年", "description": "木拱廊桥发展时期"},
		{"id": "qing", "name": "清代", "period": "公元1644年-1912年", "description": "工艺精细化与装饰时期"},
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    dynasties,
	})
}

func (h *ComparisonHandler) GetTechEvolution(c *gin.Context) {
	hc := historicalcomparison.NewHistoricalComparison()
	allBridges := hc.GetAllBridges()
	if len(allBridges) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough bridges"})
		return
	}
	result := hc.CompareBridges(allBridges[0], allBridges[1])
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result.TechEvolution,
	})
}
