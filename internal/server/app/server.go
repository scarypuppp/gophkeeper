package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

// Run инициализирует зависимости, запускает HTTP-сервер и AccrualPoller.
// Блокирует выполнение до получения сигнала завершения (SIGINT/SIGTERM),
// после чего выполняет graceful shutdown с таймаутом 5 секунд.
func (s *Server) Run() error {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Could not create logger: %v", err)
	}
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := repository2.NewPostgresDB(s.Config.DatabaseURI)
	if err != nil {
		logger.Fatal("Could not create database connection", zap.Error(err))
	}

	if err := runMigrations(s.Config.DatabaseURI); err != nil {
		logger.Fatal("Error running migrations", zap.Error(err))
	}

	s3Storage, err := filestorage.NewS3Client(ctx, s.Config.S3Endpoint, s.Config.S3AccessKey, s.Config.S3SecretKey, s.Config.S3Bucket)
	if err != nil {
		logger.Fatal("Could not create S3 client", zap.Error(err))
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

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Could not start server", zap.Error(err))
			cancel()
		}
	}()
	logger.Info("Server started", zap.String("address", s.Config.Address))
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	cancel()

	shutdownCtx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()
	logger.Info("Server stopped gracefully")
	return srv.Shutdown(shutdownCtx)
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
