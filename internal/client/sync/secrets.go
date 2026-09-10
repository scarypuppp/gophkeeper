package sync

import (
	"github.com/scarypuppp/gophkeeper/internal/api"
	"github.com/scarypuppp/gophkeeper/internal/client/storage"
)

// secretChecksums собирает контрольные суммы локальных секретов.
func secretChecksums(items []storage.Secret) map[string]string {
	checksums := make(map[string]string, len(items))
	for _, item := range items {
		checksums[item.Name] = item.Checksum
	}
	return checksums
}

// secretItemChecksum возвращает контрольную сумму секрета на сервере.
func secretItemChecksum(item api.SecretItemResponse) string {
	return item.Checksum
}

// secretName возвращает имя секрета — его идентификатор.
func secretName(item storage.Secret) string {
	return item.Name
}

// secretConflict собирает описание конфликта по секрету.
func (s *Syncer) secretConflict(st step) (Conflict, error) {
	conflict := Conflict{Kind: kindSecret, Name: st.name}

	if st.local.Present {
		local, err := s.storage.GetSecret(st.name)
		if err != nil {
			return conflict, err
		}
		conflict.Local = []Field{
			{Name: "login", Value: local.Login},
			{Name: "password", Value: local.Password},
			{Name: "meta", Value: local.Metadata},
		}
	}
	if st.remote.Present {
		remote, err := s.client.GetSecret(s.token, st.name)
		if err != nil {
			return conflict, err
		}
		conflict.Remote = []Field{
			{Name: "login", Value: remote.Login},
			{Name: "password", Value: remote.Password},
			{Name: "meta", Value: remote.Metadata},
		}
	}
	return conflict, nil
}

// applySecret приводит секрет к выбранной версии.
func (s *Syncer) applySecret(st step, result *storage.Data) error {
	switch st.action {
	case ActionPush:
		local, err := s.storage.GetSecret(st.name)
		if err != nil {
			return err
		}
		if st.remote.Present {
			_, err = s.client.UpdateSecret(s.token, st.name, api.UpdateSecretRequest{
				Login:    local.Login,
				Password: local.Password,
				Metadata: local.Metadata,
			})
		} else {
			_, err = s.client.CreateSecret(s.token, api.CreateSecretRequest{
				Name:     local.Name,
				Login:    local.Login,
				Password: local.Password,
				Metadata: local.Metadata,
			})
		}
		if err != nil {
			return err
		}
		s.report(kindSecret, st.name, createdOrUpdated(st.remote.Present))

	case ActionPull:
		remote, err := s.client.GetSecret(s.token, st.name)
		if err != nil {
			return err
		}
		result.Secrets = upsert(result.Secrets, secretName, storage.Secret{
			Name:      remote.Name,
			Login:     remote.Login,
			Password:  remote.Password,
			Metadata:  remote.Metadata,
			Checksum:  remote.Checksum,
			UpdatedAt: remote.UpdatedAt,
		})
		s.report(kindSecret, st.name, createdOrUpdated(st.local.Present))

	case ActionDeleteRemote:
		if err := s.client.DeleteSecret(s.token, st.name); err != nil {
			return err
		}
		s.report(kindSecret, st.name, "deleted")

	case ActionDeleteLocal:
		result.Secrets = remove(result.Secrets, secretName, st.name)
		s.report(kindSecret, st.name, "deleted")
	}
	return nil
}
