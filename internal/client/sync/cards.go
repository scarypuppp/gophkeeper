package sync

import (
	"github.com/scarypuppp/gophkeeper/internal/api"
	"github.com/scarypuppp/gophkeeper/internal/client/storage"
)

// cardChecksums собирает контрольные суммы локальных карт.
func cardChecksums(items []storage.Card) map[string]string {
	checksums := make(map[string]string, len(items))
	for _, item := range items {
		checksums[item.Name] = item.Checksum
	}
	return checksums
}

// cardItemChecksum возвращает контрольную сумму карты на сервере.
func cardItemChecksum(item api.CardItemResponse) string {
	return item.Checksum
}

// cardName возвращает имя карты — её идентификатор.
func cardName(item storage.Card) string {
	return item.Name
}

// cardConflict собирает описание конфликта по карте.
func (s *Syncer) cardConflict(st step) (Conflict, error) {
	conflict := Conflict{Kind: kindCard, Name: st.name}

	if st.local.Present {
		local, err := s.storage.GetCard(st.name)
		if err != nil {
			return conflict, err
		}
		conflict.Local = cardFields(local.Number, local.Holder, local.ExpiresAt, local.CVV, local.Metadata)
	}
	if st.remote.Present {
		remote, err := s.client.GetCard(s.token, st.name)
		if err != nil {
			return conflict, err
		}
		conflict.Remote = cardFields(remote.Number, remote.Holder, remote.ExpiresAt, remote.CVV, remote.Metadata)
	}
	return conflict, nil
}

// cardFields перечисляет поля карты для показа в конфликте.
func cardFields(number, holder, expires, cvv, metadata string) []Field {
	return []Field{
		{Name: "number", Value: number},
		{Name: "holder", Value: holder},
		{Name: "expires", Value: expires},
		{Name: "cvv", Value: cvv},
		{Name: "meta", Value: metadata},
	}
}

// applyCard приводит карту к выбранной версии.
func (s *Syncer) applyCard(st step, result *storage.Data) error {
	switch st.action {
	case ActionPush:
		local, err := s.storage.GetCard(st.name)
		if err != nil {
			return err
		}
		if st.remote.Present {
			_, err = s.client.UpdateCard(s.token, st.name, api.UpdateCardRequest{
				Number:    local.Number,
				Holder:    local.Holder,
				ExpiresAt: local.ExpiresAt,
				CVV:       local.CVV,
				Metadata:  local.Metadata,
			})
		} else {
			_, err = s.client.CreateCard(s.token, api.CreateCardRequest{
				Name:      local.Name,
				Number:    local.Number,
				Holder:    local.Holder,
				ExpiresAt: local.ExpiresAt,
				CVV:       local.CVV,
				Metadata:  local.Metadata,
			})
		}
		if err != nil {
			return err
		}
		s.report(kindCard, st.name, createdOrUpdated(st.remote.Present))

	case ActionPull:
		remote, err := s.client.GetCard(s.token, st.name)
		if err != nil {
			return err
		}
		result.Cards = upsert(result.Cards, cardName, storage.Card{
			Name:      remote.Name,
			Number:    remote.Number,
			Holder:    remote.Holder,
			ExpiresAt: remote.ExpiresAt,
			CVV:       remote.CVV,
			Metadata:  remote.Metadata,
			Checksum:  remote.Checksum,
			UpdatedAt: remote.UpdatedAt,
		})
		s.report(kindCard, st.name, createdOrUpdated(st.local.Present))

	case ActionDeleteRemote:
		if err := s.client.DeleteCard(s.token, st.name); err != nil {
			return err
		}
		s.report(kindCard, st.name, "deleted")

	case ActionDeleteLocal:
		result.Cards = remove(result.Cards, cardName, st.name)
		s.report(kindCard, st.name, "deleted")
	}
	return nil
}
