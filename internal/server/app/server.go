package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/scarypuppp/gophkeeper/internal/server/config"
	"github.com/scarypuppp/gophkeeper/internal/server/filestorage"
	"github.com/scarypuppp/gophkeeper/internal/server/handlers"
	repository2 "github.com/scarypuppp/gophkeeper/internal/server/repository"
	"github.com/scarypuppp/gophkeeper/internal/server/service"
	"go.uber.org/zap"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Server хранит конфигурацию приложения и является точкой входа для запуска всех компонентов.
type Server struct {
	Config *config.Config
}

// NewServer создаёт новый Server с переданной конфигурацией.
func NewServer(config *config.Config) *Server {
	return &Server{config}
}

// shutdownTimeout ограничивает время на завершение начатых запросов.
const shutdownTimeout = 5 * time.Second

// Run инициализирует зависимости и запускает HTTP-сервер.
func (s *Server) Run(ctx context.Context) error {
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("could not create logger: %w", err)
	}
	defer logger.Sync()

	db, err := repository2.NewPostgresDB(s.Config.DatabaseURI)
	if err != nil {
		return fmt.Errorf("could not create database connection: %w", err)
	}

	if err := runMigrations(s.Config.DatabaseURI); err != nil {
		return fmt.Errorf("error running migrations: %w", err)
	}

	s3Storage, err := filestorage.NewS3Client(ctx, s.Config.S3Endpoint, s.Config.S3AccessKey, s.Config.S3SecretKey, s.Config.S3Bucket)
	if err != nil {
		return fmt.Errorf("could not create S3 client: %w", err)
	}

	uow := repository2.NewUnitOfWorkPostgres(db)
	userService := service.NewUserService(uow)
	fileService := service.NewFileService(uow, s3Storage)
	secretsService := service.NewSecretsService(uow)
	cardsService := service.NewCardsService(uow)
	textsService := service.NewTextsService(uow)
	handler := handlers.NewHandler(s.Config, logger, userService, fileService, secretsService, cardsService, textsService)

	srv := &http.Server{
		Addr:         s.Config.Address,
		Handler:      handler.GetRouter(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	srvErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()
	logger.Info("Server started", zap.String("address", s.Config.Address))

	select {
	case err := <-srvErr:
		return fmt.Errorf("could not start server: %w", err)
	case <-ctx.Done():
		logger.Info("Shutdown signal received")
	}

	shutdownCtx, shutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	logger.Info("Server stopped gracefully")
	return nil
}

func runMigrations(dsn string) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return fmt.Errorf("create migrate: %w", err)
	}
	defer m.Close()

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
