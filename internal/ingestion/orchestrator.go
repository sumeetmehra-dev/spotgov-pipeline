package ingestion

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sumeetmehra/spotgov-pipeline/internal/model"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Orchestrator manages concurrent ingestion from multiple procurement sources.
type Orchestrator struct {
	sources []Source
	db      *gorm.DB
	logger  *zap.Logger
}

func NewOrchestrator(db *gorm.DB, logger *zap.Logger, sources ...Source) *Orchestrator {
	return &Orchestrator{
		sources: sources,
		db:      db,
		logger:  logger.Named("orchestrator"),
	}
}

// Run fetches tenders from all sources concurrently and upserts them to the database.
func (o *Orchestrator) Run(ctx context.Context) (int, error) {
	since := time.Now().AddDate(0, -3, 0) // Last 3 months

	var mu sync.Mutex
	var allTenders []model.Tender

	g, gctx := errgroup.WithContext(ctx)

	for _, src := range o.sources {
		src := src
		g.Go(func() error {
			o.logger.Info("starting ingestion", zap.String("source", src.Name()))

			tenders, err := src.Fetch(gctx, since)
			if err != nil {
				o.logger.Error("source fetch failed",
					zap.String("source", src.Name()),
					zap.Error(err),
				)
				// Don't fail the entire pipeline for one source
				return nil
			}

			o.logger.Info("source fetch complete",
				zap.String("source", src.Name()),
				zap.Int("count", len(tenders)),
			)

			mu.Lock()
			allTenders = append(allTenders, tenders...)
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return 0, fmt.Errorf("ingestion failed: %w", err)
	}

	if len(allTenders) == 0 {
		o.logger.Info("no new tenders found")
		return 0, nil
	}

	// Bulk upsert: insert or update on source_id conflict
	inserted, err := o.upsertTenders(allTenders)
	if err != nil {
		return 0, fmt.Errorf("upsert tenders: %w", err)
	}

	o.logger.Info("ingestion complete",
		zap.Int("fetched", len(allTenders)),
		zap.Int("upserted", inserted),
	)

	return inserted, nil
}

func (o *Orchestrator) upsertTenders(tenders []model.Tender) (int, error) {
	batchSize := 100
	total := 0

	for i := 0; i < len(tenders); i += batchSize {
		end := i + batchSize
		if end > len(tenders) {
			end = len(tenders)
		}
		batch := tenders[i:end]

		result := o.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "source_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"title", "description", "cpv_codes", "buyer_name", "buyer_country",
				"place_of_performance", "procurement_method", "estimated_value",
				"currency", "deadline", "publication_date", "notice_url", "raw_data", "updated_at",
			}),
		}).Create(&batch)

		if result.Error != nil {
			return total, result.Error
		}
		total += int(result.RowsAffected)
	}

	return total, nil
}
