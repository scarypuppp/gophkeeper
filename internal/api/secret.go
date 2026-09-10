package api

import "time"

type CreateSecretRequest struct {
	Name     string `json:"name"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Metadata string `json:"metadata"`
}

type UpdateSecretRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Metadata string `json:"metadata"`
}

type SecretResponse struct {
	Name     string `json:"name"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Metadata string `json:"metadata"`

	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SecretItemResponse struct {
	Name     string `json:"name"`
	Login    string `json:"login"`
	Metadata string `json:"metadata"`

	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetSecretsResponse []SecretItemResponse
