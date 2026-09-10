package entities

import "time"

type Card struct {
	Id        int64     `db:"id"`
	Owner     int64     `db:"owner"`
	Name      string    `db:"name"`
	Number    string    `db:"number"`
	Holder    string    `db:"holder"`
	ExpiresAt string    `db:"expires_at"`
	CVV       string    `db:"cvv"`
	Metadata  string    `db:"metadata"`
	Checksum  string    `db:"checksum"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
