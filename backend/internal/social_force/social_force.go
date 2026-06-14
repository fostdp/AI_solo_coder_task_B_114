package socialforce

import (
	movingloadsimulator "ancient-bridge-system/internal/moving_load_simulator"
)

type AgentType = movingloadsimulator.AgentType

const (
	AgentPedestrian = movingloadsimulator.AgentPedestrian
	AgentOxCart     = movingloadsimulator.AgentOxCart
	AgentHorseCart  = movingloadsimulator.AgentHorseCart
	AgentSedanChair = movingloadsimulator.AgentSedanChair
	AgentMilitary   = movingloadsimulator.AgentMilitary
	AgentPeddler    = movingloadsimulator.AgentPeddler
)

type Agent = movingloadsimulator.Agent

type SocialForceModel = movingloadsimulator.SocialForceModel

type LoadTimeStep = movingloadsimulator.LoadTimeStep

func NewSocialForceModel(spanLength float64) *SocialForceModel {
	return movingloadsimulator.NewSocialForceModel(spanLength)
}
