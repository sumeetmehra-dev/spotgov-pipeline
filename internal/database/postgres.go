package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	maxRetries    = 3
	retryBaseWait = 2 * time.Second
)

func Connect(dsn string, logger *zap.Logger) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	gormCfg := &gorm.Config{
		Logger: newZapGormLogger(logger),
	}

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), gormCfg)
		if err == nil {
			sqlDB, pingErr := db.DB()
			if pingErr != nil {
				err = pingErr
			} else if pingErr = sqlDB.Ping(); pingErr != nil {
				err = pingErr
			} else {
				sqlDB.SetMaxOpenConns(25)
				sqlDB.SetMaxIdleConns(10)
				sqlDB.SetConnMaxLifetime(5 * time.Minute)
				logger.Info("connected to PostgreSQL")
				return db, nil
			}
		}

		wait := retryBaseWait * time.Duration(1<<uint(i))
		logger.Warn("failed to connect to PostgreSQL, retrying",
			zap.Int("attempt", i+1),
			zap.Duration("backoff", wait),
			zap.Error(err),
		)
		time.Sleep(wait)
	}

	return nil, fmt.Errorf("failed to connect to PostgreSQL after %d attempts: %w", maxRetries, err)
}

func EnablePgvector(db *gorm.DB) error {
	return db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error
}

// zapGormLogger adapts zap.Logger to GORM's logger interface.
type zapGormLogger struct {
	logger *zap.Logger
}

func newZapGormLogger(logger *zap.Logger) gormlogger.Interface {
	return &zapGormLogger{logger: logger.Named("gorm")}
}

func (l *zapGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *zapGormLogger) Info(_ context.Context, msg string, data ...interface{}) {
	l.logger.Sugar().Infof(msg, data...)
}

func (l *zapGormLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	l.logger.Sugar().Warnf(msg, data...)
}

func (l *zapGormLogger) Error(_ context.Context, msg string, data ...interface{}) {
	l.logger.Sugar().Errorf(msg, data...)
}

func (l *zapGormLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		l.logger.Error("query error",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
			zap.Error(err),
		)
	case elapsed > time.Second:
		l.logger.Warn("slow query",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
		)
	}
}
