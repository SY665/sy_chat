package response

import "github.com/gin-gonic/gin"

type Envelope struct {
	Success bool      `json:"success"`
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, Envelope{
		Success: true,
		Data:    data,
	})
}

func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, Envelope{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}
