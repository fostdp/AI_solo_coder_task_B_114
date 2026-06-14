package reinforcement

import (
	retrofitoptimizer "ancient-bridge-system/internal/retrofit_optimizer"
)

type ReinforcementMethod = retrofitoptimizer.ReinforcementMethod

const (
	MethodIronHoop     = retrofitoptimizer.MethodIronHoop
	MethodCFRP         = retrofitoptimizer.MethodCFRP
	MethodSteelPlate   = retrofitoptimizer.MethodSteelPlate
	MethodWoodenSplice = retrofitoptimizer.MethodWoodenSplice
	MethodCombined     = retrofitoptimizer.MethodCombined
)

type ReinforcementParams = retrofitoptimizer.ReinforcementParams

type ReinforcementResult = retrofitoptimizer.ReinforcementResult

type Individual = retrofitoptimizer.Individual

type MultiObjectiveOptimizer = retrofitoptimizer.MultiObjectiveOptimizer

func NewMultiObjectiveOptimizer() *MultiObjectiveOptimizer {
	return retrofitoptimizer.NewMultiObjectiveOptimizer()
}

func GetReinforcementMethods() []map[string]interface{} {
	return retrofitoptimizer.GetReinforcementMethods()
}
