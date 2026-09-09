package handlers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strings"
	"unicode/utf8"

	"sy_chat/internal/middleware"
	"sy_chat/internal/response"
	voiceservice "sy_chat/internal/services/voice"

	"github.com/gin-gonic/gin"
)

const (
	maxSpeechFileSize          int64 = 10 << 20
	maxSpeechRequestSize       int64 = 11 << 20
	maxTextToSpeechRequestSize int64 = 128 << 10
	maxSpeechTextLength              = 10000
	defaultSpeechVoice               = "FunAudioLLM/CosyVoice2-0.5B:diana"
)

var supportedAudioTypes = map[string]struct{}{
	"audio/webm":  {},
	"audio/mp4":   {},
	"audio/mpeg":  {},
	"audio/wav":   {},
	"audio/x-wav": {},
	"audio/ogg":   {},
}

var supportedSpeechVoices = map[string]struct{}{
	"FunAudioLLM/CosyVoice2-0.5B:diana":    {},
	"FunAudioLLM/CosyVoice2-0.5B:claire":   {},
	"FunAudioLLM/CosyVoice2-0.5B:anna":     {},
	"FunAudioLLM/CosyVoice2-0.5B:bella":    {},
	"FunAudioLLM/CosyVoice2-0.5B:alex":     {},
	"FunAudioLLM/CosyVoice2-0.5B:david":    {},
	"FunAudioLLM/CosyVoice2-0.5B:charles":  {},
	"FunAudioLLM/CosyVoice2-0.5B:benjamin": {},
}

type VoiceService interface {
	Transcribe(
		ctx context.Context,
		audio io.Reader,
		filename string,
	) (string, error)

	Synthesize(
		ctx context.Context,
		text string,
		voice string,
		speed float64,
	) (*voiceservice.SpeechAudio, error)
}

type VoiceHandler struct {
	service VoiceService
}

type SpeechData struct {
	Text string `json:"text"`
}

type TextToSpeechRequest struct {
	Text  string  `json:"text"`
	Voice string  `json:"voice"`
	Speed float64 `json:"speed"`
}

func NewVoiceHandler(service VoiceService) *VoiceHandler {
	return &VoiceHandler{
		service: service,
	}
}

// SpeechToText godoc
// @Summary 将录音转换为文字
// @Description 接收不超过 10 MB 的音频并返回识别文本。
// @Tags Voice
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "录音文件"
// @Success 200 {object} response.Envelope{data=SpeechData}
// @Failure 400,401,413,500,502,503 {object} response.Envelope
// @Router /speech [post]
func (handler *VoiceHandler) SpeechToText(c *gin.Context) {
	if _, ok := middleware.CurrentUserID(c); !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxSpeechRequestSize,
	)

	fileHeader, err := c.FormFile("audio")
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			response.Error(
				c,
				http.StatusRequestEntityTooLarge,
				"AUDIO_TOO_LARGE",
				"录音文件不能超过 10 MB",
			)
			return
		}

		response.Error(
			c,
			http.StatusBadRequest,
			"AUDIO_REQUIRED",
			"请提供录音文件",
		)
		return
	}

	if fileHeader.Size <= 0 {
		response.Error(c, http.StatusBadRequest, "EMPTY_AUDIO", "录音内容为空")
		return
	}

	if fileHeader.Size > maxSpeechFileSize {
		response.Error(
			c,
			http.StatusRequestEntityTooLarge,
			"AUDIO_TOO_LARGE",
			"录音文件不能超过 10 MB",
		)
		return
	}

	contentType, _, err := mime.ParseMediaType(
		fileHeader.Header.Get("Content-Type"),
	)
	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"INVALID_AUDIO_TYPE",
			"无法识别录音格式",
		)
		return
	}

	if _, supported := supportedAudioTypes[contentType]; !supported {
		response.Error(
			c,
			http.StatusBadRequest,
			"UNSUPPORTED_AUDIO_TYPE",
			"不支持该录音格式",
		)
		return
	}

	audioFile, err := fileHeader.Open()
	if err != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			"AUDIO_OPEN_FAILED",
			"无法读取录音文件",
		)
		return
	}
	defer audioFile.Close()

	filename := path.Base(
		strings.ReplaceAll(fileHeader.Filename, "\\", "/"),
	)

	text, err := handler.service.Transcribe(
		c.Request.Context(),
		audioFile,
		filename,
	)
	if err != nil {
		if c.Request.Context().Err() != nil {
			return
		}

		switch {
		case errors.Is(err, voiceservice.ErrUnavailable):
			response.Error(
				c,
				http.StatusServiceUnavailable,
				"VOICE_UNAVAILABLE",
				"语音服务尚未配置",
			)
		case errors.Is(err, voiceservice.ErrEmptyAudio):
			response.Error(
				c,
				http.StatusBadRequest,
				"EMPTY_AUDIO",
				"录音内容为空",
			)
		case errors.Is(err, voiceservice.ErrEmptyTranscript):
			response.Error(
				c,
				http.StatusBadGateway,
				"EMPTY_TRANSCRIPT",
				"没有识别到有效语音",
			)
		default:
			slog.Error("transcribe speech", "error", err)
			response.Error(
				c,
				http.StatusBadGateway,
				"SPEECH_TRANSCRIPTION_FAILED",
				"语音识别失败，请稍后重试",
			)
		}
		return
	}

	response.JSON(c, http.StatusOK, SpeechData{
		Text: text,
	})
}

// TextToSpeech godoc
// @Summary 将文本转换为语音
// @Description 使用指定的系统音色生成 MP3 音频。
// @Tags Voice
// @Accept json
// @Produce audio/mpeg
// @Param body body TextToSpeechRequest true "朗读参数"
// @Success 200 {file} binary
// @Failure 400,401,413,500,502,503 {object} response.Envelope
// @Router /tts [post]
func (handler *VoiceHandler) TextToSpeech(c *gin.Context) {
	if _, ok := middleware.CurrentUserID(c); !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxTextToSpeechRequestSize,
	)

	var request TextToSpeechRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		var maxBytesError *http.MaxBytesError

		if errors.As(err, &maxBytesError) {
			response.Error(
				c,
				http.StatusRequestEntityTooLarge,
				"TTS_REQUEST_TOO_LARGE",
				"朗读请求内容过大",
			)
			return
		}

		response.Error(
			c,
			http.StatusBadRequest,
			"INVALID_TTS_INPUT",
			"朗读参数格式不正确",
		)
		return
	}

	text := strings.TrimSpace(request.Text)
	if text == "" {
		response.Error(
			c,
			http.StatusBadRequest,
			"TEXT_REQUIRED",
			"朗读文本不能为空",
		)
		return
	}

	if utf8.RuneCountInString(text) > maxSpeechTextLength {
		response.Error(
			c,
			http.StatusBadRequest,
			"TEXT_TOO_LONG",
			"单次朗读内容不能超过 10000 个字符",
		)
		return
	}

	voice := strings.TrimSpace(request.Voice)
	if voice == "" {
		voice = defaultSpeechVoice
	}

	if _, supported := supportedSpeechVoices[voice]; !supported {
		response.Error(
			c,
			http.StatusBadRequest,
			"UNSUPPORTED_VOICE",
			"不支持所选音色",
		)
		return
	}

	speed := request.Speed
	if speed == 0 {
		speed = 1
	}

	if speed < 0.25 || speed > 4 {
		response.Error(
			c,
			http.StatusBadRequest,
			"INVALID_SPEECH_SPEED",
			"语速必须在 0.25 到 4.0 之间",
		)
		return
	}

	audio, err := handler.service.Synthesize(
		c.Request.Context(),
		text,
		voice,
		speed,
	)
	if err != nil {
		if c.Request.Context().Err() != nil {
			return
		}

		if errors.Is(err, voiceservice.ErrUnavailable) {
			response.Error(
				c,
				http.StatusServiceUnavailable,
				"VOICE_UNAVAILABLE",
				"语音服务尚未配置",
			)
			return
		}

		slog.Error("synthesize speech", "error", err)
		response.Error(
			c,
			http.StatusBadGateway,
			"SPEECH_SYNTHESIS_FAILED",
			"语音生成失败，请稍后重试",
		)
		return
	}
	defer audio.Body.Close()

	c.Header("Content-Type", audio.ContentType)
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Status(http.StatusOK)

	if _, err := io.Copy(c.Writer, audio.Body); err != nil {
		slog.Warn("stream speech response", "error", err)
	}
}
