package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	backend "github.com/samcm/pyre"
	"github.com/samcm/pyre/internal/api"
	"github.com/samcm/pyre/internal/backfill"
	"github.com/samcm/pyre/internal/config"
	"github.com/samcm/pyre/internal/polymarket"
	"github.com/samcm/pyre/internal/server"
	"github.com/samcm/pyre/internal/storage"
	"github.com/sirupsen/logrus"
)

var (
	configPath = flag.String("config", "config.yaml", "path to config file")
	logLevel   = flag.String("log-level", "info", "log level (debug, info, warn, error)")
)

func main() {
	flag.Parse()

	// Setup logger
	log := setupLogger(*logLevel)
	log.Info("starting pyre")

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.WithError(err).Fatal("failed to load config")
	}
	log.WithField("config_path", *configPath).Info("configuration loaded")

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize storage
	log.Info("initializing storage")
	store := storage.NewStorage(cfg.Database.Path, log)
	if err := store.Start(ctx); err != nil {
		log.WithError(err).Fatal("failed to start storage")
	}
	defer func() {
		if err := store.Stop(); err != nil {
			log.WithError(err).Error("failed to stop storage")
		}
	}()

	// Initialize Polymarket client
	log.Info("initializing polymarket client")
	pmClient := polymarket.NewClient(log)

	// Ensure personas exist in database
	log.Info("ensuring personas exist")
	if err := ensurePersonas(ctx, store, cfg, log); err != nil {
		log.WithError(err).Fatal("failed to ensure personas")
	}

	// Initialize sync service with all users (from both legacy and personas)
	log.Info("initializing sync service")
	syncService := polymarket.NewService(pmClient, store, cfg.GetAllUsers(), cfg.Sync.IntervalMinutes, log)
	if err := syncService.Start(ctx); err != nil {
		log.WithError(err).Fatal("failed to start sync service")
	}
	defer func() {
		if err := syncService.Stop(); err != nil {
			log.WithError(err).Error("failed to stop sync service")
		}
	}()

	// Initialize backfill service
	log.Info("initializing backfill service")
	backfillService := backfill.NewService(store, log)

	// Initialize API handler
	log.Info("initializing API handler")
	handler := api.NewHandler(store, syncService, backfillService, log)

	// Get frontend embed
	frontendFS := backend.FrontendFiles

	// Initialize HTTP server
	log.Info("initializing HTTP server")
	httpServer := server.NewServer(cfg.Server.Host, cfg.Server.Port, handler, frontendFS, log)
	if err := httpServer.Start(ctx); err != nil {
		log.WithError(err).Fatal("failed to start HTTP server")
	}
	defer func() {
		if err := httpServer.Stop(); err != nil {
			log.WithError(err).Error("failed to stop HTTP server")
		}
	}()

	log.WithFields(logrus.Fields{
		"host": cfg.Server.Host,
		"port": cfg.Server.Port,
	}).Info("pyre started successfully")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Info("shutting down gracefully")
	cancel()
}

// setupLogger configures and returns a logger
func setupLogger(level string) *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		log.WithError(err).Warn("invalid log level, defaulting to info")
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	return log
}

// ensurePersonas creates personas and their users in the database from config
func ensurePersonas(ctx context.Context, store storage.Storage, cfg *config.Config, log *logrus.Logger) error {
	// Process personas from config
	for slug, personaCfg := range cfg.Personas {
		log.WithFields(logrus.Fields{
			"slug":         slug,
			"display_name": personaCfg.DisplayName,
			"usernames":    len(personaCfg.Usernames),
		}).Info("ensuring persona exists")

		// Check if persona exists, create if not
		persona, err := store.GetPersona(ctx, slug)
		if err != nil {
			// Persona doesn't exist, create it
			if personaCfg.Image != "" {
				persona, err = store.CreatePersonaWithImage(ctx, slug, personaCfg.DisplayName, personaCfg.Image)
			} else {
				persona, err = store.CreatePersona(ctx, slug, personaCfg.DisplayName)
			}
			if err != nil {
				return fmt.Errorf("failed to create persona %s: %w", slug, err)
			}
			log.WithField("slug", slug).Info("created persona")
		} else if personaCfg.Image != "" {
			// Update persona image if it changed
			if err := store.UpdatePersonaImage(ctx, persona.ID, personaCfg.Image); err != nil {
				log.WithError(err).WithField("slug", slug).Warn("failed to update persona image")
			}
		}

		// Process usernames for this persona
		for username, addresses := range personaCfg.Usernames {
			log.WithFields(logrus.Fields{
				"persona":   slug,
				"username":  username,
				"addresses": len(addresses),
			}).Debug("ensuring user exists")

			// Check if user exists
			user, err := store.GetUser(ctx, username)
			if err != nil {
				// User doesn't exist, create with persona
				_, err = store.CreateUserWithPersona(ctx, username, addresses, persona.ID)
				if err != nil {
					return fmt.Errorf("failed to create user %s for persona %s: %w", username, slug, err)
				}
				log.WithFields(logrus.Fields{
					"username": username,
					"persona":  slug,
				}).Info("created user with persona")
			} else {
				// User exists, update persona association
				if err := store.UpdateUserPersona(ctx, user.ID, persona.ID); err != nil {
					return fmt.Errorf("failed to update persona for user %s: %w", username, err)
				}
				log.WithFields(logrus.Fields{
					"username": username,
					"persona":  slug,
				}).Debug("linked existing user to persona")
			}
		}
	}

	// Process legacy users (users without personas)
	for username, addresses := range cfg.Users {
		log.WithFields(logrus.Fields{
			"username":  username,
			"addresses": len(addresses),
		}).Debug("ensuring legacy user exists")

		// Check if user exists
		_, err := store.GetUser(ctx, username)
		if err != nil {
			// User doesn't exist, create without persona
			_, err = store.CreateUser(ctx, username, addresses)
			if err != nil {
				return fmt.Errorf("failed to create legacy user %s: %w", username, err)
			}
			log.WithField("username", username).Info("created legacy user")
		}
	}

	return nil
}
