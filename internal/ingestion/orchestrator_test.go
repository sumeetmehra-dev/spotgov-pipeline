package ingestion

import (
	"context"
	"testing"
	"time"

	"github.com/sumeetmehra/spotgov-pipeline/internal/model"
)

// mockSource implements Source for testing.
type mockSource struct {
	name    string
	tenders []model.Tender
	err     error
}

func (m *mockSource) Name() string { return m.name }
func (m *mockSource) Fetch(_ context.Context, _ time.Time) ([]model.Tender, error) {
	return m.tenders, m.err
}

func TestOrchestrator_SourceInterface(t *testing.T) {
	src := &mockSource{
		name: "test-source",
		tenders: []model.Tender{
			{Source: "test", SourceID: "test-001", Title: "Test Tender 1"},
			{Source: "test", SourceID: "test-002", Title: "Test Tender 2"},
		},
	}

	if src.Name() != "test-source" {
		t.Errorf("expected name 'test-source', got %q", src.Name())
	}

	tenders, err := src.Fetch(context.Background(), time.Now())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(tenders) != 2 {
		t.Errorf("expected 2 tenders, got %d", len(tenders))
	}
}
