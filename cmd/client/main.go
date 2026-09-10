package main

import (
	"fmt"
	"os"

	"github.com/scarypuppp/gophkeeper/internal/client/app"
	"github.com/scarypuppp/gophkeeper/internal/client/cli"
	"github.com/scarypuppp/gophkeeper/internal/client/config"
	"github.com/spf13/cobra"
)

func main() {
	// Приложение готовится перед самой командой: version, help и справка
	// должны работать и без настроенного адреса сервера.
	a := &app.App{}
	root := cli.RootCommand(a)
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if !needsApp(cmd) {
			return nil
		}
		c, err := config.GetAgentConfig()
		if err != nil {
			return err
		}
		return a.Init(c)
	}

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// needsApp сообщает, нужны ли команде конфигурация и локальные данные.
func needsApp(cmd *cobra.Command) bool {
	if !cmd.HasParent() {
		// Корневая команда только печатает справку.
		return false
	}
	for c := cmd; c.HasParent(); c = c.Parent() {
		switch c.Name() {
		case "version", "help", "completion":
			return false
		}
	}
	return true
}
