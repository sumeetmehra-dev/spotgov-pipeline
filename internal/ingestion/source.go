package ingestion

import (
	"context"
	"time"

	"github.com/sumeetmehra/procura/internal/model"
)

// Source defines the contract for any procurement data source.
// Implementing this interface allows new sources to be added
// without modifying the orchestrator.
type Source interface {
	Name() string
	Fetch(ctx context.Context, since time.Time) ([]model.Tender, error)
}
