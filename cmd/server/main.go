package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/scarypuppp/gophkeeper/internal/server/app"
	"github.com/scarypuppp/gophkeeper/internal/server/config"

	_ "github.com/scarypuppp/gophkeeper/docs"
)

// @title GophKeeper API
// @version 1.0
// @description Менеджер паролей GophKeeper
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {token}
func main() {
	if err := run(); err != nil {
		log.Printf("gophkeeper server: %v", err)
		os.Exit(1)
	}
}

// run поднимает сервер и возвращает ошибку вызывающей стороне.
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}
	return app.NewServer(cfg).Run(ctx)
}
