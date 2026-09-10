package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// Content считает SHA-256 содержимого записи.
// Части склеиваются с префиксом длины, поэтому разные наборы значений
// не могут дать одинаковый вход хэш-функции.
func Content(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		fmt.Fprintf(h, "%d:%s", len(part), part)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Reader считает SHA-256 содержимого файла. Сервер хранит контрольную сумму
// файла в том же виде, поэтому суммы локальной копии и копии в хранилище
// можно сравнивать напрямую.
func Reader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Secret считает контрольную сумму содержимого пары логин-пароль.
func Secret(login, password, metadata string) string {
	return Content(login, password, metadata)
}

// Card считает контрольную сумму содержимого банковской карты.
func Card(number, holder, expiresAt, cvv, metadata string) string {
	return Content(number, holder, expiresAt, cvv, metadata)
}

// Text считает контрольную сумму содержимого текстовой записи.
func Text(text, metadata string) string {
	return Content(text, metadata)
}
