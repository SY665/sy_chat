package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// GeneratedImageCleaner 定义生成图片资源需要提供的最小清理能力。
type GeneratedImageCleaner interface {
	DeleteByURL(value string) error
}

// ParseGeneratedImageURLs 从持久化的工具结果中提取不重复的图片地址。
func ParseGeneratedImageURLs(data []byte) ([]string, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var results []Result
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("decode tool results: %w", err)
	}

	urls := make([]string, 0)
	seen := make(map[string]struct{})

	for _, result := range results {
		if result.Image == nil {
			continue
		}

		imageURL := strings.TrimSpace(result.Image.URL)
		if imageURL == "" {
			continue
		}
		if _, exists := seen[imageURL]; exists {
			continue
		}

		seen[imageURL] = struct{}{}
		urls = append(urls, imageURL)
	}

	return urls, nil
}

// DeleteGeneratedImages 解析一条消息的工具结果，并尽量删除其中全部图片。
func DeleteGeneratedImages(
	cleaner GeneratedImageCleaner,
	data []byte,
) error {
	if cleaner == nil {
		return nil
	}

	urls, err := ParseGeneratedImageURLs(data)
	if err != nil {
		return err
	}

	var cleanupErrors []error

	for _, imageURL := range urls {
		if err := cleaner.DeleteByURL(imageURL); err != nil {
			cleanupErrors = append(
				cleanupErrors,
				fmt.Errorf(
					"delete generated image %q: %w",
					imageURL,
					err,
				),
			)
		}
	}

	return errors.Join(cleanupErrors...)
}
