package app

import (
	"net/http"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/scarypuppp/gophkeeper/internal/client/config"
	"github.com/scarypuppp/gophkeeper/internal/client/interactor"
	"github.com/scarypuppp/gophkeeper/internal/client/storage"
)

type App struct {
	Config     config.AgentConfig
	Interactor *interactor.Interactor
	Httpclient *resty.Client
	Session    *Session
	// Storage — локальные данные пользователя (data.json), уже загруженные с диска.
	Storage *storage.LocalStorage
	// BaseStorage — состояние на момент последней синхронизации (data_base.json).
	// Его читает синхронизация: ей важно отличать отсутствующий файл от пустого.
	BaseStorage *storage.LocalStorage
}

// Init заполняет приложение: клиент сервера, сессию и локальные хранилища.
// Отдельно от создания структуры, чтобы команды вроде version работали
// без конфигурации и файлов на диске.
func (a *App) Init(config config.AgentConfig) error {
	httpclient := resty.NewWithClient(&http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	})
	httpclient.SetBaseURL(config.BaseURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip")

	i := interactor.NewInteractor(httpclient)

	session, err := LoadSession(config.SessionFilePath)
	if err != nil {
		return err
	}

	// Создание директории с загруженными файлами
	err = os.MkdirAll(config.DownloadedFilesPath, 0755)
	if err != nil {
		return err
	}

	localStorage := storage.NewLocalStorage(config.DataFilePath)
	if err := localStorage.Load(); err != nil {
		return err
	}

	*a = App{
		Config:      config,
		Interactor:  i,
		Httpclient:  httpclient,
		Session:     session,
		Storage:     localStorage,
		BaseStorage: storage.NewLocalStorage(config.BaseDataFilePath),
	}
	return nil
}
