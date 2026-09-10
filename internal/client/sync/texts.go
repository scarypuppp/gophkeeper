package sync

import (
	"github.com/scarypuppp/gophkeeper/internal/api"
	"github.com/scarypuppp/gophkeeper/internal/client/storage"
)

// textChecksums собирает контрольные суммы локальных текстовых записей.
func textChecksums(items []storage.Text) map[string]string {
	checksums := make(map[string]string, len(items))
	for _, item := range items {
		checksums[item.Name] = item.Checksum
	}
	return checksums
}

// textItemChecksum возвращает контрольную сумму текстовой записи на сервере.
func textItemChecksum(item api.TextItemResponse) string {
	return item.Checksum
}

// textName возвращает имя текстовой записи — её идентификатор.
func textName(item storage.Text) string {
	return item.Name
}

// textConflict собирает описание конфликта по текстовой записи.
func (s *Syncer) textConflict(st step) (Conflict, error) {
	conflict := Conflict{Kind: kindText, Name: st.name}

	if st.local.Present {
		local, err := s.storage.GetText(st.name)
		if err != nil {
			return conflict, err
		}
		conflict.Local = []Field{
			{Name: "text", Value: local.Text},
			{Name: "meta", Value: local.Metadata},
		}
	}
	if st.remote.Present {
		remote, err := s.client.GetText(s.token, st.name)
		if err != nil {
			return conflict, err
		}
		conflict.Remote = []Field{
			{Name: "text", Value: remote.Text},
			{Name: "meta", Value: remote.Metadata},
		}
	}
	return conflict, nil
}

// applyText приводит текстовую запись к выбранной версии.
func (s *Syncer) applyText(st step, result *storage.Data) error {
	switch st.action {
	case ActionPush:
		local, err := s.storage.GetText(st.name)
		if err != nil {
			return err
		}
		if st.remote.Present {
			_, err = s.client.UpdateText(s.token, st.name, api.UpdateTextRequest{
				Text:     local.Text,
				Metadata: local.Metadata,
			})
		} else {
			_, err = s.client.CreateText(s.token, api.CreateTextRequest{
				Name:     local.Name,
				Text:     local.Text,
				Metadata: local.Metadata,
			})
		}
		if err != nil {
			return err
		}
		s.report(kindText, st.name, createdOrUpdated(st.remote.Present))

	case ActionPull:
		remote, err := s.client.GetText(s.token, st.name)
		if err != nil {
			return err
		}
		result.Texts = upsert(result.Texts, textName, storage.Text{
			Name:      remote.Name,
			Text:      remote.Text,
			Metadata:  remote.Metadata,
			Checksum:  remote.Checksum,
			UpdatedAt: remote.UpdatedAt,
		})
		s.report(kindText, st.name, createdOrUpdated(st.local.Present))

	case ActionDeleteRemote:
		if err := s.client.DeleteText(s.token, st.name); err != nil {
			return err
		}
		s.report(kindText, st.name, "deleted")

	case ActionDeleteLocal:
		result.Texts = remove(result.Texts, textName, st.name)
		s.report(kindText, st.name, "deleted")
	}
	return nil
}
