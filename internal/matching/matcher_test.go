package matching

import (
	"math"
	"testing"

	"github.com/sumeetmehra/procura/internal/model"
)

func TestCPVOverlap(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected float64
	}{
		{
			name:     "identical sets",
			a:        []string{"45000000", "72000000"},
			b:        []string{"45000000", "72000000"},
			expected: 1.0,
		},
		{
			name:     "partial overlap",
			a:        []string{"45000000", "72000000"},
			b:        []string{"45000000", "79000000"},
			expected: 1.0 / 3.0,
		},
		{
			name:     "no overlap",
			a:        []string{"45000000"},
			b:        []string{"72000000"},
			expected: 0.0,
		},
		{
			name:     "empty first set",
			a:        []string{},
			b:        []string{"45000000"},
			expected: 0.0,
		},
		{
			name:     "both empty",
			a:        []string{},
			b:        []string{},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CPVOverlap(tt.a, tt.b)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestValueFit(t *testing.T) {
	minVal := 100000.0
	maxVal := 500000.0
	company := model.Company{
		MinValue: &minVal,
		MaxValue: &maxVal,
	}

	tests := []struct {
		name     string
		value    *float64
		expected float64
	}{
		{
			name:     "within range",
			value:    floatPtr(250000),
			expected: 1.0,
		},
		{
			name:     "at minimum",
			value:    floatPtr(100000),
			expected: 1.0,
		},
		{
			name:     "at maximum",
			value:    floatPtr(500000),
			expected: 1.0,
		},
		{
			name:     "nil value",
			value:    nil,
			expected: 0.5,
		},
		{
			name:  "below range decays",
			value: floatPtr(50000),
		},
		{
			name:  "above range decays",
			value: floatPtr(1000000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValueFit(tt.value, company)
			if tt.expected != 0 {
				if math.Abs(result-tt.expected) > 0.001 {
					t.Errorf("expected %f, got %f", tt.expected, result)
				}
			} else {
				// For decay tests, just verify it's between 0 and 1 exclusive
				if result <= 0 || result >= 1.0 {
					t.Errorf("expected value between 0 and 1, got %f", result)
				}
			}
		})
	}
}

func TestComputeScore(t *testing.T) {
	// Perfect match
	score := ComputeScore(1.0, 1.0, 1.0, 1.0)
	if math.Abs(score-1.0) > 0.001 {
		t.Errorf("expected 1.0, got %f", score)
	}

	// Zero match
	score = ComputeScore(0.0, 0.0, 0.0, 0.0)
	if math.Abs(score) > 0.001 {
		t.Errorf("expected 0.0, got %f", score)
	}

	// Verify weights: 0.4 + 0.3 + 0.2 + 0.1 = 1.0
	score = ComputeScore(0.5, 0.5, 0.5, 0.5)
	if math.Abs(score-0.5) > 0.001 {
		t.Errorf("expected 0.5, got %f", score)
	}
}

func floatPtr(v float64) *float64 {
	return &v
}
