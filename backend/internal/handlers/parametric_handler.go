package handlers

import (
	"net/http"
	"strconv"

	"ancient-bridge-system/internal/database"
	"ancient-bridge-system/internal/fea"

	"github.com/gin-gonic/gin"
)

type ParametricHandler struct{}

func NewParametricHandler() *ParametricHandler {
	return &ParametricHandler{}
}

type ParametricRequest struct {
	BridgeID   int     `json:"bridge_id" binding:"required"`
	SpanLength float64 `json:"span_length" binding:"required,min=10,max=50"`
	ArchRise   float64 `json:"arch_rise" binding:"required,min=2,max=15"`
	DeckWidth  float64 `json:"deck_width" binding:"required,min=3,max=12"`
	LoadValue  float64 `json:"load_value" binding:"min=0,max=500"`
}

type ParametricResponse struct {
	AnalysisID int                        `json:"analysis_id"`
	Valid      bool                       `json:"valid"`
	Message    string                     `json:"message"`
	Result     *fea.ParametricAnalysisResult `json:"result,omitempty"`
	Options    []fea.GeometryOption       `json:"options,omitempty"`
}

func (h *ParametricHandler) GetGeometryOptions(c *gin.Context) {
	bridgeIDStr := c.Param("bridge_id")
	bridgeID, err := strconv.Atoi(bridgeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bridge ID"})
		return
	}

	var bridge struct {
		SpanLength float64 `db:"span_length"`
		ArchRise   float64 `db:"arch_rise"`
		DeckWidth  float64 `db:"deck_width"`
	}
	err = database.DB.Get(&bridge,
		"SELECT span_length, arch_rise, deck_width FROM bridges WHERE bridge_id = $1", bridgeID)
	if err != nil {
		pb := fea.NewParametricBridge(bridgeID, 25.6, 5.8, 6.5)
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data":    pb.GetGeometryOptions(),
		})
		return
	}

	pb := fea.NewParametricBridge(bridgeID, bridge.SpanLength, bridge.ArchRise, bridge.DeckWidth)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    pb.GetGeometryOptions(),
	})
}

func (h *ParametricHandler) Analyze(c *gin.Context) {
	var req ParametricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pb := fea.NewParametricBridge(req.BridgeID, 25.6, 5.8, 6.5)

	valid, message := pb.ValidateParams(req.SpanLength, req.ArchRise, req.DeckWidth)

	if req.LoadValue <= 0 {
		req.LoadValue = 50.0
	}

	var result *fea.ParametricAnalysisResult
	if valid {
		analysisResult := fea.CalculateParametricAnalysis(
			req.BridgeID,
			req.SpanLength,
			req.ArchRise,
			req.DeckWidth,
			req.LoadValue,
		)
		result = &analysisResult
	}

	response := ParametricResponse{
		AnalysisID: int(generateAnalysisID()),
		Valid:      valid,
		Message:    message,
		Result:     result,
	}

	if !valid {
		response.Options = pb.GetGeometryOptions()
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

func (h *ParametricHandler) BatchAnalyze(c *gin.Context) {
	type BatchPoint struct {
		SpanLength float64 `json:"span_length"`
		ArchRise   float64 `json:"arch_rise"`
		DeckWidth  float64 `json:"deck_width"`
	}

	type BatchRequest struct {
		BridgeID  int          `json:"bridge_id" binding:"required"`
		Points    []BatchPoint `json:"points" binding:"required,min=1,max=50"`
		LoadValue float64      `json:"load_value"`
	}

	var req BatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.LoadValue <= 0 {
		req.LoadValue = 50.0
	}

	results := make([]fea.ParametricAnalysisResult, 0, len(req.Points))
	pb := fea.NewParametricBridge(req.BridgeID, 25.6, 5.8, 6.5)

	for _, point := range req.Points {
		valid, _ := pb.ValidateParams(point.SpanLength, point.ArchRise, point.DeckWidth)
		if valid {
			result := fea.CalculateParametricAnalysis(
				req.BridgeID,
				point.SpanLength,
				point.ArchRise,
				point.DeckWidth,
				req.LoadValue,
			)
			results = append(results, result)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"analysis_id": generateAnalysisID(),
			"results":     results,
			"count":       len(results),
		},
	})
}

func (h *ParametricHandler) GetDesignRecommendations(c *gin.Context) {
	spanStr := c.Query("span")
	riseStr := c.Query("rise")

	span, _ := strconv.ParseFloat(spanStr, 64)
	rise, _ := strconv.ParseFloat(riseStr, 64)

	recommendations := make([]string, 0)

	if span > 0 && rise > 0 {
		ratio := rise / span

		if ratio < 0.1 {
			recommendations = append(recommendations, "矢跨比较小，桥面较平缓，有利于通行，但拱肋受力较大")
		} else if ratio > 0.25 {
			recommendations = append(recommendations, "矢跨比较大，拱脚水平推力较大，需加强桥台设计")
		} else {
			recommendations = append(recommendations, "矢跨比适中，受力与通行性能均衡")
		}

		if span > 40 {
			recommendations = append(recommendations, "大跨度设计，建议采用多节拱组合结构")
		} else if span > 25 {
			recommendations = append(recommendations, "中等跨度，适合采用贯木拱技术")
		}

		if span > 20 && ratio < 0.12 {
			recommendations = append(recommendations, "建议增加矢高或设置预拱度，改善结构受力")
		}
	}

	recommendations = append(recommendations, "木材选择：推荐使用杉木或松木，顺纹抗压强度满足要求")
	recommendations = append(recommendations, "节点设计：拱脚节点采用半榫结合，提高抗剪能力")
	recommendations = append(recommendations, "构造要求：拱肋截面高跨比建议不小于1:30")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    recommendations,
	})
}
