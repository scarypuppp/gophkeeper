// Package buildinfo хранит сведения о сборке клиента.
//
// Значения подставляет линковщик при сборке, например:
//
//	go build -ldflags "-X github.com/scarypuppp/gophkeeper/internal/buildinfo.version=v1.0.0" ./cmd/client
//
// Если значение не задано, возвращается N/A: собранный вручную бинарник
// не должен выдавать себя за версию из релиза.
package buildinfo

// unknown подставляется вместо значения, не заданного при сборке.
const unknown = "N/A"

var (
	version string
	date    string
	commit  string
)

// Version возвращает версию клиента.
func Version() string {
	return orUnknown(version)
}

// Date возвращает дату сборки.
func Date() string {
	return orUnknown(date)
}

// Commit возвращает коммит, из которого собран клиент.
func Commit() string {
	return orUnknown(commit)
}

// orUnknown заменяет пустое значение на N/A.
func orUnknown(value string) string {
	if value == "" {
		return unknown
	}
	return value
}
