package cli

import (
	"os"
	"text/tabwriter"
)

// metaPreviewLen — длина превью длинных полей (текст, метаданные) в списках.
const metaPreviewLen = 20

// newTable создаёт writer, выравнивающий разделённые табуляцией колонки списков.
// Вызывающий обязан вызвать Flush.
func newTable() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
}

// truncate обрезает строку до limit символов, добавляя многоточие.
// Короткие строки возвращаются без изменений.
func truncate(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit]) + "..."
}

// maskCardNumber скрывает номер карты, оставляя последние четыре цифры.
func maskCardNumber(number string) string {
	const visible = 4
	runes := []rune(number)
	if len(runes) <= visible {
		return number
	}
	return "..." + string(runes[len(runes)-visible:])
}

// storedMark возвращает отметку о наличии файла на локальном диске.
func storedMark(stored bool) string {
	if stored {
		return "✅"
	}
	return "❌"
}
