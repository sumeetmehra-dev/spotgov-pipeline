package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/sumeetmehra/spotgov-pipeline/internal/config"
	"github.com/sumeetmehra/spotgov-pipeline/internal/database"
	"github.com/sumeetmehra/spotgov-pipeline/internal/embedding"
	"github.com/sumeetmehra/spotgov-pipeline/internal/ingestion"
	"github.com/sumeetmehra/spotgov-pipeline/internal/ingestion/dados"
	"github.com/sumeetmehra/spotgov-pipeline/internal/ingestion/ted"
	"github.com/sumeetmehra/spotgov-pipeline/internal/matching"
	"github.com/sumeetmehra/spotgov-pipeline/internal/search"
	"github.com/sumeetmehra/spotgov-pipeline/internal/server"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	// Initialize logger
	var logger *zap.Logger
	var err error
	if cfg.IsDev() {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("starting spotgov-pipeline",
		zap.String("env", cfg.Env),
		zap.Int("port", cfg.Port),
	)

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("database connection failed", zap.Error(err))
	}

	// Run migrations
	if err := database.AutoMigrate(db, logger); err != nil {
		logger.Fatal("migration failed", zap.Error(err))
	}

	// Initialize Elasticsearch
	var esClient *search.ESClient
	var indexer *search.Indexer
	esClient, err = search.NewESClient(cfg.ElasticsearchURL, logger)
	if err != nil {
		logger.Warn("Elasticsearch unavailable, search will use DB fallback", zap.Error(err))
	} else {
		if err := esClient.CreateIndex(context.Background()); err != nil {
			logger.Error("failed to create ES index", zap.Error(err))
		}
		indexer = search.NewIndexer(esClient, logger)
	}

	// Initialize data sources
	tedClient := ted.NewClient(cfg.TEDBaseURL, logger)
	dadosClient := dados.NewClient(cfg.DadosBaseURL, logger)
	orchestrator := ingestion.NewOrchestrator(db, logger, indexer, tedClient, dadosClient)

	// Initialize embedding client (optional — degrades gracefully if no API key)
	var matcher *matching.Matcher
	embedCl, err := embedding.NewClient(cfg.MistralKey, logger)
	if err != nil {
		logger.Warn("embedding client unavailable, matching features disabled", zap.Error(err))
		matcher = matching.NewMatcher(db, nil, embedding.NewStore(db, logger), logger)
	} else {
		store := embedding.NewStore(db, logger)
		matcher = matching.NewMatcher(db, embedCl, store, logger)
	}

	// Build router
	r := server.NewRouter(db, orchestrator, matcher, esClient, logger)

	// Start HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("HTTP server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}

	logger.Info("server stopped")
}
