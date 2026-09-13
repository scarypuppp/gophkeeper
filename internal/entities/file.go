package entities

import "time"

type File struct {
	Id              int64     `db:"id"`
	Owner           int64     `db:"owner"`
	FileName        string    `db:"file_name"`
	FileHash        string    `db:"file_hash"`
	StorageFilePath string    `db:"storage_file_path"`
	Metadata        string    `db:"metadata"`
	Checksum        string    `db:"checksum"`
	Size            int64     `db:"size"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}
