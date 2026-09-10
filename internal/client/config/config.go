package config

import (
	"fmt"
	"os"
)

// serverURLEnv — переменная окружения с адресом сервера. Обязательна:
// клиент без сервера не знает, куда ходить, а молчаливый localhost
// приводил бы к синхронизации не с тем стендом.
const serverURLEnv = "SERVER_URL"

type AgentConfig struct {
	BaseURL             string
	SessionFilePath     string
	DownloadedFilesPath string
	// DataFilePath — локальные данные пользователя, которые правят команды клиента.
	DataFilePath string
	// BaseDataFilePath — состояние данных на момент последней синхронизации (BASE).
	BaseDataFilePath string
}

// GetAgentConfig собирает конфигурацию клиента.
// Возвращает ошибку, если не задан адрес сервера.
func GetAgentConfig() (AgentConfig, error) {
	baseURL := os.Getenv(serverURLEnv)
	if baseURL == "" {
		return AgentConfig{}, fmt.Errorf("environment variable %s is required, for example %s=http://localhost:8081", serverURLEnv, serverURLEnv)
	}

	return AgentConfig{
		BaseURL:             baseURL,
		SessionFilePath:     "./session.json",
		DownloadedFilesPath: "./downloads",
		DataFilePath:        "./data.json",
		BaseDataFilePath:    "./data_base.json",
	}, nil
}
