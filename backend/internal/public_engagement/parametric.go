package publicengagement

import (
	"fmt"
	"math"
)

type JointType string

const (
	JointMortiseTenon JointType = "mortise_tenon"
	JointDovetail     JointType = "dovetail"
	JointLap          JointType = "lap"
	JointRigid        JointType = "rigid"
	JointPin          JointType = "pin"
)

type SemiRigidJoint struct {
	Ktheta float64
	Kt     float64
	Kaxial float64
	Rtheta float64
	Type   JointType
}

var defaultJointStiffness = map[JointType]SemiRigidJoint{
	JointMortiseTenon: {Ktheta: 0.65, Kt: 0.80, Kaxial: 0.90, Rtheta: 1.0, Type: JointMortiseTenon},
	JointDovetail:     {Ktheta: 0.75, Kt: 0.85, Kaxial: 0.70, Rtheta: 1.0, Type: JointDovetail},
	JointLap:          {Ktheta: 0.55, Kt: 0.60, Kaxial: 0.50, Rtheta: 1.0, Type: JointLap},
	JointRigid:        {Ktheta: 1.00, Kt: 1.00, Kaxial: 1.00, Rtheta: 1.0, Type: JointRigid},
	JointPin:          {Ktheta: 0.02, Kt: 0.95, Kaxial: 0.98, Rtheta: 1.0, Type: JointPin},
}

type Node struct {
	ID       int
	X, Y     float64
	UX, UY   float64
	RZ       float64
	FixedX   bool
	FixedY   bool
	FixedR   bool
	Joint    SemiRigidJoint
	Ktheta   float64
	KshearX  float64
	KshearY  float64
}

type Member struct {
	ID         int
	Start      *Node
	End        *Node
	E          float64
	A          float64
	I          float64
	Length     float64
	Angle      float64
	JointStart JointType
	JointEnd   JointType
}

type TrussStructure struct {
	Nodes          []*Node
	Members        []*Member
	NodeMap        map[int]*Node
	MemberMap      map[int]*Member
	K              [][]float64
	F              []float64
	U              []float64
	dofs           int
	jointStiffness map[JointType]SemiRigidJoint
	useSemiRigid   bool
	MemberTypes    map[int]string
}

func NewTrussStructure() *TrussStructure {
	return &TrussStructure{
		NodeMap:        make(map[int]*Node),
		MemberMap:      make(map[int]*Member),
		jointStiffness: defaultJointStiffness,
		useSemiRigid:   true,
		MemberTypes:    make(map[int]string),
	}
}

func (ts *TrussStructure) AddNode(id int, x, y float64) *Node {
	node := &Node{
		ID:      id,
		X:       x,
		Y:       y,
		Joint:   ts.jointStiffness[JointMortiseTenon],
		Ktheta:  0,
		KshearX: 0,
		KshearY: 0,
	}
	ts.Nodes = append(ts.Nodes, node)
	ts.NodeMap[id] = node
	return node
}

func (ts *TrussStructure) SetNodeConstraint(id int, fx, fy, fr bool) {
	if node, exists := ts.NodeMap[id]; exists {
		node.FixedX = fx
		node.FixedY = fy
		node.FixedR = fr
	}
}

func (ts *TrussStructure) AddMember(id int, startID, endID int, e, a, i float64, jointTypes ...JointType) *Member {
	start := ts.NodeMap[startID]
	end := ts.NodeMap[endID]
	if start == nil || end == nil {
		return nil
	}

	dx := end.X - start.X
	dy := end.Y - start.Y
	length := math.Sqrt(dx*dx + dy*dy)
	angle := math.Atan2(dy, dx)

	jointStart := JointMortiseTenon
	jointEnd := JointMortiseTenon
	if len(jointTypes) >= 1 {
		jointStart = jointTypes[0]
	}
	if len(jointTypes) >= 2 {
		jointEnd = jointTypes[1]
	}

	member := &Member{
		ID:         id,
		Start:      start,
		End:        end,
		E:          e,
		A:          a,
		I:          i,
		Length:     length,
		Angle:      angle,
		JointStart: jointStart,
		JointEnd:   jointEnd,
	}

	ts.Members = append(ts.Members, member)
	ts.MemberMap[id] = member
	return member
}

func (m *Member) LocalStiffness() [6][6]float64 {
	ea := m.E * m.A
	ei := m.E * m.I
	l := m.Length
	l2 := l * l
	l3 := l * l * l

	return [6][6]float64{
		{ea / l, 0, 0, -ea / l, 0, 0},
		{0, 12 * ei / l3, 6 * ei / l2, 0, -12 * ei / l3, 6 * ei / l2},
		{0, 6 * ei / l2, 4 * ei / l, 0, -6 * ei / l2, 2 * ei / l},
		{-ea / l, 0, 0, ea / l, 0, 0},
		{0, -12 * ei / l3, -6 * ei / l2, 0, 12 * ei / l3, -6 * ei / l2},
		{0, 6 * ei / l2, 2 * ei / l, 0, -6 * ei / l2, 4 * ei / l},
	}
}

func (m *Member) LocalStiffnessSemiRigid() [6][6]float64 {
	k0 := m.LocalStiffness()
	if len(m.Start.Joint.Type) == 0 || len(m.End.Joint.Type) == 0 {
		return k0
	}

	js := m.Start.Joint
	je := m.End.Joint
	l := m.Length
	ei := m.E * m.I

	alphaS := ei / l
	alphaE := ei / l

	var k1, k4 float64
	if js.Ktheta > 0 && js.Ktheta < 1.0 {
		k1 = 4 * alphaS / (1 + 4*alphaS/js.Ktheta/ei)
	} else {
		k1 = k0[2][2]
	}

	if je.Ktheta > 0 && je.Ktheta < 1.0 {
		k4 = 4 * alphaE / (1 + 4*alphaE/je.Ktheta/ei)
	} else {
		k4 = k0[5][5]
	}

	shearFactor := (js.Kt + je.Kt) / 2.0
	axialFactor := (js.Kaxial + je.Kaxial) / 2.0

	var k [6][6]float64
	for i := 0; i < 6; i++ {
		for j := 0; j < 6; j++ {
			k[i][j] = k0[i][j]
		}
	}

	rowFactor := []float64{axialFactor, shearFactor, js.Ktheta, axialFactor, shearFactor, je.Ktheta}
	colFactor := []float64{axialFactor, shearFactor, js.Ktheta, axialFactor, shearFactor, je.Ktheta}
	for i := 0; i < 6; i++ {
		for j := 0; j < 6; j++ {
			avg := math.Sqrt(rowFactor[i] * colFactor[j])
			k[i][j] *= avg
		}
	}

	k[2][2] = k1
	k[5][5] = k4
	k[2][5] = k0[2][5] * math.Sqrt(js.Ktheta*je.Ktheta)
	k[5][2] = k[2][5]

	return k
}

func (m *Member) TransformationMatrix() [6][6]float64 {
	c := math.Cos(m.Angle)
	s := math.Sin(m.Angle)

	return [6][6]float64{
		{c, s, 0, 0, 0, 0},
		{-s, c, 0, 0, 0, 0},
		{0, 0, 1, 0, 0, 0},
		{0, 0, 0, c, s, 0},
		{0, 0, 0, -s, c, 0},
		{0, 0, 0, 0, 0, 1},
	}
}

func (m *Member) GlobalStiffness() [6][6]float64 {
	var kLocal [6][6]float64
	kLocal = m.LocalStiffnessSemiRigid()
	T := m.TransformationMatrix()
	var kGlobal [6][6]float64

	for i := 0; i < 6; i++ {
		for j := 0; j < 6; j++ {
			for k := 0; k < 6; k++ {
				for l := 0; l < 6; l++ {
					kGlobal[i][j] += T[k][i] * kLocal[k][l] * T[l][j]
				}
			}
		}
	}

	return kGlobal
}

func (ts *TrussStructure) AssembleStiffness() {
	ts.dofs = len(ts.Nodes) * 3

	ts.K = make([][]float64, ts.dofs)
	for i := range ts.K {
		ts.K[i] = make([]float64, ts.dofs)
	}

	for _, member := range ts.Members {
		kGlobal := member.GlobalStiffness()

		startIdx := (member.Start.ID - 1) * 3
		endIdx := (member.End.ID - 1) * 3

		indices := []int{
			startIdx, startIdx + 1, startIdx + 2,
			endIdx, endIdx + 1, endIdx + 2,
		}

		for i := 0; i < 6; i++ {
			for j := 0; j < 6; j++ {
				ts.K[indices[i]][indices[j]] += kGlobal[i][j]
			}
		}
	}

	if ts.useSemiRigid {
		ts.assembleNodeSprings()
	}
}

func (ts *TrussStructure) assembleNodeSprings() {
	for _, node := range ts.Nodes {
		if node.FixedR && node.FixedX && node.FixedY {
			continue
		}
		nodeIdx := (node.ID - 1) * 3

		connectedMembers := 0
		totalEIL := 0.0
		for _, member := range ts.Members {
			if member.Start.ID == node.ID || member.End.ID == node.ID {
				connectedMembers++
				if member.I > 0 {
					totalEIL += member.E * member.I / member.Length
				}
			}
		}
		if connectedMembers == 0 {
			continue
		}

		if !node.FixedX && node.KshearX > 0 {
			ts.K[nodeIdx][nodeIdx] += node.KshearX
		}
		if !node.FixedY && node.KshearY > 0 {
			ts.K[nodeIdx+1][nodeIdx+1] += node.KshearY
		}

		ktheta := node.Ktheta
		if ktheta == 0 && node.Joint.Ktheta > 0 {
			ktheta = node.Joint.Ktheta * totalEIL * 0.1
			if ktheta < 0 {
				ktheta = 0
			}
		}

		if !node.FixedR && ktheta > 0 {
			ts.K[nodeIdx+2][nodeIdx+2] += ktheta
		}
	}
}

type NodeLoad struct {
	NodeID int
	FX     float64
	FY     float64
	MZ     float64
}

func (ts *TrussStructure) ApplyLoads(loads map[int]float64) {
	ts.F = make([]float64, ts.dofs)
	ts.U = make([]float64, ts.dofs)

	for nodeID, fy := range loads {
		nodeIdx := (nodeID - 1) * 3
		if nodeIdx < ts.dofs {
			ts.F[nodeIdx+1] = fy
		}
	}
}

func (ts *TrussStructure) Solve() error {
	freeDofs := ts.getFreeDofs()
	nFree := len(freeDofs)

	Kff := make([][]float64, nFree)
	for i := range Kff {
		Kff[i] = make([]float64, nFree)
	}

	Ff := make([]float64, nFree)

	for i, fi := range freeDofs {
		Ff[i] = ts.F[fi]
		for j, fj := range freeDofs {
			Kff[i][j] = ts.K[fi][fj]
		}
	}

	Uf, err := gaussElimination(Kff, Ff)
	if err != nil {
		return err
	}

	for i, fi := range freeDofs {
		ts.U[fi] = Uf[i]
	}

	for i, node := range ts.Nodes {
		idx := i * 3
		node.UX = ts.U[idx]
		node.UY = ts.U[idx+1]
		node.RZ = ts.U[idx+2]
	}

	return nil
}

func (ts *TrussStructure) getFreeDofs() []int {
	var free []int
	for i, node := range ts.Nodes {
		idx := i * 3
		if !node.FixedX {
			free = append(free, idx)
		}
		if !node.FixedY {
			free = append(free, idx+1)
		}
		if !node.FixedR {
			free = append(free, idx+2)
		}
	}
	return free
}

func gaussElimination(A [][]float64, B []float64) ([]float64, error) {
	n := len(B)
	aug := make([][]float64, n)
	for i := range aug {
		aug[i] = make([]float64, n+1)
		copy(aug[i][:n], A[i])
		aug[i][n] = B[i]
	}

	for col := 0; col < n; col++ {
		pivotRow := col
		for row := col + 1; row < n; row++ {
			if math.Abs(aug[row][col]) > math.Abs(aug[pivotRow][col]) {
				pivotRow = row
			}
		}
		aug[col], aug[pivotRow] = aug[pivotRow], aug[col]

		if math.Abs(aug[col][col]) < 1e-10 {
			return nil, nil
		}

		for row := col + 1; row < n; row++ {
			factor := aug[row][col] / aug[col][col]
			for j := col; j <= n; j++ {
				aug[row][j] -= factor * aug[col][j]
			}
		}
	}

	x := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		sum := 0.0
		for j := i + 1; j < n; j++ {
			sum += aug[i][j] * x[j]
		}
		x[i] = (aug[i][n] - sum) / aug[i][i]
	}

	return x, nil
}

type MemberForces struct {
	MemberID      int
	AxialForce    float64
	ShearForce    float64
	BendingMoment float64
	AxialStress   float64
	BendingStress float64
	NormalStress  float64
	CombinedStress float64
	StressRatio   float64
	Length        float64
}

func (ts *TrussStructure) CalculateMemberForces() []MemberForces {
	var forces []MemberForces

	for _, member := range ts.Members {
		uLocal := ts.getMemberLocalDisplacements(member)
		fLocal := ts.calculateLocalForces(member, uLocal)

		axialForce := fLocal[0]
		shearForce := fLocal[1]
		bendingMoment := fLocal[2]

		axialStress := axialForce / member.A
		bendingStress := 0.0
		if member.I > 0 {
			bendingStress = bendingMoment * (member.Length / 12) / member.I
		}
		normalStress := axialStress
		combinedStress := normalStress + bendingStress

		allowableStress := 8.5
		memberType := ts.MemberTypes[member.ID]
		if memberType == "deck_beam" {
			allowableStress = 8.5
		} else if memberType == "arch_rib" {
			allowableStress = 8.5
		}
		stressRatio := 0.0
		if allowableStress > 0 {
			stressRatio = math.Abs(combinedStress) / allowableStress
		}

		forces = append(forces, MemberForces{
			MemberID:       member.ID,
			AxialForce:     axialForce,
			ShearForce:     shearForce,
			BendingMoment:  bendingMoment,
			AxialStress:    axialStress,
			BendingStress:  bendingStress,
			NormalStress:   normalStress,
			CombinedStress: combinedStress,
			StressRatio:    stressRatio,
			Length:         member.Length,
		})
	}

	return forces
}

func (ts *TrussStructure) getMemberLocalDisplacements(m *Member) [6]float64 {
	startIdx := (m.Start.ID - 1) * 3
	endIdx := (m.End.ID - 1) * 3

	uGlobal := [6]float64{
		ts.U[startIdx], ts.U[startIdx+1], ts.U[startIdx+2],
		ts.U[endIdx], ts.U[endIdx+1], ts.U[endIdx+2],
	}

	T := m.TransformationMatrix()
	var uLocal [6]float64
	for i := 0; i < 6; i++ {
		for j := 0; j < 6; j++ {
			uLocal[i] += T[i][j] * uGlobal[j]
		}
	}

	return uLocal
}

func (ts *TrussStructure) calculateLocalForces(m *Member, uLocal [6]float64) [6]float64 {
	kLocal := m.LocalStiffness()
	var fLocal [6]float64

	for i := 0; i < 6; i++ {
		for j := 0; j < 6; j++ {
			fLocal[i] += kLocal[i][j] * uLocal[j]
		}
	}

	return fLocal
}

type NodeDisplacement struct {
	NodeID    int
	UX        float64
	UY        float64
	TotalDisp float64
}

func (ts *TrussStructure) GetNodeDisplacements() []NodeDisplacement {
	var displacements []NodeDisplacement

	for _, node := range ts.Nodes {
		totalDisp := math.Sqrt(node.UX*node.UX + node.UY*node.UY)
		displacements = append(displacements, NodeDisplacement{
			NodeID:    node.ID,
			UX:        node.UX,
			UY:        node.UY,
			TotalDisp: totalDisp,
		})
	}

	return displacements
}

type YingzaoSpec struct {
	ComponentType     string
	GradeLevel        string
	MaterialGrade     string
	MaxSpanRatio      float64
	MinSectionModulus float64
	AllowableStress   float64
	SafetyFactor      float64
}

type YingzaoComparison struct {
	MemberID          int
	MemberType        string
	ActualStress      float64
	AllowableStress   float64
	StressRatio       float64
	MaxSpanRatio      float64
	ActualSpanRatio   float64
	SectionModulus    float64
	MinSectionModulus float64
	Compliant         bool
	GradeUsed         string
	SpecGrade         string
}

func (ts *TrussStructure) CompareWithYingzaoFashi() []YingzaoComparison {
	var comparisons []YingzaoComparison
	memberForces := ts.CalculateMemberForces()
	allowableStressDefault := 8.5
	maxSpanRatioDefault := 12.0
	minSectionModulusDefault := 300.0

	for _, mf := range memberForces {
		memberType := ts.MemberTypes[mf.MemberID]
		member := ts.MemberMap[mf.MemberID]
		if member == nil {
			continue
		}

		sectionModulus := 0.0
		if member.A > 0 {
			sectionModulus = member.I / (member.A / 6.0)
		}
		actualSpanRatio := 0.0
		if math.Sqrt(member.A) > 0 {
			actualSpanRatio = member.Length / math.Sqrt(member.A)
		}

		allowableStress := allowableStressDefault
		maxSpanRatio := maxSpanRatioDefault
		minSectionModulus := minSectionModulusDefault

		stressRatio := 0.0
		if allowableStress > 0 {
			stressRatio = math.Abs(mf.CombinedStress) / allowableStress
		}
		compliant := stressRatio <= 1.0 && actualSpanRatio <= maxSpanRatio

		comparisons = append(comparisons, YingzaoComparison{
			MemberID:          mf.MemberID,
			MemberType:        memberType,
			ActualStress:      mf.CombinedStress,
			AllowableStress:   allowableStress,
			StressRatio:       stressRatio,
			MaxSpanRatio:      maxSpanRatio,
			ActualSpanRatio:   actualSpanRatio,
			SectionModulus:    sectionModulus,
			MinSectionModulus: minSectionModulus,
			Compliant:         compliant,
			SpecGrade:         "",
		})
	}

	return comparisons
}

type BridgeModel struct {
	Structure   *TrussStructure
	BridgeName  string
	SpanLength  float64
	ArchRise    float64
	DeckWidth   float64
	MemberTypes map[int]string
}

func GenerateArchBridge(bridgeID int, spanLength, archRise, deckWidth float64) *BridgeModel {
	structure := NewTrussStructure()
	memberTypes := make(map[int]string)

	numArchSegments := 10

	archRadius := (spanLength*spanLength/4 + archRise*archRise) / (2 * archRise)
	archCenterY := archRise - archRadius
	archCenterX := spanLength / 2

	nodeID := 1

	for i := 0; i <= numArchSegments; i++ {
		angle := -math.Pi/2 + math.Pi*float64(i)/float64(numArchSegments)
		x := archCenterX + archRadius*math.Cos(angle)
		y := archCenterY + archRadius*math.Sin(angle)
		structure.AddNode(nodeID, x, y)
		nodeID++
	}

	for i := 0; i <= numArchSegments; i++ {
		x := float64(i) * spanLength / float64(numArchSegments)
		y := 0.0
		structure.AddNode(nodeID, x, y)
		nodeID++
	}

	structure.SetNodeConstraint(1, true, true, true)
	structure.SetNodeConstraint(numArchSegments+1, true, true, true)

	memberID := 1
	e := 10000.0
	aArch := 0.12
	iArch := 0.0015
	aDeck := 0.08
	iDeck := 0.0008
	aVertical := 0.06
	iVertical := 0.0004

	for i := 1; i <= numArchSegments; i++ {
		startID := i
		endID := i + 1
		member := structure.AddMember(memberID, startID, endID, e, aArch, iArch)
		if member != nil {
			memberTypes[memberID] = "arch_rib"
			structure.MemberTypes[memberID] = "arch_rib"
		}
		memberID++
	}

	deckStartID := numArchSegments + 2
	for i := 0; i < numArchSegments; i++ {
		startID := deckStartID + i
		endID := deckStartID + i + 1
		member := structure.AddMember(memberID, startID, endID, e, aDeck, iDeck)
		if member != nil {
			memberTypes[memberID] = "deck_beam"
			structure.MemberTypes[memberID] = "deck_beam"
		}
		memberID++
	}

	for i := 0; i <= numArchSegments; i++ {
		archNodeID := i + 1
		deckNodeID := deckStartID + i
		member := structure.AddMember(memberID, archNodeID, deckNodeID, e, aVertical, iVertical)
		if member != nil {
			memberTypes[memberID] = "vertical_post"
			structure.MemberTypes[memberID] = "vertical_post"
		}
		memberID++
	}

	numDiagonals := numArchSegments
	for i := 0; i < numDiagonals; i++ {
		topID := deckStartID + i
		botID := i + 2
		member := structure.AddMember(memberID, topID, botID, e, aVertical*0.7, iVertical*0.5)
		if member != nil {
			memberTypes[memberID] = "diagonal_brace"
			structure.MemberTypes[memberID] = "diagonal_brace"
		}
		memberID++
	}

	structure.AssembleStiffness()

	return &BridgeModel{
		Structure:   structure,
		BridgeName:  "bridge_" + string(rune('0'+bridgeID)),
		SpanLength:  spanLength,
		ArchRise:    archRise,
		DeckWidth:   deckWidth,
		MemberTypes: memberTypes,
	}
}

type ParametricBridge struct {
	BridgeID     int
	BaseBridge   *BridgeModel
	MinSpan      float64
	MaxSpan      float64
	MinRise      float64
	MaxRise      float64
	MinWidth     float64
	MaxWidth     float64
	UseSurrogate bool
	Surrogate    *SurrogateModel
}

func NewParametricBridge(bridgeID int, defaultSpan, defaultRise, defaultWidth float64) *ParametricBridge {
	base := GenerateArchBridge(bridgeID, defaultSpan, defaultRise, defaultWidth)
	return &ParametricBridge{
		BridgeID:     bridgeID,
		BaseBridge:   base,
		MinSpan:      10.0,
		MaxSpan:      50.0,
		MinRise:      2.0,
		MaxRise:      15.0,
		MinWidth:     3.0,
		MaxWidth:     12.0,
		UseSurrogate: false,
	}
}

func (pb *ParametricBridge) SetConstraints(minSpan, maxSpan, minRise, maxRise, minWidth, maxWidth float64) {
	pb.MinSpan = minSpan
	pb.MaxSpan = maxSpan
	pb.MinRise = minRise
	pb.MaxRise = maxRise
	pb.MinWidth = minWidth
	pb.MaxWidth = maxWidth
}

func (pb *ParametricBridge) ValidateParams(span, rise, width float64) (valid bool, message string) {
	if span < pb.MinSpan || span > pb.MaxSpan {
		return false, "拱跨超出范围: " + formatFloat(pb.MinSpan) + "m - " + formatFloat(pb.MaxSpan) + "m"
	}
	if rise < pb.MinRise || rise > pb.MaxRise {
		return false, "矢高超出范围: " + formatFloat(pb.MinRise) + "m - " + formatFloat(pb.MaxRise) + "m"
	}
	if width < pb.MinWidth || width > pb.MaxWidth {
		return false, "桥面宽度超出范围: " + formatFloat(pb.MinWidth) + "m - " + formatFloat(pb.MaxWidth) + "m"
	}

	riseSpanRatio := rise / span
	if riseSpanRatio < 0.08 {
		return false, "矢跨比过小，建议大于1:12"
	}
	if riseSpanRatio > 0.35 {
		return false, "矢跨比过大，建议小于1:3"
	}

	return true, "参数合法"
}

func (pb *ParametricBridge) UpdateGeometry(span, rise, width float64) *BridgeModel {
	valid, _ := pb.ValidateParams(span, rise, width)
	if !valid {
		return pb.BaseBridge
	}

	newModel := GenerateArchBridge(pb.BridgeID, span, rise, width)

	if len(newModel.Structure.Members) > 0 {
		newModel.Structure.AssembleStiffness()
	}

	return newModel
}

type GeometryOption struct {
	Name    string  `json:"name"`
	Label   string  `json:"label"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Step    float64 `json:"step"`
	Default float64 `json:"default"`
}

func (pb *ParametricBridge) GetGeometryOptions() []GeometryOption {
	return []GeometryOption{
		{Name: "span_length", Label: "拱跨 (m)", Min: pb.MinSpan, Max: pb.MaxSpan, Step: 0.5, Default: pb.BaseBridge.SpanLength},
		{Name: "arch_rise", Label: "矢高 (m)", Min: pb.MinRise, Max: pb.MaxRise, Step: 0.1, Default: pb.BaseBridge.ArchRise},
		{Name: "deck_width", Label: "桥面宽度 (m)", Min: pb.MinWidth, Max: pb.MaxWidth, Step: 0.1, Default: pb.BaseBridge.DeckWidth},
	}
}

type SurrogateModel struct {
	CoeffsStress []float64
	CoeffsDisp   []float64
	CoeffsVolume []float64
	TrainPoints  []SurrogateTrainPoint
	IsTrained    bool
	SpanMin      float64
	SpanMax      float64
	RiseMin      float64
	RiseMax      float64
	WidthMin     float64
	WidthMax     float64
}

type SurrogateTrainPoint struct {
	Span          float64
	Rise          float64
	Width         float64
	MaxStressRatio float64
	MaxDisp        float64
	TotalVolume    float64
}

type SurrogateResult struct {
	MaxStressRatio   float64
	MaxDisplacement  float64
	TotalVolume      float64
	IsApproximation  bool
}

func NewSurrogateModel() *SurrogateModel {
	return &SurrogateModel{
		IsTrained: false,
	}
}

func (sm *SurrogateModel) Train(bridgeID int, loadValue, spanMin, spanMax, riseMin, riseMax, widthMin, widthMax float64) {
	sm.SpanMin = spanMin
	sm.SpanMax = spanMax
	sm.RiseMin = riseMin
	sm.RiseMax = riseMax
	sm.WidthMin = widthMin
	sm.WidthMax = widthMax

	spanLevels := []float64{spanMin, (spanMin + spanMax) / 2, spanMax}
	riseLevels := []float64{riseMin, (riseMin + riseMax) / 2, riseMax}
	widthLevels := []float64{widthMin, (widthMin + widthMax) / 2, widthMax}

	sm.TrainPoints = make([]SurrogateTrainPoint, 0)

	for _, s := range spanLevels {
		for _, r := range riseLevels {
			for _, w := range widthLevels {
				result := CalculateParametricAnalysis(bridgeID, s, r, w, loadValue)
				sm.TrainPoints = append(sm.TrainPoints, SurrogateTrainPoint{
					Span:           s,
					Rise:           r,
					Width:          w,
					MaxStressRatio: result.MaxStressRatio,
					MaxDisp:        result.MaxDisplacement,
					TotalVolume:    result.TotalVolume,
				})
			}
		}
	}

	sm.fitPolynomial()
	sm.IsTrained = true
}

func (sm *SurrogateModel) fitPolynomial() {
	n := len(sm.TrainPoints)
	if n < 10 {
		return
	}

	X := make([][]float64, n)
	yStress := make([]float64, n)
	yDisp := make([]float64, n)
	yVol := make([]float64, n)

	for i, p := range sm.TrainPoints {
		s := sm.normalize(p.Span, sm.SpanMin, sm.SpanMax)
		r := sm.normalize(p.Rise, sm.RiseMin, sm.RiseMax)
		w := sm.normalize(p.Width, sm.WidthMin, sm.WidthMax)

		X[i] = []float64{
			1.0,
			s, r, w,
			s * s, r * r, w * w,
			s * r, s * w, r * w,
		}
		yStress[i] = p.MaxStressRatio
		yDisp[i] = p.MaxDisp
		yVol[i] = p.TotalVolume
	}

	sm.CoeffsStress = leastSquares(X, yStress)
	sm.CoeffsDisp = leastSquares(X, yDisp)
	sm.CoeffsVolume = leastSquares(X, yVol)
}

func (sm *SurrogateModel) Predict(span, rise, width float64) *SurrogateResult {
	if !sm.IsTrained {
		return nil
	}

	s := sm.normalize(span, sm.SpanMin, sm.SpanMax)
	r := sm.normalize(rise, sm.RiseMin, sm.RiseMax)
	w := sm.normalize(width, sm.WidthMin, sm.WidthMax)

	x := []float64{
		1.0,
		s, r, w,
		s * s, r * r, w * w,
		s * r, s * w, r * w,
	}

	stress := polyEval(sm.CoeffsStress, x)
	disp := polyEval(sm.CoeffsDisp, x)
	vol := polyEval(sm.CoeffsVolume, x)

	if stress < 0 {
		stress = 0
	}
	if disp < 0 {
		disp = 0
	}
	if vol < 0 {
		vol = 0
	}

	return &SurrogateResult{
		MaxStressRatio:   stress,
		MaxDisplacement:  disp,
		TotalVolume:      vol,
		IsApproximation:  true,
	}
}

func (sm *SurrogateModel) normalize(val, min, max float64) float64 {
	if max <= min {
		return 0.5
	}
	return (val - min) / (max - min)
}

func polyEval(coeffs, x []float64) float64 {
	result := 0.0
	for i, c := range coeffs {
		if i < len(x) {
			result += c * x[i]
		}
	}
	return result
}

func leastSquares(X [][]float64, y []float64) []float64 {
	n := len(X[0])
	m := len(X)

	XTX := make([][]float64, n)
	for i := range XTX {
		XTX[i] = make([]float64, n)
	}
	XTy := make([]float64, n)

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			for k := 0; k < m; k++ {
				XTX[i][j] += X[k][i] * X[k][j]
			}
		}
		for k := 0; k < m; k++ {
			XTy[i] += X[k][i] * y[k]
		}
	}

	augmented := make([][]float64, n)
	for i := range augmented {
		augmented[i] = make([]float64, n+1)
		copy(augmented[i], XTX[i])
		augmented[i][n] = XTy[i]
	}

	for col := 0; col < n; col++ {
		maxRow := col
		for row := col + 1; row < n; row++ {
			if math.Abs(augmented[row][col]) > math.Abs(augmented[maxRow][col]) {
				maxRow = row
			}
		}
		augmented[col], augmented[maxRow] = augmented[maxRow], augmented[col]

		if math.Abs(augmented[col][col]) < 1e-12 {
			continue
		}

		for row := col + 1; row < n; row++ {
			factor := augmented[row][col] / augmented[col][col]
			for j := col; j <= n; j++ {
				augmented[row][j] -= factor * augmented[col][j]
			}
		}
	}

	coeffs := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		if math.Abs(augmented[i][i]) < 1e-12 {
			coeffs[i] = 0
			continue
		}
		coeffs[i] = augmented[i][n]
		for j := i + 1; j < n; j++ {
			coeffs[i] -= augmented[i][j] * coeffs[j]
		}
		coeffs[i] /= augmented[i][i]
	}

	return coeffs
}

func (pb *ParametricBridge) TrainSurrogate(loadValue float64) {
	if pb.Surrogate == nil {
		pb.Surrogate = NewSurrogateModel()
	}
	pb.Surrogate.Train(pb.BridgeID, loadValue, pb.MinSpan, pb.MaxSpan, pb.MinRise, pb.MaxRise, pb.MinWidth, pb.MaxWidth)
	pb.UseSurrogate = true
}

func (pb *ParametricBridge) AnalyzeWithSurrogate(span, rise, width float64) *SurrogateResult {
	if pb.Surrogate == nil || !pb.UseSurrogate {
		return nil
	}
	return pb.Surrogate.Predict(span, rise, width)
}

type ParametricAnalysisResult struct {
	BridgeID            int
	SpanLength          float64
	ArchRise            float64
	DeckWidth           float64
	RiseSpanRatio       float64
	MemberCount         int
	NodeCount           int
	MaxStressRatio      float64
	MaxDisplacement     float64
	MaxDisplacementNode int
	TotalVolume         float64
	MaterialEfficiency  float64
	AnalysisDuration_ms float64
	MemberForces        []MemberForces
	Displacements       []NodeDisplacement
	YingzaoComparison   []YingzaoComparison
	Nodes               []NodeInfo
	Members             []MemberInfo
	IsSurrogateResult   bool
}

type NodeInfo struct {
	ID        int     `json:"id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	UX        float64 `json:"ux"`
	UY        float64 `json:"uy"`
	IsSupport bool    `json:"is_support"`
}

type MemberInfo struct {
	ID          int     `json:"id"`
	StartID     int     `json:"start_id"`
	EndID       int     `json:"end_id"`
	Type        string  `json:"type"`
	Length      float64 `json:"length"`
	Area        float64 `json:"area"`
	Stress      float64 `json:"stress"`
	StressRatio float64 `json:"stress_ratio"`
}

func CalculateParametricAnalysis(bridgeID int, span, rise, width float64, loadValue float64) ParametricAnalysisResult {
	model := GenerateArchBridge(bridgeID, span, rise, width)
	model.Structure.AssembleStiffness()

	loads := make(map[int]float64)
	numDeckNodes := (len(model.Structure.Nodes) / 2)
	startDeckID := numDeckNodes + 1
	for i := 0; i < numDeckNodes; i++ {
		nodeID := startDeckID + i
		loads[nodeID] = -loadValue / float64(numDeckNodes)
	}

	model.Structure.ApplyLoads(loads)
	model.Structure.Solve()

	forces := model.Structure.CalculateMemberForces()
	displacements := model.Structure.GetNodeDisplacements()
	comparison := model.Structure.CompareWithYingzaoFashi()

	maxStressRatio := 0.0
	maxDisp := 0.0
	maxDispNode := 0
	totalVolume := 0.0

	for _, f := range forces {
		if math.Abs(f.StressRatio) > maxStressRatio {
			maxStressRatio = math.Abs(f.StressRatio)
		}
		totalVolume += f.Length * model.Structure.Members[f.MemberID-1].A
	}

	for _, d := range displacements {
		disp := math.Sqrt(d.UX*d.UX + d.UY*d.UY)
		if disp > maxDisp {
			maxDisp = disp
			maxDispNode = d.NodeID
		}
	}

	riseSpanRatio := rise / span
	materialEfficiency := maxStressRatio / (totalVolume / (span * width))
	if materialEfficiency == 0 {
		materialEfficiency = 0
	}

	nodeInfos := make([]NodeInfo, 0, len(model.Structure.Nodes))
	for _, node := range model.Structure.Nodes {
		nodeInfos = append(nodeInfos, NodeInfo{
			ID:        node.ID,
			X:         node.X,
			Y:         node.Y,
			UX:        node.UX,
			UY:        node.UY,
			IsSupport: node.FixedX && node.FixedY,
		})
	}

	memberInfos := make([]MemberInfo, 0, len(forces))
	for _, f := range forces {
		member := model.Structure.Members[f.MemberID-1]
		memberInfos = append(memberInfos, MemberInfo{
			ID:          member.ID,
			StartID:     member.Start.ID,
			EndID:       member.End.ID,
			Type:        model.MemberTypes[member.ID],
			Length:      member.Length,
			Area:        member.A,
			Stress:      f.NormalStress + f.BendingStress,
			StressRatio: f.StressRatio,
		})
	}

	return ParametricAnalysisResult{
		BridgeID:            bridgeID,
		SpanLength:          span,
		ArchRise:            rise,
		DeckWidth:           width,
		RiseSpanRatio:       riseSpanRatio,
		MemberCount:         len(model.Structure.Members),
		NodeCount:           len(model.Structure.Nodes),
		MaxStressRatio:      maxStressRatio,
		MaxDisplacement:     maxDisp,
		MaxDisplacementNode: maxDispNode,
		TotalVolume:         totalVolume,
		MaterialEfficiency:  materialEfficiency,
		MemberForces:        forces,
		Displacements:       displacements,
		YingzaoComparison:   comparison,
		Nodes:               nodeInfos,
		Members:             memberInfos,
	}
}

func formatFloat(v float64) string {
	if v < 0 {
		v = -v
	}
	return fmt.Sprintf("%.1f", v)
}
