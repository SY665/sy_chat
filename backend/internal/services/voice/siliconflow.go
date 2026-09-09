package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

var (
	ErrUnavailable     = errors.New("voice service is unavailable")
	ErrEmptyAudio      = errors.New("audio content is empty")
	ErrEmptyTranscript = errors.New("transcription result is empty")
)

type SiliconFlowClient struct {
	apiKey     string
	baseURL    string
	sttModel   string
	ttsModel   string
	httpClient *http.Client
}

func NewSiliconFlowClient(
	apiKey string,
	baseURL string,
	sttModel string,
	ttsModel string,
	timeout time.Duration,
) *SiliconFlowClient {
	return &SiliconFlowClient{
		apiKey:   strings.TrimSpace(apiKey),
		baseURL:  strings.TrimRight(baseURL, "/"),
		sttModel: strings.TrimSpace(sttModel),
		ttsModel: strings.TrimSpace(ttsModel),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type transcriptionResponse struct {
	Text string `json:"text"`
}

type speechRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format"`
	Stream         bool    `json:"stream"`
	Speed          float64 `json:"speed"`
	Gain           float64 `json:"gain"`
}

type SpeechAudio struct {
	Body        io.ReadCloser
	ContentType string
}

// Transcribe 将浏览器录制的音频转发给 SiliconFlow 语音转写接口。
func (client *SiliconFlowClient) Transcribe(
	ctx context.Context,
	audio io.Reader,
	filename string,
) (string, error) {
	if client.apiKey == "" {
		return "", ErrUnavailable
	}

	var requestBody bytes.Buffer
	formWriter := multipart.NewWriter(&requestBody)

	audioPart, err := formWriter.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("create audio form field: %w", err)
	}

	written, err := io.Copy(audioPart, audio)
	if err != nil {
		return "", fmt.Errorf("copy audio content: %w", err)
	}
	if written == 0 {
		return "", ErrEmptyAudio
	}

	if err := formWriter.WriteField("model", client.sttModel); err != nil {
		return "", fmt.Errorf("write transcription model: %w", err)
	}

	if err := formWriter.Close(); err != nil {
		return "", fmt.Errorf("close multipart writer: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		client.baseURL+"/audio/transcriptions",
		&requestBody,
	)
	if err != nil {
		return "", fmt.Errorf("create transcription request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+client.apiKey)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())

	response, err := client.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("send transcription request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 64<<10))
		return "", fmt.Errorf(
			"transcription provider returned status %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(message)),
		)
	}

	var payload transcriptionResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode transcription response: %w", err)
	}

	text := strings.TrimSpace(payload.Text)
	if text == "" {
		return "", ErrEmptyTranscript
	}

	return text, nil
}

// Synthesize 将文本转换为 MP3 音频，并把响应流交给 Handler。
func (client *SiliconFlowClient) Synthesize(
	ctx context.Context,
	text string,
	voice string,
	speed float64,
) (*SpeechAudio, error) {
	if client.apiKey == "" {
		return nil, ErrUnavailable
	}

	payload, err := json.Marshal(speechRequest{
		Model:          client.ttsModel,
		Input:          text,
		Voice:          voice,
		ResponseFormat: "mp3",
		Stream:         true,
		Speed:          speed,
		Gain:           0,
	})
	if err != nil {
		return nil, fmt.Errorf("encode speech request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		client.baseURL+"/audio/speech",
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("create speech request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+client.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send speech request: %w", err)
	}

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		defer response.Body.Close()

		message, _ := io.ReadAll(
			io.LimitReader(response.Body, 64<<10),
		)
		return nil, fmt.Errorf(
			"speech provider returned status %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(message)),
		)
	}

	contentType := response.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "audio/mpeg"
	}

	return &SpeechAudio{
		Body:        response.Body,
		ContentType: contentType,
	}, nil
}
