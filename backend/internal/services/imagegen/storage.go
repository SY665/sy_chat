package imagegen

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxStoredImageSize = 20 << 20

// StoredImage 描述已经持久化到本地的图片。
type StoredImage struct {
	URL      string
	Filename string
}

// Storage 负责下载并持久化生成图片。
type Storage struct {
	directory string
	urlPrefix string
	client    *http.Client
}

func NewStorage(
	directory string,
	urlPrefix string,
	timeout time.Duration,
) (*Storage, error) {
	directory = strings.TrimSpace(directory)
	if directory == "" {
		return nil, fmt.Errorf("generated image directory is required")
	}

	urlPrefix = strings.TrimSpace(urlPrefix)
	if urlPrefix == "" {
		return nil, fmt.Errorf("generated image URL prefix is required")
	}

	cleanDirectory := filepath.Clean(directory)
	if err := os.MkdirAll(cleanDirectory, 0o755); err != nil {
		return nil, fmt.Errorf(
			"create generated image directory: %w",
			err,
		)
	}

	return &Storage{
		directory: cleanDirectory,
		urlPrefix: "/" + strings.Trim(urlPrefix, "/"),
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(
				request *http.Request,
				via []*http.Request,
			) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many image download redirects")
				}

				// 重定向后的地址仍然必须满足远程图片 URL 约束。
				_, err := validateRemoteImageURL(request.URL.String())
				return err
			},
		},
	}, nil
}

func (storage *Storage) Directory() string {
	return storage.directory
}

func (storage *Storage) URLPrefix() string {
	return storage.urlPrefix
}

func validateRemoteImageURL(value string) (string, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return "", fmt.Errorf("parse remote image URL: %w", err)
	}
	if parsedURL.Scheme != "https" || parsedURL.Host == "" {
		return "", fmt.Errorf("remote image URL must use HTTPS")
	}

	return parsedURL.String(), nil
}

func imageExtension(contentType string) (string, error) {
	contentType = strings.ToLower(strings.TrimSpace(contentType))

	switch contentType {
	case "image/png":
		return ".png", nil
	case "image/jpeg":
		return ".jpg", nil
	case "image/webp":
		return ".webp", nil
	default:
		return "", fmt.Errorf(
			"unsupported generated image type: %s",
			contentType,
		)
	}
}

// SaveRemote 下载远程图片，并在完整写入后移动到正式文件名。
func (storage *Storage) SaveRemote(
	ctx context.Context,
	value string,
) (StoredImage, error) {
	remoteURL, err := validateRemoteImageURL(value)
	if err != nil {
		return StoredImage{}, err
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		remoteURL,
		nil,
	)
	if err != nil {
		return StoredImage{}, fmt.Errorf(
			"create generated image download request: %w",
			err,
		)
	}

	response, err := storage.client.Do(request)
	if err != nil {
		return StoredImage{}, fmt.Errorf(
			"download generated image: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		return StoredImage{}, fmt.Errorf(
			"download generated image returned status %d",
			response.StatusCode,
		)
	}
	if response.ContentLength > maxStoredImageSize {
		return StoredImage{}, fmt.Errorf(
			"generated image exceeds size limit",
		)
	}

	data, err := io.ReadAll(io.LimitReader(
		response.Body,
		maxStoredImageSize+1,
	))
	if err != nil {
		return StoredImage{}, fmt.Errorf(
			"read generated image: %w",
			err,
		)
	}
	if len(data) > maxStoredImageSize {
		return StoredImage{}, fmt.Errorf(
			"generated image exceeds size limit",
		)
	}

	extension, err := imageExtension(http.DetectContentType(data))
	if err != nil {
		return StoredImage{}, err
	}

	filename := uuid.NewString() + extension
	targetPath := filepath.Join(storage.directory, filename)

	tempFile, err := os.CreateTemp(storage.directory, ".image-*")
	if err != nil {
		return StoredImage{}, fmt.Errorf(
			"create temporary image file: %w",
			err,
		)
	}

	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		return StoredImage{}, fmt.Errorf(
			"write temporary image file: %w",
			err,
		)
	}
	if err := tempFile.Close(); err != nil {
		return StoredImage{}, fmt.Errorf(
			"close temporary image file: %w",
			err,
		)
	}
	if err := os.Rename(tempPath, targetPath); err != nil {
		return StoredImage{}, fmt.Errorf(
			"store generated image: %w",
			err,
		)
	}

	return StoredImage{
		URL:      storage.urlPrefix + "/" + filename,
		Filename: filename,
	}, nil
}

// DeleteByURL 删除由当前存储实例生成的本地图片。
// 文件不存在时视为删除成功，方便清理流程安全重试。
func (storage *Storage) DeleteByURL(value string) error {
	parsedURL, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return fmt.Errorf("parse generated image URL: %w", err)
	}

	if parsedURL.IsAbs() ||
		parsedURL.Host != "" ||
		parsedURL.RawQuery != "" ||
		parsedURL.Fragment != "" {
		return fmt.Errorf("generated image URL must be a local path")
	}

	prefix := storage.urlPrefix + "/"
	if !strings.HasPrefix(parsedURL.Path, prefix) {
		return fmt.Errorf("generated image URL is outside storage prefix")
	}

	filename := strings.TrimPrefix(parsedURL.Path, prefix)
	extension := strings.ToLower(filepath.Ext(filename))

	switch extension {
	case ".png", ".jpg", ".webp":
	default:
		return fmt.Errorf("unsupported generated image extension")
	}

	imageID := strings.TrimSuffix(filename, extension)
	if _, err := uuid.Parse(imageID); err != nil {
		return fmt.Errorf("invalid generated image filename: %w", err)
	}

	targetPath := filepath.Join(storage.directory, filename)
	if err := os.Remove(targetPath); err != nil &&
		!errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete generated image: %w", err)
	}

	return nil
}
