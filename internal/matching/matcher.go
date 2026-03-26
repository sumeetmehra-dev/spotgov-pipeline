package matching

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sumeetmehra/spotgov-pipeline/internal/embedding"
	"github.com/sumeetmehra/spotgov-pipeline/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Matcher orchestrates company-tender matching using vector similarity and scoring.
type Matcher struct {
	db      *gorm.DB
	embedCl *embedding.Client
	store   *embedding.Store
	logger  *zap.Logger
}

func NewMatcher(db *gorm.DB, embedCl *embedding.Client, store *embedding.Store, logger *zap.Logger) *Matcher {
	return &Matcher{
		db:      db,
		embedCl: embedCl,
		store:   store,
		logger:  logger.Named("matcher"),
	}
}

// MatchCompany finds and scores tenders for the given company.
func (m *Matcher) MatchCompany(ctx context.Context, companyID uuid.UUID) ([]model.Match, error) {
	var company model.Company
	if err := m.db.First(&company, "id = ?", companyID).Error; err != nil {
		return nil, fmt.Errorf("find company: %w", err)
	}

	// Generate company embedding from description + keywords
	companyText := buildCompanyText(company)
	companyVec, err := m.embedCl.EmbedText(ctx, companyText)
	if err != nil {
		return nil, fmt.Errorf("embed company profile: %w", err)
	}

	// Store the company embedding
	if err := m.store.StoreCompanyEmbedding(ctx, companyID, companyVec); err != nil {
		m.logger.Warn("failed to store company embedding", zap.Error(err))
	}

	// Find similar tenders via pgvector
	tenders, vectorScores, err := m.store.SimilarTenders(ctx, companyVec, 100)
	if err != nil {
		return nil, fmt.Errorf("similarity search: %w", err)
	}

	m.logger.Info("found candidate tenders",
		zap.String("company", company.Name),
		zap.Int("candidates", len(tenders)),
	)

	// Score each tender
	var matches []model.Match
	for i, tender := range tenders {
		cpvScore := CPVOverlap(company.CPVCodes, tender.CPVCodes)
		valScore := ValueFit(tender.EstimatedValue, company)

		// BM25 score placeholder — will be filled by hybrid search in Phase 5
		bm25Score := 0.0

		score := ComputeScore(vectorScores[i], bm25Score, cpvScore, valScore)

		matches = append(matches, model.Match{
			CompanyID:   companyID,
			TenderID:    tender.ID,
			Score:       score,
			VectorScore: vectorScores[i],
			BM25Score:   bm25Score,
			CPVOverlap:  cpvScore,
		})
	}

	// Upsert matches
	if len(matches) > 0 {
		if err := m.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "company_id"}, {Name: "tender_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"score", "vector_score", "bm25_score", "cpv_overlap", "matched_at",
			}),
		}).Create(&matches).Error; err != nil {
			return nil, fmt.Errorf("upsert matches: %w", err)
		}
	}

	m.logger.Info("matching complete",
		zap.String("company", company.Name),
		zap.Int("matches", len(matches)),
	)

	return matches, nil
}

// GenerateTenderEmbeddings creates embeddings for tenders that don't have them yet.
func (m *Matcher) GenerateTenderEmbeddings(ctx context.Context, batchSize int) (int, error) {
	tenders, err := m.store.TendersWithoutEmbeddings(ctx, batchSize)
	if err != nil {
		return 0, fmt.Errorf("fetch tenders: %w", err)
	}

	if len(tenders) == 0 {
		return 0, nil
	}

	texts := make([]string, len(tenders))
	for i, t := range tenders {
		texts[i] = buildTenderText(t)
	}

	vectors, err := m.embedCl.EmbedTexts(ctx, texts)
	if err != nil {
		return 0, fmt.Errorf("generate embeddings: %w", err)
	}

	for i, vec := range vectors {
		if err := m.store.StoreTenderEmbedding(ctx, tenders[i].ID, vec); err != nil {
			m.logger.Error("failed to store embedding",
				zap.String("tender_id", tenders[i].ID.String()),
				zap.Error(err),
			)
		}
	}

	m.logger.Info("generated tender embeddings", zap.Int("count", len(vectors)))
	return len(vectors), nil
}

func buildCompanyText(c model.Company) string {
	parts := []string{c.Name}
	if c.Description != "" {
		parts = append(parts, c.Description)
	}
	if len(c.Keywords) > 0 {
		parts = append(parts, "Keywords: "+strings.Join(c.Keywords, ", "))
	}
	if len(c.CPVCodes) > 0 {
		parts = append(parts, "CPV: "+strings.Join(c.CPVCodes, ", "))
	}
	return strings.Join(parts, ". ")
}

func buildTenderText(t model.Tender) string {
	parts := []string{t.Title}
	if t.Description != "" {
		parts = append(parts, t.Description)
	}
	if t.BuyerName != "" {
		parts = append(parts, "Buyer: "+t.BuyerName)
	}
	if len(t.CPVCodes) > 0 {
		parts = append(parts, "CPV: "+strings.Join(t.CPVCodes, ", "))
	}
	return strings.Join(parts, ". ")
}
