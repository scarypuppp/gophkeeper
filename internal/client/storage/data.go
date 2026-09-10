package storage

import (
	"cmp"
	"slices"
	"time"
)

// Data — формат файла локального хранилища.
type Data struct {
	Files   []File   `json:"Files"`
	Texts   []Text   `json:"Texts"`
	Cards   []Card   `json:"Cards"`
	Secrets []Secret `json:"Secrets"`
}

// File — запись о бинарных данных. Содержимое файла лежит в каталоге downloads,
type File struct {
	FileName  string    `json:"file_name"`
	FileHash  string    `json:"file_hash"`
	Size      int64     `json:"size"`
	Metadata  string    `json:"meta"`
	Checksum  string    `json:"checksum"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Text — произвольные текстовые данные.
type Text struct {
	Name      string    `json:"name"`
	Text      string    `json:"text"`
	Metadata  string    `json:"meta"`
	Checksum  string    `json:"checksum"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Card — данные банковской карты.
type Card struct {
	Name      string    `json:"name"`
	Number    string    `json:"number"`
	Holder    string    `json:"holder"`
	ExpiresAt string    `json:"expires"`
	CVV       string    `json:"cvv"`
	Metadata  string    `json:"meta"`
	Checksum  string    `json:"checksum"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Secret — пара логин-пароль.
type Secret struct {
	Name      string    `json:"name"`
	Login     string    `json:"login"`
	Password  string    `json:"password"`
	Metadata  string    `json:"meta"`
	Checksum  string    `json:"checksum"`
	UpdatedAt time.Time `json:"updated_at"`
}

// record — запись хранилища, идентифицируемая уникальным в рамках типа именем.
type record interface {
	File | Text | Card | Secret
}

// recordName возвращает имя записи — её идентификатор внутри своего типа.
func recordName[T record](item T) string {
	switch v := any(item).(type) {
	case File:
		return v.FileName
	case Text:
		return v.Name
	case Card:
		return v.Name
	case Secret:
		return v.Name
	}
	return ""
}

// indexOf возвращает позицию записи с именем name или -1, если её нет.
func indexOf[T record](items []T, name string) int {
	return slices.IndexFunc(items, func(item T) bool { return recordName(item) == name })
}

// sortedByName возвращает копию записей, упорядоченную по имени, чтобы вывод
// списков не зависел от порядка записей в файле.
func sortedByName[T record](items []T) []T {
	sorted := slices.Clone(items)
	slices.SortFunc(sorted, func(a, b T) int { return cmp.Compare(recordName(a), recordName(b)) })
	return sorted
}
