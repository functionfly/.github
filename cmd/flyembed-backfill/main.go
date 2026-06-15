package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/recommendations"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func main() {
	batchSize := flag.Int("batch", 10, "Number of functions to process per batch")
	dryRun := flag.Bool("dry-run", false, "Print what would be done without making changes")
	limit := flag.Int("limit", 0, "Maximum number of functions to backfill (0 = all)")
	flag.Parse()

	logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
	logrus.SetLevel(logrus.InfoLevel)

	ctx := context.Background()

	db, err := storage.NewPostgresDB()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	repo := recommendations.NewRepository(db.GORM)
	registryRepo := registry.NewRegistryRepository(db.GORM, nil)
	svc := recommendations.NewService(db.GORM, registryRepo, nil)

	existingCount, err := repo.CountTripleEmbeddings(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to count existing triple embeddings")
	}
	logrus.WithField("existing_triple_embeddings", existingCount).Info("Starting FlyEmbed backfill")

	pageSize := *batchSize
	if *limit > 0 && *limit < pageSize {
		pageSize = *limit
	}

	totalBackfilled := 0
	totalErrors := 0
	processed := 0

	for {
		ids, err := repo.ListFunctionsWithoutTriples(ctx, pageSize)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to list functions without triples")
		}

		if len(ids) == 0 {
			break
		}

		for _, fnID := range ids {
			if *limit > 0 && totalBackfilled >= *limit {
				goto done
			}

			if err := backfillFunction(ctx, svc, registryRepo, fnID, *dryRun); err != nil {
				totalErrors++
				logrus.WithError(err).WithField("function_id", fnID).Error("Failed to backfill function")
			} else {
				totalBackfilled++
			}

			processed++
			if processed%10 == 0 {
				logrus.WithFields(logrus.Fields{
					"backfilled": totalBackfilled,
					"errors":     totalErrors,
					"processed":  processed,
				}).Info("Progress")
			}
		}

		if len(ids) < pageSize {
			break
		}
	}

done:
	finalCount, _ := repo.CountTripleEmbeddings(ctx)
	logrus.WithFields(logrus.Fields{
		"total_backfilled":  totalBackfilled,
		"total_errors":      totalErrors,
		"triple_embeddings": finalCount,
	}).Info("FlyEmbed backfill complete")

	if totalErrors > 0 {
		os.Exit(1)
	}
}

func backfillFunction(ctx context.Context, svc *recommendations.Service, registryRepo *registry.RegistryRepository, fnID uuid.UUID, dryRun bool) error {
	fn, err := registryRepo.GetFunctionByID(ctx, fnID)
	if err != nil {
		return fmt.Errorf("get function: %w", err)
	}

	fnVersion, err := registryRepo.GetLatestFunctionVersion(fnID)
	if err != nil {
		return fmt.Errorf("get latest version: %w", err)
	}

	var manifest map[string]interface{}
	if fnVersion.Manifest != nil {
		if err := json.Unmarshal(fnVersion.Manifest, &manifest); err != nil {
			logrus.WithError(err).Warn("Failed to parse manifest")
			manifest = map[string]interface{}{}
		}
	}
	if manifest == nil {
		manifest = map[string]interface{}{}
	}

	var tags []string
	if fn.Tags != nil {
		_ = json.Unmarshal(fn.Tags, &tags)
	}

	var capabilities []string
	if fnVersion.Capabilities != nil {
		_ = json.Unmarshal(fnVersion.Capabilities, &capabilities)
	}

	sourceCode := ""
	if fnVersion.SourceCode.Valid {
		sourceCode = fnVersion.SourceCode.String
	}

	runtime := fnVersion.Runtime

	title := ""
	if fn.Title.Valid {
		title = fn.Title.String
	}
	description := ""
	if fn.Description.Valid {
		description = fn.Description.String
	}
	category := ""
	if fn.Category.Valid {
		category = fn.Category.String
	}

	if dryRun {
		logrus.WithFields(logrus.Fields{
			"function_id": fnID,
			"author":      fn.Author,
			"name":        fn.Name,
			"runtime":     runtime,
		}).Info("Would backfill (dry-run)")
		return nil
	}

	return svc.EmbedFunctionViaAIService(ctx, fnID, fn.Name, title, description, category, tags, manifest, sourceCode, runtime, capabilities)
}
