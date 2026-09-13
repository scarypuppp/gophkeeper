package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/filestorage/mocks"
	"github.com/scarypuppp/gophkeeper/internal/server/repository"
	repoMocks "github.com/scarypuppp/gophkeeper/internal/server/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupFileMocks(t *testing.T) (*repoMocks.MockUnitOfWork, *repoMocks.MockFileRepository, *mocks.MockFileStorage) {
	ctrl := gomock.NewController(t)
	uowMock := repoMocks.NewMockUnitOfWork(ctrl)
	fileRepoMock := repoMocks.NewMockFileRepository(ctrl)
	storageMock := mocks.NewMockFileStorage(ctrl)
	return uowMock, fileRepoMock, storageMock
}

func TestGetFileByName(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		existing := &entities.File{Id: 1, Owner: 1, FileHash: "hash"}

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByName(ctx, int64(1), "hash").Return(existing, nil)

		s := NewFileService(uowMock, storageMock)
		file, err := s.GetFileByName(ctx, 1, "hash")

		assert.NoError(t, err)
		assert.Equal(t, existing, file)
	})

	t.Run("not exist", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByName(ctx, int64(1), "hash").Return(nil, repository.ErrNoRows)

		s := NewFileService(uowMock, storageMock)
		file, err := s.GetFileByName(ctx, 1, "hash")

		assert.ErrorIs(t, err, ErrFileNotExists)
		assert.Nil(t, file)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByName(ctx, int64(1), "hash").Return(nil, wantErr)

		s := NewFileService(uowMock, storageMock)
		file, err := s.GetFileByName(ctx, 1, "hash")

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, file)
	})
}

func TestGetFile(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		existing := &entities.File{Id: 1, Owner: 1, FileHash: "hash", FileName: "name"}

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(existing, nil)

		s := NewFileService(uowMock, storageMock)
		file, err := s.GetFile(ctx, 1, "hash", "name")

		assert.NoError(t, err)
		assert.Equal(t, existing, file)
	})

	t.Run("not exist", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(nil, repository.ErrNoRows)

		s := NewFileService(uowMock, storageMock)
		file, err := s.GetFile(ctx, 1, "hash", "name")

		assert.ErrorIs(t, err, ErrFileNotExists)
		assert.Nil(t, file)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(nil, wantErr)

		s := NewFileService(uowMock, storageMock)
		file, err := s.GetFile(ctx, 1, "hash", "name")

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, file)
	})
}

func TestGetFileContent(t *testing.T) {
	ctx := context.Background()
	uowMock, _, storageMock := setupFileMocks(t)
	content := io.NopCloser(strings.NewReader("data"))

	storageMock.EXPECT().GetFile(ctx, "path").Return(content, nil)

	s := NewFileService(uowMock, storageMock)
	got, err := s.GetFileContent(ctx, "path")

	assert.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestGetFiles(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		want := []entities.File{{Id: 1, Owner: 1}, {Id: 2, Owner: 1}}

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetFilesByOwner(ctx, int64(1)).Return(want, nil)

		s := NewFileService(uowMock, storageMock)
		files, err := s.GetFiles(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, want, files)
	})

	t.Run("not exist", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetFilesByOwner(ctx, int64(1)).Return(nil, repository.ErrNoRows)

		s := NewFileService(uowMock, storageMock)
		files, err := s.GetFiles(ctx, 1)

		assert.ErrorIs(t, err, ErrFileNotExists)
		assert.Nil(t, files)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetFilesByOwner(ctx, int64(1)).Return(nil, wantErr)

		s := NewFileService(uowMock, storageMock)
		files, err := s.GetFiles(ctx, 1)

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, files)
	})
}

// fileEntityMatcher проверяет поля entities.File при создании файла,
// не завязываясь на случайно сгенерированный FileHash.
type fileEntityMatcher struct {
	owner    int64
	fileName string
	metadata string
	checksum string
	size     int64
}

func (m fileEntityMatcher) Matches(x any) bool {
	f, ok := x.(entities.File)
	if !ok {
		return false
	}
	wantPath := fmt.Sprintf("%d/%s/%s", m.owner, f.FileHash, strings.ReplaceAll(m.fileName, " ", "_"))
	return f.Owner == m.owner &&
		f.FileName == m.fileName &&
		f.Metadata == m.metadata &&
		f.Checksum == m.checksum &&
		f.Size == m.size &&
		f.FileHash != "" &&
		f.StorageFilePath == wantPath
}

func (m fileEntityMatcher) String() string {
	return fmt.Sprintf("matches file entity for owner=%d name=%q", m.owner, m.fileName)
}

func TestCreateFile(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		content := strings.NewReader("data")
		created := &entities.File{Id: 1, Owner: 1, FileName: "my file.txt"}

		storageMock.EXPECT().
			PutFile(ctx, gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, path string, _ io.Reader) (string, error) {
				assert.True(t, strings.HasPrefix(path, "1/"))
				assert.True(t, strings.HasSuffix(path, "/my_file.txt"))
				return "checksum123", nil
			})
		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().
			CreateFile(ctx, fileEntityMatcher{owner: 1, fileName: "my file.txt", metadata: "meta", checksum: "checksum123", size: 4}).
			Return(created, nil)

		s := NewFileService(uowMock, storageMock)
		file, err := s.CreateFile(ctx, 1, "my file.txt", "meta", 4, content)

		assert.NoError(t, err)
		assert.Equal(t, created, file)
	})

	t.Run("upload error", func(t *testing.T) {
		uowMock, _, storageMock := setupFileMocks(t)
		content := strings.NewReader("data")
		wantErr := errors.New("upload error")

		storageMock.EXPECT().PutFile(ctx, gomock.Any(), gomock.Any()).Return("", wantErr)

		s := NewFileService(uowMock, storageMock)
		file, err := s.CreateFile(ctx, 1, "file.txt", "meta", 4, content)

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, file)
	})

	t.Run("already exists", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		content := strings.NewReader("data")

		storageMock.EXPECT().PutFile(ctx, gomock.Any(), gomock.Any()).Return("checksum123", nil)
		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().CreateFile(ctx, gomock.Any()).Return(nil, repository.ErrUniqueViolation)
		storageMock.EXPECT().DeleteFile(ctx, gomock.Any()).Return(nil)

		s := NewFileService(uowMock, storageMock)
		file, err := s.CreateFile(ctx, 1, "file.txt", "meta", 4, content)

		assert.ErrorIs(t, err, ErrFileAlreadyExists)
		assert.Nil(t, file)
	})

	t.Run("unexpected create error", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		content := strings.NewReader("data")
		wantErr := errors.New("db error")

		storageMock.EXPECT().PutFile(ctx, gomock.Any(), gomock.Any()).Return("checksum123", nil)
		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().CreateFile(ctx, gomock.Any()).Return(nil, wantErr)
		storageMock.EXPECT().DeleteFile(ctx, gomock.Any()).Return(nil)

		s := NewFileService(uowMock, storageMock)
		file, err := s.CreateFile(ctx, 1, "file.txt", "meta", 4, content)

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, file)
	})
}

func TestUpdateFileMetadata(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		existing := &entities.File{Id: 1, Owner: 1, FileHash: "hash", FileName: "name", Metadata: "old"}
		updatedAt := time.Now()

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(existing, nil)
		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().UpdateFileMetadata(ctx, int64(1), "hash", "name", "new").Return(updatedAt, nil)

		s := NewFileService(uowMock, storageMock)
		file, err := s.UpdateFileMetadata(ctx, 1, "hash", "name", "new")

		assert.NoError(t, err)
		assert.Equal(t, "new", file.Metadata)
		assert.Equal(t, updatedAt, file.UpdatedAt)
	})

	t.Run("current not exist", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(nil, repository.ErrNoRows)

		s := NewFileService(uowMock, storageMock)
		file, err := s.UpdateFileMetadata(ctx, 1, "hash", "name", "new")

		assert.ErrorIs(t, err, ErrFileNotExists)
		assert.Nil(t, file)
	})

	t.Run("update not exist", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		existing := &entities.File{Id: 1, Owner: 1, FileHash: "hash", FileName: "name"}

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(existing, nil)
		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().UpdateFileMetadata(ctx, int64(1), "hash", "name", "new").Return(time.Time{}, repository.ErrNoRows)

		s := NewFileService(uowMock, storageMock)
		file, err := s.UpdateFileMetadata(ctx, 1, "hash", "name", "new")

		assert.ErrorIs(t, err, ErrFileNotExists)
		assert.Nil(t, file)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		existing := &entities.File{Id: 1, Owner: 1, FileHash: "hash", FileName: "name"}
		wantErr := errors.New("db error")

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(existing, nil)
		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().UpdateFileMetadata(ctx, int64(1), "hash", "name", "new").Return(time.Time{}, wantErr)

		s := NewFileService(uowMock, storageMock)
		file, err := s.UpdateFileMetadata(ctx, 1, "hash", "name", "new")

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, file)
	})
}

func TestUpdateFileContent(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		existing := &entities.File{Id: 1, Owner: 1, FileHash: "hash", FileName: "name", StorageFilePath: "1/hash/name"}
		content := strings.NewReader("newdata")
		updatedAt := time.Now()

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(existing, nil)
		storageMock.EXPECT().PutFile(ctx, "1/hash/name", content).Return("newchecksum", nil)
		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().UpdateFileContent(ctx, int64(1), "hash", "name", "newchecksum", int64(7)).Return(updatedAt, nil)

		s := NewFileService(uowMock, storageMock)
		file, err := s.UpdateFileContent(ctx, 1, "hash", "name", 7, content)

		assert.NoError(t, err)
		assert.Equal(t, "newchecksum", file.Checksum)
		assert.Equal(t, int64(7), file.Size)
		assert.Equal(t, updatedAt, file.UpdatedAt)
	})

	t.Run("current not exist", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		content := strings.NewReader("newdata")

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(nil, repository.ErrNoRows)

		s := NewFileService(uowMock, storageMock)
		file, err := s.UpdateFileContent(ctx, 1, "hash", "name", 7, content)

		assert.ErrorIs(t, err, ErrFileNotExists)
		assert.Nil(t, file)
	})

	t.Run("upload error", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		existing := &entities.File{Id: 1, Owner: 1, FileHash: "hash", FileName: "name", StorageFilePath: "1/hash/name"}
		content := strings.NewReader("newdata")
		wantErr := errors.New("upload error")

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(existing, nil)
		storageMock.EXPECT().PutFile(ctx, "1/hash/name", content).Return("", wantErr)

		s := NewFileService(uowMock, storageMock)
		file, err := s.UpdateFileContent(ctx, 1, "hash", "name", 7, content)

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, file)
	})

	t.Run("update not exist", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		existing := &entities.File{Id: 1, Owner: 1, FileHash: "hash", FileName: "name", StorageFilePath: "1/hash/name"}
		content := strings.NewReader("newdata")

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(existing, nil)
		storageMock.EXPECT().PutFile(ctx, "1/hash/name", content).Return("newchecksum", nil)
		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().UpdateFileContent(ctx, int64(1), "hash", "name", "newchecksum", int64(7)).Return(time.Time{}, repository.ErrNoRows)

		s := NewFileService(uowMock, storageMock)
		file, err := s.UpdateFileContent(ctx, 1, "hash", "name", 7, content)

		assert.ErrorIs(t, err, ErrFileNotExists)
		assert.Nil(t, file)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		existing := &entities.File{Id: 1, Owner: 1, FileHash: "hash", FileName: "name", StorageFilePath: "1/hash/name"}
		content := strings.NewReader("newdata")
		wantErr := errors.New("db error")

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(existing, nil)
		storageMock.EXPECT().PutFile(ctx, "1/hash/name", content).Return("newchecksum", nil)
		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().UpdateFileContent(ctx, int64(1), "hash", "name", "newchecksum", int64(7)).Return(time.Time{}, wantErr)

		s := NewFileService(uowMock, storageMock)
		file, err := s.UpdateFileContent(ctx, 1, "hash", "name", 7, content)

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, file)
	})
}

func TestDeleteFile(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		existing := &entities.File{Id: 1, Owner: 1, FileHash: "hash", FileName: "name", StorageFilePath: "1/hash/name"}

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(existing, nil)
		storageMock.EXPECT().DeleteFile(ctx, "1/hash/name").Return(nil)
		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().DeleteFile(ctx, int64(1), "hash", "name").Return(nil)

		s := NewFileService(uowMock, storageMock)
		err := s.DeleteFile(ctx, 1, "hash", "name")

		assert.NoError(t, err)
	})

	t.Run("current not exist", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(nil, repository.ErrNoRows)

		s := NewFileService(uowMock, storageMock)
		err := s.DeleteFile(ctx, 1, "hash", "name")

		assert.ErrorIs(t, err, ErrFileNotExists)
	})

	t.Run("storage delete error", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		existing := &entities.File{Id: 1, Owner: 1, FileHash: "hash", FileName: "name", StorageFilePath: "1/hash/name"}
		wantErr := errors.New("storage error")

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(existing, nil)
		storageMock.EXPECT().DeleteFile(ctx, "1/hash/name").Return(wantErr)

		s := NewFileService(uowMock, storageMock)
		err := s.DeleteFile(ctx, 1, "hash", "name")

		assert.ErrorIs(t, err, wantErr)
	})

	t.Run("delete not exist", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		existing := &entities.File{Id: 1, Owner: 1, FileHash: "hash", FileName: "name", StorageFilePath: "1/hash/name"}

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(existing, nil)
		storageMock.EXPECT().DeleteFile(ctx, "1/hash/name").Return(nil)
		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().DeleteFile(ctx, int64(1), "hash", "name").Return(repository.ErrNoRows)

		s := NewFileService(uowMock, storageMock)
		err := s.DeleteFile(ctx, 1, "hash", "name")

		assert.ErrorIs(t, err, ErrFileNotExists)
	})

	t.Run("unexpected delete error", func(t *testing.T) {
		uowMock, fileRepoMock, storageMock := setupFileMocks(t)
		existing := &entities.File{Id: 1, Owner: 1, FileHash: "hash", FileName: "name", StorageFilePath: "1/hash/name"}
		wantErr := errors.New("db error")

		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().GetByHashAndName(ctx, int64(1), "hash", "name").Return(existing, nil)
		storageMock.EXPECT().DeleteFile(ctx, "1/hash/name").Return(nil)
		uowMock.EXPECT().Files().Return(fileRepoMock)
		fileRepoMock.EXPECT().DeleteFile(ctx, int64(1), "hash", "name").Return(wantErr)

		s := NewFileService(uowMock, storageMock)
		err := s.DeleteFile(ctx, 1, "hash", "name")

		assert.ErrorIs(t, err, wantErr)
	})
}
