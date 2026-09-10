package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/scarypuppp/gophkeeper/internal/checksum"
	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/repository"
	"github.com/scarypuppp/gophkeeper/internal/server/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupCardMocks(t *testing.T) (*mocks.MockUnitOfWork, *mocks.MockCardRepository) {
	ctrl := gomock.NewController(t)
	uowMock := mocks.NewMockUnitOfWork(ctrl)
	cardRepoMock := mocks.NewMockCardRepository(ctrl)
	return uowMock, cardRepoMock
}

func TestGetCard(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)
		existing := &entities.Card{Id: 1, Owner: 1, Name: "name"}

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(existing, nil)

		s := NewCardsService(uowMock)
		card, err := s.GetCard(ctx, 1, "name")

		assert.NoError(t, err)
		assert.Equal(t, existing, card)
	})

	t.Run("not exist", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(nil, repository.ErrNoRows)

		s := NewCardsService(uowMock)
		card, err := s.GetCard(ctx, 1, "name")

		assert.ErrorIs(t, err, ErrCardNotExists)
		assert.Nil(t, card)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(nil, wantErr)

		s := NewCardsService(uowMock)
		card, err := s.GetCard(ctx, 1, "name")

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, card)
	})
}

func TestGetCards(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)
		want := []entities.Card{{Id: 1, Owner: 1, Name: "a"}, {Id: 2, Owner: 1, Name: "b"}}

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().GetCardsByOwner(ctx, int64(1)).Return(want, nil)

		s := NewCardsService(uowMock)
		cards, err := s.GetCards(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, want, cards)
	})

	t.Run("error", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().GetCardsByOwner(ctx, int64(1)).Return(nil, wantErr)

		s := NewCardsService(uowMock)
		cards, err := s.GetCards(ctx, 1)

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, cards)
	})
}

func TestCreateCard(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)
		created := &entities.Card{Id: 1, Owner: 1, Name: "name", Number: "4111", Holder: "holder", ExpiresAt: "12/30", CVV: "123", Metadata: "meta"}

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().
			CreateCard(ctx, entities.Card{
				Owner:     1,
				Name:      "name",
				Number:    "4111",
				Holder:    "holder",
				ExpiresAt: "12/30",
				CVV:       "123",
				Metadata:  "meta",
				Checksum:  checksum.Card("4111", "holder", "12/30", "123", "meta"),
			}).
			Return(created, nil)

		s := NewCardsService(uowMock)
		card, err := s.CreateCard(ctx, 1, "name", "4111", "holder", "12/30", "123", "meta")

		assert.NoError(t, err)
		assert.Equal(t, created, card)
	})

	t.Run("already exists", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().
			CreateCard(ctx, gomock.Any()).
			Return(nil, repository.ErrUniqueViolation)

		s := NewCardsService(uowMock)
		card, err := s.CreateCard(ctx, 1, "name", "4111", "holder", "12/30", "123", "meta")

		assert.ErrorIs(t, err, ErrCardAlreadyExists)
		assert.Nil(t, card)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().
			CreateCard(ctx, gomock.Any()).
			Return(nil, wantErr)

		s := NewCardsService(uowMock)
		card, err := s.CreateCard(ctx, 1, "name", "4111", "holder", "12/30", "123", "meta")

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, card)
	})
}

func TestUpdateCard(t *testing.T) {
	ctx := context.Background()
	createdAt := time.Now().Add(-time.Hour)

	t.Run("success", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)
		current := &entities.Card{Id: 1, Owner: 1, Name: "name", CreatedAt: createdAt}
		updatedAt := time.Now()

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(current, nil)

		expected := entities.Card{
			Id:        1,
			Owner:     1,
			Name:      "name",
			Number:    "newNumber",
			Holder:    "newHolder",
			ExpiresAt: "01/31",
			CVV:       "999",
			Metadata:  "newMeta",
			Checksum:  checksum.Card("newNumber", "newHolder", "01/31", "999", "newMeta"),
			CreatedAt: createdAt,
		}
		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().UpdateCard(ctx, expected).Return(updatedAt, nil)

		s := NewCardsService(uowMock)
		card, err := s.UpdateCard(ctx, 1, "name", "newNumber", "newHolder", "01/31", "999", "newMeta")

		assert.NoError(t, err)
		assert.Equal(t, expected.Id, card.Id)
		assert.Equal(t, expected.Number, card.Number)
		assert.Equal(t, updatedAt, card.UpdatedAt)
	})

	t.Run("current not exist", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(nil, repository.ErrNoRows)

		s := NewCardsService(uowMock)
		card, err := s.UpdateCard(ctx, 1, "name", "number", "holder", "01/31", "999", "meta")

		assert.ErrorIs(t, err, ErrCardNotExists)
		assert.Nil(t, card)
	})

	t.Run("update not exist", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)
		current := &entities.Card{Id: 1, Owner: 1, Name: "name", CreatedAt: createdAt}

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(current, nil)

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().UpdateCard(ctx, gomock.Any()).Return(time.Time{}, repository.ErrNoRows)

		s := NewCardsService(uowMock)
		card, err := s.UpdateCard(ctx, 1, "name", "number", "holder", "01/31", "999", "meta")

		assert.ErrorIs(t, err, ErrCardNotExists)
		assert.Nil(t, card)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)
		current := &entities.Card{Id: 1, Owner: 1, Name: "name", CreatedAt: createdAt}
		wantErr := errors.New("db error")

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(current, nil)

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().UpdateCard(ctx, gomock.Any()).Return(time.Time{}, wantErr)

		s := NewCardsService(uowMock)
		card, err := s.UpdateCard(ctx, 1, "name", "number", "holder", "01/31", "999", "meta")

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, card)
	})
}

func TestDeleteCard(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().DeleteCard(ctx, int64(1), "name").Return(nil)

		s := NewCardsService(uowMock)
		err := s.DeleteCard(ctx, 1, "name")

		assert.NoError(t, err)
	})

	t.Run("not exist", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().DeleteCard(ctx, int64(1), "name").Return(repository.ErrNoRows)

		s := NewCardsService(uowMock)
		err := s.DeleteCard(ctx, 1, "name")

		assert.ErrorIs(t, err, ErrCardNotExists)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, cardRepoMock := setupCardMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Cards().Return(cardRepoMock)
		cardRepoMock.EXPECT().DeleteCard(ctx, int64(1), "name").Return(wantErr)

		s := NewCardsService(uowMock)
		err := s.DeleteCard(ctx, 1, "name")

		assert.ErrorIs(t, err, wantErr)
	})
}
