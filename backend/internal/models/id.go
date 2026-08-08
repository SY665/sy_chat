package models

import "github.com/google/uuid"

// newID 为新记录生成 UUID 主键。
func newID() string {
	return uuid.NewString()
}
