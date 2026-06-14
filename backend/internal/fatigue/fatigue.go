package fatigue

import (
	movingloadsimulator "ancient-bridge-system/internal/moving_load_simulator"
)

type FatigueAnalysis = movingloadsimulator.FatigueAnalysis

type MemberFatigueResult = movingloadsimulator.MemberFatigueResult

type FatigueResult = movingloadsimulator.FatigueResult

func NewFatigueAnalysis(spanLength float64, memberCount int) *FatigueAnalysis {
	return movingloadsimulator.NewFatigueAnalysis(spanLength, memberCount)
}
