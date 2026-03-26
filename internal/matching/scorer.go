package matching

import (
	"math"

	"github.com/sumeetmehra/procura/internal/model"
)

const (
	weightVector  = 0.4
	weightBM25    = 0.3
	weightCPV     = 0.2
	weightValue   = 0.1
)

// ComputeScore calculates the composite match score.
func ComputeScore(vectorSim, bm25Score, cpvOverlap, valueFit float64) float64 {
	score := weightVector*vectorSim +
		weightBM25*bm25Score +
		weightCPV*cpvOverlap +
		weightValue*valueFit

	return math.Min(1.0, math.Max(0.0, score))
}

// CPVOverlap computes the Jaccard index between two sets of CPV codes.
func CPVOverlap(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	setA := make(map[string]bool, len(a))
	for _, code := range a {
		setA[code] = true
	}

	intersection := 0
	setB := make(map[string]bool, len(b))
	for _, code := range b {
		setB[code] = true
		if setA[code] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

// ValueFit returns 1.0 if the tender value falls within the company's range,
// and decays exponentially as it moves outside.
func ValueFit(tenderValue *float64, company model.Company) float64 {
	if tenderValue == nil {
		return 0.5 // Neutral score when value is unknown
	}

	val := *tenderValue

	minVal := 0.0
	maxVal := math.MaxFloat64
	if company.MinValue != nil {
		minVal = *company.MinValue
	}
	if company.MaxValue != nil {
		maxVal = *company.MaxValue
	}

	if val >= minVal && val <= maxVal {
		return 1.0
	}

	// Exponential decay outside range
	var distance float64
	if val < minVal {
		distance = (minVal - val) / math.Max(minVal, 1)
	} else {
		distance = (val - maxVal) / math.Max(maxVal, 1)
	}

	return math.Exp(-distance)
}
