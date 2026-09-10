package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type File struct {
	Name string
	Hash string
}

type Session struct {
	Login     string
	AuthToken string
}

func (a *App) SetAuthData(login string, token string) error {
	if a.Session == nil {
		a.Session = &Session{}
	}
	a.Session.Login = login
	a.Session.AuthToken = token
	if err := a.saveSession(); err != nil {
		return err
	}
	return nil
}

func (a *App) ResetSession() error {
	a.Session = nil
	if err := a.saveSession(); err != nil {
		return err
	}
	return nil
}

func (a *App) saveSession() error {
	if a.Session == nil {
		err := os.Remove(a.Config.SessionFilePath)
		if err != nil {
			return fmt.Errorf("error removing session file: %w", err)
		}
		return nil
	}
	data, err := json.Marshal(a.Session)
	if err != nil {
		return fmt.Errorf("error marshal session: %w", err)
	}

	f, err := os.OpenFile(a.Config.SessionFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening session file: %w", err)
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("error writing session file: %w", err)
	}
	return nil
}

func LoadSession(sessionFile string) (*Session, error) {
	data, err := os.ReadFile(sessionFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading session file: %w", err)
	}
	var s Session
	err = json.Unmarshal(data, &s)
	if err != nil {
		return nil, fmt.Errorf("error unmarshal session: %w", err)
	}
	return &s, nil
}
