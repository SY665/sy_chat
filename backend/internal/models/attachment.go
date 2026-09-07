package models

import (
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
)

const (
	AttachmentTypeText     = "txt"
	AttachmentTypeMarkdown = "md"
)

// FileAttachment 描述附加到用户消息中的文本文件。
// Content 会作为模型上下文使用，并随消息一起保存到 JSON 字段。
type FileAttachment struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Size    int64  `json:"size"`
	Content string `json:"content"`
}

// EncodeFileAttachments 将附件数组转换为 Message.Attachments 使用的 JSON。
func EncodeFileAttachments(
	attachments []FileAttachment,
) (datatypes.JSON, error) {
	if len(attachments) == 0 {
		return nil, nil
	}

	data, err := json.Marshal(attachments)
	if err != nil {
		return nil, fmt.Errorf("encode file attachments: %w", err)
	}

	return datatypes.JSON(data), nil
}

// DecodeFileAttachments 将数据库 JSON 恢复为附件数组。
func DecodeFileAttachments(
	data datatypes.JSON,
) ([]FileAttachment, error) {
	if len(data) == 0 {
		return []FileAttachment{}, nil
	}

	var attachments []FileAttachment
	if err := json.Unmarshal(data, &attachments); err != nil {
		return nil, fmt.Errorf("decode file attachments: %w", err)
	}

	if attachments == nil {
		return []FileAttachment{}, nil
	}

	return attachments, nil
}
