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

	return db.AutoMigrate(
		&model.Tender{},
		&model.Company{},
		&model.Match{},
	)
}
