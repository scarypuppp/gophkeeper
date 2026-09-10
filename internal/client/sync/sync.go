package sync

import (
	"fmt"
	"io"
	"slices"

	"github.com/scarypuppp/gophkeeper/internal/api"
	"github.com/scarypuppp/gophkeeper/internal/checksum"
	"github.com/scarypuppp/gophkeeper/internal/client/interactor"
	"github.com/scarypuppp/gophkeeper/internal/client/storage"
)

// Виды записей, которые синхронизируются.
const (
	kindFile   = "file"
	kindText   = "text"
	kindCard   = "card"
	kindSecret = "secret"
)

// step — решение по одной записи.
type step struct {
	kind   string
	name   string
	action Action
	local  State
	remote State
}

// remoteData — состояние данных на сервере. Списки не содержат самих секретов
// (пароля, текста, cvv), поэтому полная запись запрашивается отдельно —
// только когда её нужно показать или сохранить у себя.
type remoteData struct {
	files   map[string]api.FileItemResponse
	texts   map[string]api.TextItemResponse
	cards   map[string]api.CardItemResponse
	secrets map[string]api.SecretItemResponse
}

// Syncer приводит локальные данные и данные сервера к общему состоянию.
type Syncer struct {
	storage   *storage.LocalStorage
	base      *storage.LocalStorage
	client    *interactor.Interactor
	token     string
	downloads string
	resolver  Resolver
	out       io.Writer
}

// NewSyncer создаёт синхронизацию поверх локального хранилища, файла BASE
// и клиента сервера. downloads — каталог с содержимым файлов, resolver
// разрешает конфликты, out принимает отчёт о применённых изменениях.
func NewSyncer(
	local *storage.LocalStorage,
	base *storage.LocalStorage,
	client *interactor.Interactor,
	token string,
	downloads string,
	resolver Resolver,
	out io.Writer,
) *Syncer {
	return &Syncer{
		storage:   local,
		base:      base,
		client:    client,
		token:     token,
		downloads: downloads,
		resolver:  resolver,
		out:       out,
	}
}

// Run выполняет синхронизацию:
//
//  1. запрашивает у сервера полное состояние данных;
//  2. вычисляет решения по таблице local / base / remote;
//  3. разрешает конфликты;
//  4. отправляет изменения на сервер и забирает изменения сервера;
//  5. перезаписывает data_base.json.
//
// Изменения уходят на сервер только после того, как разрешены все конфликты.
// Если применение прервётся ошибкой, ни data.json, ни data_base.json не
// перезаписываются: следующий запуск повторит синхронизацию с того же места.
func (s *Syncer) Run() error {
	baseKnown, err := s.base.Exists()
	if err != nil {
		return err
	}
	if err := s.base.Load(); err != nil {
		return err
	}

	remote, err := s.fetchRemote()
	if err != nil {
		return err
	}

	local := s.storage.Snapshot()
	steps := s.plan(local, s.base.Snapshot(), remote, baseKnown)

	if err := s.resolveConflicts(steps, remote); err != nil {
		return err
	}

	result := local
	for _, st := range steps {
		if err := s.apply(st, remote, &result); err != nil {
			return fmt.Errorf("%s %q: %w", st.kind, st.name, err)
		}
	}

	if err := s.storage.Replace(result); err != nil {
		return err
	}
	return s.base.Replace(result)
}

// fetchRemote запрашивает у сервера состояние всех типов данных.
func (s *Syncer) fetchRemote() (remoteData, error) {
	var remote remoteData

	files, err := s.client.GetFiles(s.token)
	if err != nil {
		return remote, err
	}
	remote.files = make(map[string]api.FileItemResponse, len(files))
	for _, file := range files {
		remote.files[file.FileName] = file
	}

	texts, err := s.client.GetTexts(s.token)
	if err != nil {
		return remote, err
	}
	remote.texts = make(map[string]api.TextItemResponse, len(texts))
	for _, text := range texts {
		remote.texts[text.Name] = text
	}

	cards, err := s.client.GetCards(s.token)
	if err != nil {
		return remote, err
	}
	remote.cards = make(map[string]api.CardItemResponse, len(cards))
	for _, card := range cards {
		remote.cards[card.Name] = card
	}

	secrets, err := s.client.GetSecrets(s.token)
	if err != nil {
		return remote, err
	}
	remote.secrets = make(map[string]api.SecretItemResponse, len(secrets))
	for _, secret := range secrets {
		remote.secrets[secret.Name] = secret
	}

	return remote, nil
}

// plan вычисляет решения по всем записям всех типов.
func (s *Syncer) plan(local storage.Data, base storage.Data, remote remoteData, baseKnown bool) []step {
	steps := make([]step, 0)
	steps = append(steps, planKind(kindFile,
		fileChecksums(local.Files), fileChecksums(base.Files), remoteFileChecksums(remote.files), baseKnown)...)
	steps = append(steps, planKind(kindText,
		textChecksums(local.Texts), textChecksums(base.Texts), remoteChecksums(remote.texts, textItemChecksum), baseKnown)...)
	steps = append(steps, planKind(kindCard,
		cardChecksums(local.Cards), cardChecksums(base.Cards), remoteChecksums(remote.cards, cardItemChecksum), baseKnown)...)
	steps = append(steps, planKind(kindSecret,
		secretChecksums(local.Secrets), secretChecksums(base.Secrets), remoteChecksums(remote.secrets, secretItemChecksum), baseKnown)...)
	return steps
}

// planKind принимает решения по записям одного типа, сравнивая контрольные суммы.
func planKind(kind string, local, base, remote map[string]string, baseKnown bool) []step {
	steps := make([]step, 0, len(local)+len(remote))
	for _, name := range unionNames(local, base, remote) {
		localState := stateOf(local, name)
		remoteState := stateOf(remote, name)

		action := Decide(localState, stateOf(base, name), remoteState, baseKnown)
		if action == ActionNone {
			continue
		}
		steps = append(steps, step{kind: kind, name: name, action: action, local: localState, remote: remoteState})
	}
	return steps
}

// resolveConflicts спрашивает пользователя по каждой конфликтной записи
// и заменяет конфликт конкретным действием.
func (s *Syncer) resolveConflicts(steps []step, remote remoteData) error {
	for i, st := range steps {
		if st.action != ActionConflict {
			continue
		}

		conflict, err := s.conflict(st, remote)
		if err != nil {
			return fmt.Errorf("%s %q: %w", st.kind, st.name, err)
		}
		side, err := s.resolver.Resolve(conflict)
		if err != nil {
			return fmt.Errorf("%s %q: %w", st.kind, st.name, err)
		}
		steps[i].action = Resolve(side, st.local, st.remote)
	}
	return nil
}

// conflict собирает описание конфликтной записи для показа пользователю.
func (s *Syncer) conflict(st step, remote remoteData) (Conflict, error) {
	switch st.kind {
	case kindFile:
		return s.fileConflict(st, remote)
	case kindText:
		return s.textConflict(st)
	case kindCard:
		return s.cardConflict(st)
	case kindSecret:
		return s.secretConflict(st)
	}
	return Conflict{}, fmt.Errorf("unknown record kind %q", st.kind)
}

// apply выполняет одно решение: меняет данные на сервере или в result.
func (s *Syncer) apply(st step, remote remoteData, result *storage.Data) error {
	switch st.kind {
	case kindFile:
		return s.applyFile(st, remote, result)
	case kindText:
		return s.applyText(st, result)
	case kindCard:
		return s.applyCard(st, result)
	case kindSecret:
		return s.applySecret(st, result)
	}
	return fmt.Errorf("unknown record kind %q", st.kind)
}

// report печатает строку отчёта о применённом изменении.
func (s *Syncer) report(kind string, name string, verb string) {
	fmt.Fprintf(s.out, "%s %s %s\n", kind, name, verb)
}

// warn печатает предупреждение, не прерывающее синхронизацию.
func (s *Syncer) warn(format string, args ...any) {
	fmt.Fprintf(s.out, "Warning: "+format+"\n", args...)
}

// createdOrUpdated возвращает глагол отчёта в зависимости от того,
// существовала ли запись на принимающей стороне.
func createdOrUpdated(existed bool) string {
	if existed {
		return "updated"
	}
	return "created"
}

// stateOf возвращает присутствие записи и её контрольную сумму.
func stateOf(checksums map[string]string, name string) State {
	sum, ok := checksums[name]
	return State{Present: ok, Checksum: sum}
}

// unionNames возвращает отсортированное объединение имён записей.
func unionNames(sets ...map[string]string) []string {
	seen := make(map[string]bool)
	names := make([]string, 0)
	for _, set := range sets {
		for name := range set {
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	slices.Sort(names)
	return names
}

// remoteChecksums собирает контрольные суммы записей сервера.
func remoteChecksums[T any](items map[string]T, sum func(T) string) map[string]string {
	checksums := make(map[string]string, len(items))
	for name, item := range items {
		checksums[name] = sum(item)
	}
	return checksums
}

// upsert заменяет запись с тем же именем или добавляет новую.
func upsert[T any](items []T, name func(T) string, value T) []T {
	i := slices.IndexFunc(items, func(item T) bool { return name(item) == name(value) })
	if i < 0 {
		return append(items, value)
	}
	items[i] = value
	return items
}

// remove убирает запись с указанным именем.
func remove[T any](items []T, name func(T) string, target string) []T {
	i := slices.IndexFunc(items, func(item T) bool { return name(item) == target })
	if i < 0 {
		return items
	}
	return slices.Delete(items, i, i+1)
}

// contentChecksum склеивает контрольную сумму содержимого с метаданными:
// метаданные в сумму содержимого не входят, но их изменение тоже нужно
// синхронизировать.
func contentChecksum(sum string, metadata string) string {
	return checksum.Content(sum, metadata)
}
