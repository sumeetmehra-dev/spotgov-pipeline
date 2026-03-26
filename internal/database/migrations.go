package database

import (
	"github.com/sumeetmehra/spotgov-pipeline/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("running database migrations")

	if err := EnablePgvector(db); err != nil {
		logger.Warn("pgvector extension not available, vector features will be disabled", zap.Error(err))
	}

	if err := db.AutoMigrate(
		&model.Tender{},
		&model.Company{},
		&model.Match{},
	); err != nil {
		return err
	}

	// Add vector columns for pgvector (GORM doesn't natively support the vector type).
	vectorMigrations := []string{
		"ALTER TABLE tenders ADD COLUMN IF NOT EXISTS embedding vector(1024)",
		"ALTER TABLE companies ADD COLUMN IF NOT EXISTS embedding vector(1024)",
	}
	for _, sql := range vectorMigrations {
		if err := db.Exec(sql).Error; err != nil {
			logger.Warn("vector column migration skipped", zap.String("sql", sql), zap.Error(err))
		}
	}

	return nil
}
