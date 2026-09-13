package cli

import (
	"fmt"
	"io"
	"strings"
)

// progressBarWidth — длина полосы прогресса в символах.
const progressBarWidth = 30

// progressBar печатает ход скачивания. Сам данные никуда не пишет:
// его подставляют рядом с файлом через io.MultiWriter и считают байты.
type progressBar struct {
	out     io.Writer
	total   int64
	written int64
	shown   int
}

// newProgressBar создаёт полосу прогресса для файла размером total байт.
func newProgressBar(out io.Writer, total int64) *progressBar {
	bar := &progressBar{out: out, total: total, shown: -1}
	bar.render()
	return bar
}

// Write учитывает очередную порцию скачанных байт.
func (b *progressBar) Write(p []byte) (int, error) {
	b.written += int64(len(p))
	b.render()
	return len(p), nil
}

// done дорисовывает полосу до конца и переводит строку.
func (b *progressBar) done() {
	b.written = b.total
	b.render()
	fmt.Fprintln(b.out)
}

// render перерисовывает полосу, если процент изменился.
func (b *progressBar) render() {
	percent := 100
	if b.total > 0 {
		percent = int(b.written * 100 / b.total)
	}
	if percent == b.shown {
		return
	}
	b.shown = percent

	filled := progressBarWidth * percent / 100
	fmt.Fprintf(b.out, "\r[%s%s] %3d%%",
		strings.Repeat("=", filled),
		strings.Repeat(" ", progressBarWidth-filled),
		percent,
	)
}
