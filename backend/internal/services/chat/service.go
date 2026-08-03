package chat

import "fmt"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// 生成回复
func (*Service) GenerateReply(message string) string {
	return fmt.Sprintf(
		"我收到了你的消息：“%s”。这条回复来自 Go 后端。",
		message,
	)
}
