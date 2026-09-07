package handlers

import (
	"errors"
	"io"
	"net/http"
	"path"
	"strings"
	"unicode/utf8"

	"sy_chat/internal/middleware"
	"sy_chat/internal/models"
	"sy_chat/internal/response"

	"github.com/gin-gonic/gin"
)

const (
	maxAttachmentSize        int64 = 1 << 20
	maxAttachmentRequestSize int64 = 2 << 20
)

// UploadAttachment godoc
// @Summary 上传文本附件
// @Description 读取不超过 1 MB 的 .txt 或 .md 文件。
// @Tags Attachments
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文本附件"
// @Success 200 {object} response.Envelope{data=models.FileAttachment}
// @Failure 400,401,413,500 {object} response.Envelope
// @Router /attachments [post]
func UploadAttachment(c *gin.Context) {
	if _, ok := middleware.CurrentUserID(c); !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	// 限制整个 multipart 请求，避免超大请求先被写入临时文件。
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxAttachmentRequestSize,
	)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			response.Error(
				c,
				http.StatusRequestEntityTooLarge,
				"FILE_TOO_LARGE",
				"文件不能超过 1 MB",
			)
			return
		}

		response.Error(
			c,
			http.StatusBadRequest,
			"FILE_REQUIRED",
			"请选择需要上传的文件",
		)
		return
	}

	if fileHeader.Size > maxAttachmentSize {
		response.Error(
			c,
			http.StatusRequestEntityTooLarge,
			"FILE_TOO_LARGE",
			"文件不能超过 1 MB",
		)
		return
	}

	// 同时处理 Unix 和 Windows 风格的客户端文件路径。
	name := path.Base(strings.ReplaceAll(fileHeader.Filename, "\\", "/"))
	extension := strings.ToLower(path.Ext(name))

	var attachmentType string
	switch extension {
	case ".txt":
		attachmentType = models.AttachmentTypeText
	case ".md":
		attachmentType = models.AttachmentTypeMarkdown
	default:
		response.Error(
			c,
			http.StatusBadRequest,
			"UNSUPPORTED_FILE_TYPE",
			"仅支持 .txt 和 .md 文件",
		)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			"FILE_OPEN_FAILED",
			"无法读取上传文件",
		)
		return
	}
	defer file.Close()

	// 即使客户端提供了错误的 size，也最多读取 1 MB 加一个检测字节。
	content, err := io.ReadAll(io.LimitReader(file, maxAttachmentSize+1))
	if err != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			"FILE_READ_FAILED",
			"读取上传文件失败",
		)
		return
	}

	if int64(len(content)) > maxAttachmentSize {
		response.Error(
			c,
			http.StatusRequestEntityTooLarge,
			"FILE_TOO_LARGE",
			"文件不能超过 1 MB",
		)
		return
	}

	if !utf8.Valid(content) {
		response.Error(
			c,
			http.StatusBadRequest,
			"INVALID_FILE_CONTENT",
			"文件必须使用 UTF-8 编码",
		)
		return
	}

	response.JSON(c, http.StatusOK, models.FileAttachment{
		Name:    name,
		Type:    attachmentType,
		Size:    int64(len(content)),
		Content: string(content),
	})
}
