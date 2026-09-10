package interactor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/scarypuppp/gophkeeper/internal/api"
)

// ErrFileNotFound возвращается, если сервер ответил 404 на операцию с файлом.
var ErrFileNotFound = errors.New("file not found")

// ErrFileTooLarge возвращается, если файл больше лимита сервера.
var ErrFileTooLarge = errors.New("file is too large")

func (i *Interactor) UploadFile(token string, filePath string, metadata string) (*api.CreateFileResponse, error) {
	if err := checkUploadSize(filePath); err != nil {
		return nil, err
	}

	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetFile("file", filePath).
		SetFormData(map[string]string{"metadata": metadata}).
		Post("/api/file/upload")
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.StatusCode() == http.StatusRequestEntityTooLarge {
		return nil, fmt.Errorf("%w: server rejected the upload: %s", ErrFileTooLarge, strings.TrimSpace(resp.String()))
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}

	var createFileResponse api.CreateFileResponse
	if err = json.Unmarshal(resp.Body(), &createFileResponse); err != nil {
		return nil, err
	}

	return &createFileResponse, nil
}

func (i *Interactor) GetFiles(token string) (api.GetFilesResponse, error) {
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		Get("/api/file")
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}

	var getFilesResponse api.GetFilesResponse
	if err = json.Unmarshal(resp.Body(), &getFilesResponse); err != nil {
		return nil, err
	}

	return getFilesResponse, nil
}

// DownloadFile скачивает файл по хэшу и имени и сохраняет его содержимое в dst.
func (i *Interactor) DownloadFile(token string, fileHash string, fileName string, dst io.Writer) error {
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetDoNotParseResponse(true).
		Get(filePath(fileHash, fileName))
	if err != nil {
		return i.requestError(err)
	}
	body := resp.RawBody()
	defer body.Close()

	if resp.IsError() {
		message, _ := io.ReadAll(io.LimitReader(body, 4096))
		return fmt.Errorf("server returned %d: %s", resp.StatusCode(), strings.TrimSpace(string(message)))
	}

	if _, err = io.Copy(dst, body); err != nil {
		return fmt.Errorf("error saving file: %w", err)
	}
	return nil
}

// filePath строит путь к файлу, экранируя хэш и имя: имя может содержать '/' и пробелы.
func filePath(fileHash string, fileName string) string {
	return fmt.Sprintf("/api/file/%s/%s", url.PathEscape(fileHash), url.PathEscape(fileName))
}

// UpdateFileMetadata меняет метаданные файла, не трогая его содержимое.
func (i *Interactor) UpdateFileMetadata(token string, fileHash string, fileName string, metadata string) (*api.FileResponse, error) {
	var fileResponse api.FileResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetBody(api.UpdateFileRequest{Metadata: metadata}).
		SetResult(&fileResponse).
		Put(filePath(fileHash, fileName) + "/metadata")
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrFileNotFound
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return &fileResponse, nil
}

// UpdateFileContent перезаписывает содержимое файла на сервере.
func (i *Interactor) UpdateFileContent(token string, fileHash string, fileName string, localPath string) (*api.FileResponse, error) {
	if err := checkUploadSize(localPath); err != nil {
		return nil, err
	}

	var fileResponse api.FileResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetFile("file", localPath).
		SetResult(&fileResponse).
		Put(filePath(fileHash, fileName))
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrFileNotFound
	}
	if resp.StatusCode() == http.StatusRequestEntityTooLarge {
		return nil, fmt.Errorf("%w: server rejected the upload: %s", ErrFileTooLarge, strings.TrimSpace(resp.String()))
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return &fileResponse, nil
}

// DeleteFile удаляет файл на сервере.
func (i *Interactor) DeleteFile(token string, fileHash string, fileName string) error {
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		Delete(filePath(fileHash, fileName))
	if err != nil {
		return i.requestError(err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return ErrFileNotFound
	}
	if resp.IsError() {
		return serverError(resp)
	}
	return nil
}

// checkUploadSize проверяет, что файл не превышает api.MaxFileUploadSize.
func checkUploadSize(localPath string) error {
	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", localPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory, not a file", localPath)
	}
	if info.Size() > api.MaxFileUploadSize {
		return fmt.Errorf("%w: %d bytes, limit is %d bytes", ErrFileTooLarge, info.Size(), int64(api.MaxFileUploadSize))
	}
	return nil
}
