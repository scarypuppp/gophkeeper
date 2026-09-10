package entities

import "time"

type Text struct {
	Id        int64     `db:"id"`
	Owner     int64     `db:"owner"`
	Name      string    `db:"name"`
	Text      string    `db:"text"`
	Metadata  string    `db:"metadata"`
	Checksum  string    `db:"checksum"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
