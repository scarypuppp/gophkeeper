package api

import "time"

type CreateCardRequest struct {
	Name      string `json:"name"`
	Number    string `json:"number"`
	Holder    string `json:"holder"`
	ExpiresAt string `json:"expires_at"`
	CVV       string `json:"cvv"`
	Metadata  string `json:"metadata"`
}

type UpdateCardRequest struct {
	Number    string `json:"number"`
	Holder    string `json:"holder"`
	ExpiresAt string `json:"expires_at"`
	CVV       string `json:"cvv"`
	Metadata  string `json:"metadata"`
}

type CardResponse struct {
	Name      string `json:"name"`
	Number    string `json:"number"`
	Holder    string `json:"holder"`
	ExpiresAt string `json:"expires_at"`
	CVV       string `json:"cvv"`
	Metadata  string `json:"metadata"`

	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CardItemResponse struct {
	Name     string `json:"name"`
	Number   string `json:"number"`
	Holder   string `json:"holder"`
	Metadata string `json:"metadata"`

	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetCardsResponse []CardItemResponse
