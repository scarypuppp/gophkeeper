package api

import "time"

// MaxFileUploadSize — максимальный размер загружаемого файла.
const MaxFileUploadSize = 1 << 30 // 1 GiB

type CreateFileResponse struct {
	FileName  string    `json:"file_name"`
	FileHash  string    `json:"file_hash"`
	Metadata  string    `json:"metadata"`
	Checksum  string    `json:"checksum"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateFileRequest struct {
	Metadata string `json:"metadata"`
}

type FileResponse struct {
	FileName  string    `json:"file_name"`
	FileHash  string    `json:"file_hash"`
	Metadata  string    `json:"metadata"`
	Checksum  string    `json:"checksum"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FileItemResponse struct {
	FileName  string    `json:"file_name"`
	FileHash  string    `json:"file_hash"`
	FileURL   string    `json:"file_url"`
	Metadata  string    `json:"metadata"`
	Checksum  string    `json:"checksum"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetFilesResponse []FileItemResponse
