package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/scarypuppp/gophkeeper/internal/checksum"
	"github.com/scarypuppp/gophkeeper/internal/client/app"
	"github.com/scarypuppp/gophkeeper/internal/client/storage"
	"github.com/spf13/cobra"
)

// FileCommand объединяет команды работы с бинарными данными (файлами).
// Идентификатором файла является его имя и передаётся первым аргументом.
func FileCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "file",
		Short: "Manage files stored in gophkeeper.",
		Long: `Содержимое файлов лежит в каталоге downloads из конфигурации,
метаданные — в data.json. На сервер файлы уходят при синхронизации.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	c.AddCommand(addFileCommand(a))
	c.AddCommand(listFilesCommand(a))
	c.AddCommand(updateFileCommand(a))
	c.AddCommand(downloadFileCommand(a))
	c.AddCommand(deleteFileCommand(a))

	return c
}

func addFileCommand(a *app.App) *cobra.Command {
	var (
		filePath string
		metadata string
	)

	c := &cobra.Command{
		Use:   "add",
		Short: "Add file to gophkeeper.",
		Long: `Файл копируется в каталог downloads, запись создаётся локально.
На сервер содержимое уходит при синхронизации. Если запись с таким именем
уже есть, содержимое перезаписывается с предупреждением.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			fileName := filepath.Base(filePath)
			_, err := a.Storage.GetFile(fileName)
			exists := err == nil
			if err != nil && !errors.Is(err, storage.ErrNotFound) {
				return err
			}
			if exists {
				fmt.Printf("Warning: file '%s' is already stored, its content will be overwritten\n", fileName)
			}

			if err := copyToDownloads(a, filePath, fileName); err != nil {
				return err
			}
			size, sum, err := statAndHash(downloadPath(a, fileName))
			if err != nil {
				return err
			}

			var file storage.File
			if exists {
				patch := storage.FilePatch{Size: &size, Checksum: &sum}
				if cmd.Flags().Changed("meta") {
					patch.Metadata = &metadata
				}
				file, err = a.Storage.UpdateFile(fileName, patch)
			} else {
				file, err = a.Storage.AddFile(storage.File{
					FileName: fileName,
					Size:     size,
					Checksum: sum,
					Metadata: metadata,
				})
			}
			if err != nil {
				return err
			}

			// Время файла равно updated_at записи, иначе только что добавленный
			// файл выглядел бы изменённым локально.
			if err := touch(downloadPath(a, fileName), file.UpdatedAt); err != nil {
				return err
			}
			fmt.Printf("Successfully added file '%s' (%d bytes)\n", file.FileName, file.Size)
			return nil
		},
	}

	c.Flags().StringVarP(&filePath, "file", "f", "", "Path to the file to add")
	c.Flags().StringVarP(&metadata, "meta", "m", "", "Arbitrary metadata")
	_ = c.MarkFlagRequired("file")

	return c
}

func listFilesCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "List stored files.",
		Long: `Колонка stored показывает, лежит ли содержимое файла на локальном диске.
В data.json она не хранится и вычисляется по факту наличия файла в каталоге downloads.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			files := a.Storage.Files()
			if len(files) == 0 {
				fmt.Println("No files stored yet")
				return nil
			}

			table := newTable()
			fmt.Fprintln(table, "filename\tsize\tstored\tmeta")
			for _, file := range files {
				stored, err := fileExists(downloadPath(a, file.FileName))
				if err != nil {
					return err
				}
				fmt.Fprintf(table, "%s\t%d\t%s\t%s\n",
					file.FileName,
					file.Size,
					storedMark(stored),
					truncate(file.Metadata, metaPreviewLen),
				)
			}
			return table.Flush()
		},
	}

	return c
}

func updateFileCommand(a *app.App) *cobra.Command {
	var metadata string

	c := &cobra.Command{
		Use:   "update <filename>",
		Short: "Update file metadata.",
		Long: `Команда меняет только метаданные. Содержимое считается изменённым,
если пользователь отредактировал файл в каталоге downloads; проверка идёт
по ступеням, каждая следующая — только если предыдущая не дала ответа:

  1. size отличается от сохранённого — файл изменён;
  2. время изменения файла отличается от updated_at записи — переходим к шагу 3;
  3. checksum (SHA-256) отличается от сохранённого — файл изменён.

Полный хэш считается только на третьей ступени, поэтому неизменённые файлы
не читаются с диска.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}
			fileName := args[0]

			file, err := a.Storage.GetFile(fileName)
			if err != nil {
				return err
			}

			var patch storage.FilePatch
			if cmd.Flags().Changed("meta") {
				patch.Metadata = &metadata
			}

			size, sum, changed, err := contentChanged(a, file)
			if err != nil {
				return err
			}
			if changed {
				fmt.Printf("Local content of '%s' has changed, it will be uploaded on the next sync\n", fileName)
				patch.Size = &size
				patch.Checksum = &sum
			}

			if patch.Metadata == nil && !changed {
				fmt.Printf("File '%s' is up to date\n", fileName)
				return nil
			}

			updated, err := a.Storage.UpdateFile(fileName, patch)
			if err != nil {
				return err
			}
			if changed {
				if err := touch(downloadPath(a, fileName), updated.UpdatedAt); err != nil {
					return err
				}
			}
			fmt.Printf("Successfully updated file '%s'\n", updated.FileName)
			return nil
		},
	}

	c.Flags().StringVarP(&metadata, "meta", "m", "", "New metadata")

	return c
}

func downloadFileCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "download <filename>",
		Short: "Download file content from the server.",
		Long: `Файл сохраняется в каталог downloads из конфигурации, путь не настраивается.
После скачивания времени изменения файла присваивается значение updated_at
записи, иначе только что скачанный файл выглядел бы изменённым локально.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			file, err := a.Storage.GetFile(args[0])
			if err != nil {
				return err
			}
			if file.FileHash == "" {
				return fmt.Errorf("file %q is not uploaded to the server yet, run 'gophkeeper sync' first", file.FileName)
			}

			path := downloadPath(a, file.FileName)
			stored, err := contentMatches(path, file)
			if err != nil {
				return err
			}
			if stored {
				fmt.Printf("File '%s' is already downloaded to '%s'\n", file.FileName, path)
				return nil
			}

			if err := downloadContent(a, file, path); err != nil {
				return err
			}
			// Время файла равно updated_at записи, иначе только что скачанный
			// файл выглядел бы изменённым локально.
			if err := touch(path, file.UpdatedAt); err != nil {
				return err
			}
			fmt.Printf("Successfully downloaded '%s' to '%s'\n", file.FileName, path)
			return nil
		},
	}

	return c
}

func deleteFileCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "delete <filename>",
		Short: "Delete file.",
		Long: `Удаляет файл с локальной машины вместе с записью в data.json.
На сервере файл удаляется при следующей синхронизации.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}
			fileName := args[0]

			if err := a.Storage.DeleteFile(fileName); err != nil {
				return err
			}
			if err := os.Remove(downloadPath(a, fileName)); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("remove local copy: %w", err)
			}
			fmt.Printf("Successfully deleted file '%s'\n", fileName)
			return nil
		},
	}

	return c
}

// downloadPath возвращает путь к локальной копии файла.
func downloadPath(a *app.App, fileName string) string {
	return filepath.Join(a.Config.DownloadedFilesPath, fileName)
}

// fileExists сообщает, лежит ли файл на диске.
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// copyToDownloads копирует файл в каталог downloads под именем fileName.
// Файл, который уже лежит в каталоге, не копируется сам в себя.
func copyToDownloads(a *app.App, filePath string, fileName string) error {
	dstPath := downloadPath(a, fileName)
	same, err := samePath(filePath, dstPath)
	if err != nil {
		return err
	}
	if same {
		return nil
	}

	src, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open %s: %w", filePath, err)
	}
	defer src.Close()

	if err := os.MkdirAll(a.Config.DownloadedFilesPath, 0755); err != nil {
		return fmt.Errorf("create %s: %w", a.Config.DownloadedFilesPath, err)
	}
	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create %s: %w", dstPath, err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return fmt.Errorf("copy %s: %w", dstPath, err)
	}
	return dst.Close()
}

// samePath сообщает, указывают ли пути на один и тот же файл.
func samePath(a string, b string) (bool, error) {
	absA, err := filepath.Abs(a)
	if err != nil {
		return false, err
	}
	absB, err := filepath.Abs(b)
	if err != nil {
		return false, err
	}
	return absA == absB, nil
}

// statAndHash возвращает размер и контрольную сумму содержимого файла.
func statAndHash(path string) (int64, string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, "", err
	}
	f, err := os.Open(path)
	if err != nil {
		return 0, "", err
	}
	defer f.Close()

	sum, err := checksum.Reader(f)
	if err != nil {
		return 0, "", fmt.Errorf("hash %s: %w", path, err)
	}
	return info.Size(), sum, nil
}

// touch выставляет файлу время изменения, равное updatedAt записи.
func touch(path string, updatedAt time.Time) error {
	if err := os.Chtimes(path, updatedAt, updatedAt); err != nil {
		return fmt.Errorf("set mtime of %s: %w", path, err)
	}
	return nil
}

// contentChanged проверяет, изменилось ли содержимое локальной копии файла,
// и возвращает её размер и контрольную сумму. Проверка идёт по ступеням:
// размер, время изменения, и только потом — полный хэш, поэтому неизменённые
// файлы с диска не читаются.
func contentChanged(a *app.App, file storage.File) (int64, string, bool, error) {
	path := downloadPath(a, file.FileName)
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		// Файл не скачан на эту машину, сравнивать не с чем.
		return 0, "", false, nil
	}
	if err != nil {
		return 0, "", false, err
	}

	if info.Size() != file.Size {
		size, sum, err := statAndHash(path)
		return size, sum, true, err
	}
	if info.ModTime().Equal(file.UpdatedAt) {
		return 0, "", false, nil
	}

	size, sum, err := statAndHash(path)
	if err != nil {
		return 0, "", false, err
	}
	return size, sum, sum != file.Checksum, nil
}

// contentMatches сообщает, лежит ли на диске содержимое, отвечающее записи.
func contentMatches(path string, file storage.File) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.Size() != file.Size {
		return false, nil
	}
	_, sum, err := statAndHash(path)
	if err != nil {
		return false, err
	}
	return sum == file.Checksum, nil
}

// downloadContent скачивает содержимое файла в path, показывая прогресс.
func downloadContent(a *app.App, file storage.File, path string) error {
	if err := os.MkdirAll(a.Config.DownloadedFilesPath, 0755); err != nil {
		return fmt.Errorf("create %s: %w", a.Config.DownloadedFilesPath, err)
	}
	dst, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}

	bar := newProgressBar(os.Stdout, file.Size)
	err = a.Interactor.DownloadFile(a.Session.AuthToken, file.FileHash, file.FileName, io.MultiWriter(dst, bar))
	bar.done()

	if closeErr := dst.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(path)
		return err
	}
	return nil
}
