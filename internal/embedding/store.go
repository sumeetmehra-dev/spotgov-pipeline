package embedding

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sumeetmehra/spotgov-pipeline/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Store handles pgvector operations for tender embeddings.
type Store struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewStore(db *gorm.DB, logger *zap.Logger) *Store {
	return &Store{
		db:     db,
		logger: logger.Named("embedding.store"),
	}
}

// StoreTenderEmbedding saves an embedding vector for a tender.
func (s *Store) StoreTenderEmbedding(ctx context.Context, tenderID uuid.UUID, vec Vector) error {
	vecStr := vectorToString(vec)
	return s.db.WithContext(ctx).
		Exec("UPDATE tenders SET embedding = ? WHERE id = ?", vecStr, tenderID).
		Error
}

// StoreCompanyEmbedding saves an embedding vector for a company.
func (s *Store) StoreCompanyEmbedding(ctx context.Context, companyID uuid.UUID, vec Vector) error {
	vecStr := vectorToString(vec)
	return s.db.WithContext(ctx).
		Exec("UPDATE companies SET embedding = ? WHERE id = ?", vecStr, companyID).
		Error
}

// SimilarTenders finds tenders most similar to the given vector using cosine distance.
func (s *Store) SimilarTenders(ctx context.Context, vec Vector, limit int) ([]model.Tender, []float64, error) {
	vecStr := vectorToString(vec)

	var results []struct {
		model.Tender
		Similarity float64
	}

	err := s.db.WithContext(ctx).
		Raw(`SELECT *, 1 - (embedding <=> ?) as similarity
			 FROM tenders
			 WHERE embedding IS NOT NULL
			 ORDER BY embedding <=> ?
			 LIMIT ?`, vecStr, vecStr, limit).
		Scan(&results).Error

	if err != nil {
		return nil, nil, fmt.Errorf("similarity search: %w", err)
	}

	tenders := make([]model.Tender, len(results))
	scores := make([]float64, len(results))
	for i, r := range results {
		tenders[i] = r.Tender
		scores[i] = r.Similarity
	}

	return tenders, scores, nil
}

// TendersWithoutEmbeddings returns tenders that don't have embeddings yet.
func (s *Store) TendersWithoutEmbeddings(ctx context.Context, limit int) ([]model.Tender, error) {
	var tenders []model.Tender
	err := s.db.WithContext(ctx).
		Where("embedding IS NULL").
		Limit(limit).
		Find(&tenders).Error
	return tenders, err
}

func vectorToString(vec Vector) string {
	parts := make([]string, len(vec))
	for i, v := range vec {
		parts[i] = fmt.Sprintf("%f", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
