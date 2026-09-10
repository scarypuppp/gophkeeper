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

func setupTextMocks(t *testing.T) (*mocks.MockUnitOfWork, *mocks.MockTextRepository) {
	ctrl := gomock.NewController(t)
	uowMock := mocks.NewMockUnitOfWork(ctrl)
	textRepoMock := mocks.NewMockTextRepository(ctrl)
	return uowMock, textRepoMock
}

func TestGetText(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)
		existing := &entities.Text{Id: 1, Owner: 1, Name: "name"}

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(existing, nil)

		s := NewTextsService(uowMock)
		text, err := s.GetText(ctx, 1, "name")

		assert.NoError(t, err)
		assert.Equal(t, existing, text)
	})

	t.Run("not exist", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(nil, repository.ErrNoRows)

		s := NewTextsService(uowMock)
		text, err := s.GetText(ctx, 1, "name")

		assert.ErrorIs(t, err, ErrTextNotExists)
		assert.Nil(t, text)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(nil, wantErr)

		s := NewTextsService(uowMock)
		text, err := s.GetText(ctx, 1, "name")

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, text)
	})
}

func TestGetTexts(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)
		want := []entities.Text{{Id: 1, Owner: 1, Name: "a"}, {Id: 2, Owner: 1, Name: "b"}}

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().GetTextsByOwner(ctx, int64(1)).Return(want, nil)

		s := NewTextsService(uowMock)
		texts, err := s.GetTexts(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, want, texts)
	})

	t.Run("error", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().GetTextsByOwner(ctx, int64(1)).Return(nil, wantErr)

		s := NewTextsService(uowMock)
		texts, err := s.GetTexts(ctx, 1)

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, texts)
	})
}

func TestCreateText(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)
		created := &entities.Text{Id: 1, Owner: 1, Name: "name", Text: "content", Metadata: "meta"}

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().
			CreateText(ctx, entities.Text{
				Owner:    1,
				Name:     "name",
				Text:     "content",
				Metadata: "meta",
				Checksum: checksum.Text("content", "meta"),
			}).
			Return(created, nil)

		s := NewTextsService(uowMock)
		text, err := s.CreateText(ctx, 1, "name", "content", "meta")

		assert.NoError(t, err)
		assert.Equal(t, created, text)
	})

	t.Run("already exists", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().
			CreateText(ctx, gomock.Any()).
			Return(nil, repository.ErrUniqueViolation)

		s := NewTextsService(uowMock)
		text, err := s.CreateText(ctx, 1, "name", "content", "meta")

		assert.ErrorIs(t, err, ErrTextAlreadyExists)
		assert.Nil(t, text)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().
			CreateText(ctx, gomock.Any()).
			Return(nil, wantErr)

		s := NewTextsService(uowMock)
		text, err := s.CreateText(ctx, 1, "name", "content", "meta")

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, text)
	})
}

func TestUpdateText(t *testing.T) {
	ctx := context.Background()
	createdAt := time.Now().Add(-time.Hour)

	t.Run("success", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)
		current := &entities.Text{Id: 1, Owner: 1, Name: "name", CreatedAt: createdAt}
		updatedAt := time.Now()

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(current, nil)

		expected := entities.Text{
			Id:        1,
			Owner:     1,
			Name:      "name",
			Text:      "newContent",
			Metadata:  "newMeta",
			Checksum:  checksum.Text("newContent", "newMeta"),
			CreatedAt: createdAt,
		}
		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().UpdateText(ctx, expected).Return(updatedAt, nil)

		s := NewTextsService(uowMock)
		text, err := s.UpdateText(ctx, 1, "name", "newContent", "newMeta")

		assert.NoError(t, err)
		assert.Equal(t, expected.Id, text.Id)
		assert.Equal(t, expected.Text, text.Text)
		assert.Equal(t, updatedAt, text.UpdatedAt)
	})

	t.Run("current not exist", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(nil, repository.ErrNoRows)

		s := NewTextsService(uowMock)
		text, err := s.UpdateText(ctx, 1, "name", "content", "meta")

		assert.ErrorIs(t, err, ErrTextNotExists)
		assert.Nil(t, text)
	})

	t.Run("update not exist", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)
		current := &entities.Text{Id: 1, Owner: 1, Name: "name", CreatedAt: createdAt}

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(current, nil)

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().UpdateText(ctx, gomock.Any()).Return(time.Time{}, repository.ErrNoRows)

		s := NewTextsService(uowMock)
		text, err := s.UpdateText(ctx, 1, "name", "content", "meta")

		assert.ErrorIs(t, err, ErrTextNotExists)
		assert.Nil(t, text)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)
		current := &entities.Text{Id: 1, Owner: 1, Name: "name", CreatedAt: createdAt}
		wantErr := errors.New("db error")

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(current, nil)

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().UpdateText(ctx, gomock.Any()).Return(time.Time{}, wantErr)

		s := NewTextsService(uowMock)
		text, err := s.UpdateText(ctx, 1, "name", "content", "meta")

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, text)
	})
}

func TestDeleteText(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().DeleteText(ctx, int64(1), "name").Return(nil)

		s := NewTextsService(uowMock)
		err := s.DeleteText(ctx, 1, "name")

		assert.NoError(t, err)
	})

	t.Run("not exist", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().DeleteText(ctx, int64(1), "name").Return(repository.ErrNoRows)

		s := NewTextsService(uowMock)
		err := s.DeleteText(ctx, 1, "name")

		assert.ErrorIs(t, err, ErrTextNotExists)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, textRepoMock := setupTextMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Texts().Return(textRepoMock)
		textRepoMock.EXPECT().DeleteText(ctx, int64(1), "name").Return(wantErr)

		s := NewTextsService(uowMock)
		err := s.DeleteText(ctx, 1, "name")

		assert.ErrorIs(t, err, wantErr)
	})
}
