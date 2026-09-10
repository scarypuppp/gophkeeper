package main

import (
	"log"

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
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}
	s := app.NewServer(cfg)
	if err := s.Run(); err != nil {
		log.Fatalf("Error running server")
	}
}
