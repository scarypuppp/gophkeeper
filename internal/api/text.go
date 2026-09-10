package api

import "time"

type CreateTextRequest struct {
	Name     string `json:"name"`
	Text     string `json:"text"`
	Metadata string `json:"metadata"`
}

type UpdateTextRequest struct {
	Text     string `json:"text"`
	Metadata string `json:"metadata"`
}

type TextResponse struct {
	Name     string `json:"name"`
	Text     string `json:"text"`
	Metadata string `json:"metadata"`

	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TextItemResponse struct {
	Name     string `json:"name"`
	Metadata string `json:"metadata"`

	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetTextsResponse []TextItemResponse
