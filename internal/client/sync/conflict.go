package sync

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// ErrConflictNotResolved возвращается, если выбрать версию записи не удалось.
var ErrConflictNotResolved = errors.New("conflict is not resolved")

// Field — одно поле записи в описании конфликта.
type Field struct {
	Name  string
	Value string
}

// Conflict описывает запись, изменённую и локально, и на сервере.
// Пустой список полей означает, что на этой стороне запись удалена.
type Conflict struct {
	Kind   string
	Name   string
	Local  []Field
	Remote []Field
}

// Resolver выбирает версию записи при конфликте.
type Resolver interface {
	Resolve(conflict Conflict) (Side, error)
}

// FixedResolver разрешает все конфликты в пользу одной стороны:
// его ставят флаги --theirs и --yours.
type FixedResolver struct {
	side Side
}

// TakeLocal возвращает resolver, выбирающий локальные версии (--yours).
func TakeLocal() FixedResolver {
	return FixedResolver{side: SideLocal}
}

// TakeRemote возвращает resolver, выбирающий версии сервера (--theirs).
func TakeRemote() FixedResolver {
	return FixedResolver{side: SideRemote}
}

// Resolve возвращает заранее выбранную сторону.
func (r FixedResolver) Resolve(Conflict) (Side, error) {
	return r.side, nil
}

// PromptResolver показывает версии рядом и спрашивает пользователя по каждой записи.
type PromptResolver struct {
	in  *bufio.Reader
	out io.Writer
}

// NewPromptResolver создаёт resolver, задающий вопрос по каждому конфликту.
func NewPromptResolver(in io.Reader, out io.Writer) *PromptResolver {
	return &PromptResolver{in: bufio.NewReader(in), out: out}
}

// Resolve печатает версии рядом и ждёт выбора: s — серверная, l — локальная.
func (r *PromptResolver) Resolve(conflict Conflict) (Side, error) {
	fmt.Fprintf(r.out, "\nConflict: %s %q was changed both locally and on the server\n", conflict.Kind, conflict.Name)
	printSideBySide(r.out, conflict)

	for {
		fmt.Fprint(r.out, "Keep [s]erver or [l]ocal version? ")
		answer, err := r.in.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return SideLocal, fmt.Errorf("read answer: %w", err)
		}

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "s":
			return SideRemote, nil
		case "l":
			return SideLocal, nil
		}
		if errors.Is(err, io.EOF) {
			return SideLocal, ErrConflictNotResolved
		}
		fmt.Fprintln(r.out, "Please answer 's' or 'l'.")
	}
}

// printSideBySide печатает локальную и серверную версии записи в две колонки.
func printSideBySide(out io.Writer, conflict Conflict) {
	table := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	fmt.Fprintln(table, "\tlocal (l)\tserver (s)")

	for _, name := range fieldNames(conflict) {
		fmt.Fprintf(table, "%s\t%s\t%s\n", name, fieldValue(conflict.Local, name), fieldValue(conflict.Remote, name))
	}
	table.Flush()
	fmt.Fprintln(out)
}

// fieldNames возвращает имена полей в порядке их появления в версиях записи.
func fieldNames(conflict Conflict) []string {
	names := make([]string, 0, len(conflict.Local)+len(conflict.Remote))
	seen := make(map[string]bool, cap(names))
	for _, fields := range [][]Field{conflict.Local, conflict.Remote} {
		for _, field := range fields {
			if !seen[field.Name] {
				seen[field.Name] = true
				names = append(names, field.Name)
			}
		}
	}
	return names
}

// fieldValue возвращает значение поля версии; для удалённой записи — прочерк.
func fieldValue(fields []Field, name string) string {
	if len(fields) == 0 {
		return "(deleted)"
	}
	for _, field := range fields {
		if field.Name == name {
			return field.Value
		}
	}
	return ""
}
