package reinforcement

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

type ReinforcementMethod string

const (
	MethodIronHoop      ReinforcementMethod = "iron_hoop"
	MethodCFRP          ReinforcementMethod = "cfrp"
	MethodSteelPlate    ReinforcementMethod = "steel_plate"
	MethodWoodenSplice  ReinforcementMethod = "wooden_splice"
	MethodCombined      ReinforcementMethod = "combined"
)

type ReinforcementParams struct {
	Method          ReinforcementMethod
	CFRPThickness   float64
	CFRPLayers      int
	IronHoopCount   int
	IronHoopWidth   float64
	SteelPlateThickness float64
	WoodenSpliceLength float64
	TargetNodes     []int
}

type ReinforcementResult struct {
	方案ID             int
	Params             ReinforcementParams
	成本IncreaseFactor float64
	刚度提升率         float64
	强度提升率         float64
	耐久性提升率       float64
	施工复杂度         float64
	历史风貌影响度     float64
	综合评分           float64
	ParetoOptimal     bool
}

type MultiObjectiveOptimizer struct {
	PopulationSize    int
	MaxGenerations    int
	CrossoverRate     float64
	MutationRate      float64
	RandomSeed        int64
}

type Individual struct {
	Genes     []float64
	Fitness   []float64
	CrowdingDistance float64
	Dominated bool
	Rank      int
}

var methodCharacteristics = map[ReinforcementMethod]struct {
	BaseCostFactor     float64
	BaseStiffnessGain  float64
	BaseStrengthGain   float64
	BaseDurabilityGain float64
	BaseComplexity     float64
	HeritageImpact     float64
	Description        string
}{
	MethodIronHoop: {
		BaseCostFactor:     1.2,
		BaseStiffnessGain:  0.15,
		BaseStrengthGain:   0.20,
		BaseDurabilityGain: 0.25,
		BaseComplexity:     0.3,
		HeritageImpact:     0.4,
		Description:        "传统铁箍加固，历史真实性较好",
	},
	MethodCFRP: {
		BaseCostFactor:     2.5,
		BaseStiffnessGain:  0.35,
		BaseStrengthGain:   0.40,
		BaseDurabilityGain: 0.45,
		BaseComplexity:     0.6,
		HeritageImpact:     0.2,
		Description:        "碳纤维布加固，高效且隐蔽",
	},
	MethodSteelPlate: {
		BaseCostFactor:     1.8,
		BaseStiffnessGain:  0.30,
		BaseStrengthGain:   0.35,
		BaseDurabilityGain: 0.30,
		BaseComplexity:     0.5,
		HeritageImpact:     0.5,
		Description:        "钢板粘贴加固，刚度提升显著",
	},
	MethodWoodenSplice: {
		BaseCostFactor:     1.5,
		BaseStiffnessGain:  0.10,
		BaseStrengthGain:   0.15,
		BaseDurabilityGain: 0.20,
		BaseComplexity:     0.4,
		HeritageImpact:     0.1,
		Description:        "木榫拼接加固，历史真实性最佳",
	},
	MethodCombined: {
		BaseCostFactor:     3.0,
		BaseStiffnessGain:  0.50,
		BaseStrengthGain:   0.55,
		BaseDurabilityGain: 0.50,
		BaseComplexity:     0.8,
		HeritageImpact:     0.3,
		Description:        "铁箍+CFRP组合加固，综合性能最优",
	},
}

func NewMultiObjectiveOptimizer() *MultiObjectiveOptimizer {
	return &MultiObjectiveOptimizer{
		PopulationSize: 100,
		MaxGenerations: 50,
		CrossoverRate:  0.8,
		MutationRate:   0.1,
		RandomSeed:     time.Now().Unix(),
	}
}

func (mo *MultiObjectiveOptimizer) SetSeed(seed int64) {
	mo.RandomSeed = seed
	rand.Seed(seed)
}

func (mo *MultiObjectiveOptimizer) encodeParams(params ReinforcementParams) []float64 {
	genes := make([]float64, 8)

	methodIdx := 0.0
	switch params.Method {
	case MethodIronHoop:
		methodIdx = 0.0
	case MethodCFRP:
		methodIdx = 0.25
	case MethodSteelPlate:
		methodIdx = 0.5
	case MethodWoodenSplice:
		methodIdx = 0.75
	case MethodCombined:
		methodIdx = 1.0
	}
	genes[0] = methodIdx

	genes[1] = params.CFRPThickness / 5.0
	genes[2] = float64(params.CFRPLayers) / 10.0
	genes[3] = float64(params.IronHoopCount) / 20.0
	genes[4] = params.IronHoopWidth / 15.0
	genes[5] = params.SteelPlateThickness / 20.0
	genes[6] = params.WoodenSpliceLength / 3.0
	genes[7] = float64(len(params.TargetNodes)) / 44.0

	return genes
}

func (mo *MultiObjectiveOptimizer) decodeParams(genes []float64) ReinforcementParams {
	params := ReinforcementParams{}

	methodGene := genes[0]
	switch {
	case methodGene < 0.2:
		params.Method = MethodIronHoop
	case methodGene < 0.4:
		params.Method = MethodCFRP
	case methodGene < 0.6:
		params.Method = MethodSteelPlate
	case methodGene < 0.8:
		params.Method = MethodWoodenSplice
	default:
		params.Method = MethodCombined
	}

	params.CFRPThickness = math.Max(0.1, genes[1]*5.0)
	params.CFRPLayers = int(math.Max(1, genes[2]*10.0))
	params.IronHoopCount = int(math.Max(1, genes[3]*20.0))
	params.IronHoopWidth = math.Max(2.0, genes[4]*15.0)
	params.SteelPlateThickness = math.Max(1.0, genes[5]*20.0)
	params.WoodenSpliceLength = math.Max(0.3, genes[6]*3.0)

	return params
}

func (mo *MultiObjectiveOptimizer) evaluateIndividual(genes []float64, originalStiffness float64, originalStrength float64, memberCount int) ReinforcementResult {
	params := mo.decodeParams(genes)
	chars := methodCharacteristics[params.Method]

	costFactor := chars.BaseCostFactor *
		(1.0 + params.CFRPLayers*0.1 + params.IronHoopCount*0.05)

	stiffnessGain := chars.BaseStiffnessGain *
		(1.0 + params.CFRPThickness*0.2 + params.SteelPlateThickness*0.05)

	strengthGain := chars.BaseStrengthGain *
		(1.0 + params.CFRPLayers*0.15 + params.IronHoopWidth*0.03)

	durabilityGain := chars.BaseDurabilityGain *
		(1.0 + params.WoodenSpliceLength*0.2)

	complexity := chars.BaseComplexity *
		(1.0 + float64(params.IronHoopCount)*0.03 + float64(params.CFRPLayers)*0.05)

	heritageImpact := chars.HeritageImpact

	overallScore := (stiffnessGain*0.25 + strengthGain*0.25 + durabilityGain*0.2 +
		(1-complexity)*0.15 + (1-heritageImpact)*0.15)

	return ReinforcementResult{
		Params:             params,
		成本IncreaseFactor: costFactor,
		刚度提升率:         stiffnessGain,
		强度提升率:         strengthGain,
		耐久性提升率:       durabilityGain,
		施工复杂度:         complexity,
		历史风貌影响度:     heritageImpact,
		综合评分:           overallScore,
	}
}

func (mo *MultiObjectiveOptimizer) dominates(a, b []float64) bool {
	betterOrEqual := true
	strictlyBetter := false
	for i := range a {
		if a[i] < b[i] {
			betterOrEqual = false
			break
		}
		if a[i] > b[i] {
			strictlyBetter = true
		}
	}
	return betterOrEqual && strictlyBetter
}

func (mo *MultiObjectiveOptimizer) fastNonDominatedSort(population []Individual) [][]Individual {
	fronts := make([][]Individual, 0)
	dominationCount := make([]int, len(population))
	dominatedSet := make([][]int, len(population))

	for p := range population {
		dominatedSet[p] = make([]int, 0)
		dominationCount[p] = 0
		for q := range population {
			if p == q {
				continue
			}
			if mo.dominates(population[p].Fitness, population[q].Fitness) {
				dominatedSet[p] = append(dominatedSet[p], q)
			} else if mo.dominates(population[q].Fitness, population[p].Fitness) {
				dominationCount[p]++
			}
		}
		if dominationCount[p] == 0 {
			population[p].Rank = 1
			if len(fronts) == 0 {
				fronts = append(fronts, make([]Individual, 0))
			}
			fronts[0] = append(fronts[0], population[p])
		}
	}

	i := 0
	for len(fronts[i]) > 0 {
		nextFront := make([]Individual, 0)
		for _, p := range fronts[i] {
			for _, q := range dominatedSet[i] {
				dominationCount[q]--
				if dominationCount[q] == 0 {
					population[q].Rank = i + 2
					nextFront = append(nextFront, population[q])
				}
			}
		}
		i++
		if len(nextFront) > 0 {
			fronts = append(fronts, nextFront)
		}
	}

	return fronts
}

func (mo *MultiObjectiveOptimizer) crowdingDistance(front []Individual) {
	n := len(front)
	if n == 0 {
		return
	}

	for i := range front {
		front[i].CrowdingDistance = 0
	}

	numObjectives := len(front[0].Fitness)

	for obj := 0; obj < numObjectives; obj++ {
		sort.Slice(front, func(i, j int) bool {
			return front[i].Fitness[obj] < front[j].Fitness[obj]
		})

		front[0].CrowdingDistance = math.Inf(1)
		front[n-1].CrowdingDistance = math.Inf(1)

		min := front[0].Fitness[obj]
		max := front[n-1].Fitness[obj]
		range_ := max - min
		if range_ == 0 {
			range_ = 1
		}

		for i := 1; i < n-1; i++ {
			front[i].CrowdingDistance += (front[i+1].Fitness[obj] - front[i-1].Fitness[obj]) / range_
		}
	}
}

func (mo *MultiObjectiveOptimizer) tournamentSelection(population []Individual) Individual {
	a := rand.Intn(len(population))
	b := rand.Intn(len(population))

	if population[a].Rank < population[b].Rank {
		return population[a]
	} else if population[a].Rank > population[b].Rank {
		return population[b]
	}

	if population[a].CrowdingDistance > population[b].CrowdingDistance {
		return population[a]
	}
	return population[b]
}

func (mo *MultiObjectiveOptimizer) crossover(parent1, parent2 Individual) Individual {
	child := Individual{
		Genes: make([]float64, len(parent1.Genes)),
	}

	if rand.Float64() < mo.CrossoverRate {
		for i := range parent1.Genes {
			if rand.Float64() < 0.5 {
				child.Genes[i] = parent1.Genes[i]
			} else {
				child.Genes[i] = parent2.Genes[i]
			}
		}
	} else {
		copy(child.Genes, parent1.Genes)
	}

	return child
}

func (mo *MultiObjectiveOptimizer) mutate(individual Individual) {
	for i := range individual.Genes {
		if rand.Float64() < mo.MutationRate {
			individual.Genes[i] += rand.NormFloat64() * 0.1
			individual.Genes[i] = math.Max(0, math.Min(1, individual.Genes[i]))
		}
	}
}

func (mo *MultiObjectiveOptimizer) Optimize(originalStiffness float64, originalStrength float64, memberCount int) []ReinforcementResult {
	rand.Seed(mo.RandomSeed)

	population := make([]Individual, mo.PopulationSize)
	for i := range population {
		genes := make([]float64, 8)
		for j := range genes {
			genes[j] = rand.Float64()
		}
		result := mo.evaluateIndividual(genes, originalStiffness, originalStrength, memberCount)
		fitness := []float64{
			result.刚度提升率,
			result.强度提升率,
			result.耐久性提升率,
			1.0 - result.施工复杂度,
			1.0 - result.历史风貌影响度,
			1.0 / result.成本IncreaseFactor,
		}
		population[i] = Individual{
			Genes:   genes,
			Fitness: fitness,
		}
	}

	for gen := 0; gen < mo.MaxGenerations; gen++ {
		fronts := mo.fastNonDominatedSort(population)
		for _, front := range fronts {
			mo.crowdingDistance(front)
		}

		newPopulation := make([]Individual, 0, mo.PopulationSize)

		for _, front := range fronts {
			if len(newPopulation)+len(front) <= mo.PopulationSize {
				newPopulation = append(newPopulation, front...)
			} else {
				remaining := mo.PopulationSize - len(newPopulation)
				sort.Slice(front, func(i, j int) bool {
					return front[i].CrowdingDistance > front[j].CrowdingDistance
				})
				newPopulation = append(newPopulation, front[:remaining]...)
				break
			}
		}

		offspring := make([]Individual, 0, mo.PopulationSize)
		for len(offspring) < mo.PopulationSize {
			parent1 := mo.tournamentSelection(newPopulation)
			parent2 := mo.tournamentSelection(newPopulation)
			child := mo.crossover(parent1, parent2)
			mo.mutate(child)
			result := mo.evaluateIndividual(child.Genes, originalStiffness, originalStrength, memberCount)
			child.Fitness = []float64{
				result.刚度提升率,
				result.强度提升率,
				result.耐久性提升率,
				1.0 - result.施工复杂度,
				1.0 - result.历史风貌影响度,
				1.0 / result.成本IncreaseFactor,
			}
			offspring = append(offspring, child)
		}

		population = offspring
	}

	fronts := mo.fastNonDominatedSort(population)
	paretoFront := make([]Individual, 0)
	if len(fronts) > 0 {
		paretoFront = fronts[0]
	}

	results := make([]ReinforcementResult, 0, len(paretoFront))
	for i, ind := range paretoFront {
		result := mo.evaluateIndividual(ind.Genes, originalStiffness, originalStrength, memberCount)
		result.方案ID = i + 1
		result.ParetoOptimal = true
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].综合评分 > results[j].综合评分
	})

	return results
}

func GetReinforcementMethods() []map[string]interface{} {
	methods := make([]map[string]interface{}, 0)
	for m, chars := range methodCharacteristics {
		methods = append(methods, map[string]interface{}{
			"id":           string(m),
			"name":         getMethodName(m),
			"description":  chars.Description,
			"cost_factor":  chars.BaseCostFactor,
			"stiffness_gain": chars.BaseStiffnessGain,
			"strength_gain": chars.BaseStrengthGain,
			"durability_gain": chars.BaseDurabilityGain,
			"complexity":   chars.BaseComplexity,
			"heritage_impact": chars.HeritageImpact,
		})
	}
	return methods
}

func getMethodName(m ReinforcementMethod) string {
	names := map[ReinforcementMethod]string{
		MethodIronHoop:     "传统铁箍加固",
		MethodCFRP:         "碳纤维布(CFRP)加固",
		MethodSteelPlate:   "钢板粘贴加固",
		MethodWoodenSplice: "木榫拼接加固",
		MethodCombined:     "组合加固方案",
	}
	return names[m]
}
