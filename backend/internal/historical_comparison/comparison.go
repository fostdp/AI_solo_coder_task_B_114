package historicalcomparison

import (
	bridgecomparator "ancient-bridge-system/internal/bridge_comparator"
)

type Dynasty = bridgecomparator.Dynasty

const (
	DynastyHanJin = bridgecomparator.DynastyHanJin
	DynastyTang   = bridgecomparator.DynastyTang
	DynastySong   = bridgecomparator.DynastySong
	DynastyMing   = bridgecomparator.DynastyMing
	DynastyQing   = bridgecomparator.DynastyQing
)

type BridgeTypology = bridgecomparator.BridgeTypology

const (
	TypologyBeamBridge    = bridgecomparator.TypologyBeamBridge
	TypologyTimberArch    = bridgecomparator.TypologyTimberArch
	TypologyThroughArch   = bridgecomparator.TypologyThroughArch
	TypologyGalleryBridge = bridgecomparator.TypologyGalleryBridge
)

type HistoricalBridge = bridgecomparator.HistoricalBridge

type EfficiencyMetrics = bridgecomparator.EfficiencyMetrics

type ComparisonResult = bridgecomparator.ComparisonResult

type RadarPoint = bridgecomparator.RadarPoint

type TechEvolutionPoint = bridgecomparator.TechEvolutionPoint

type HistoricalComparison = bridgecomparator.HistoricalComparison

func NewHistoricalComparison() *HistoricalComparison {
	return bridgecomparator.NewHistoricalComparison()
}

var HistoricalBridgeDatabase = bridgecomparator.HistoricalBridgeDatabase
