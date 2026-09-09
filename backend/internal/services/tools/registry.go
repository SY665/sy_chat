package tools

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

var (
	ErrToolRequired          = errors.New("tool is required")
	ErrToolNameRequired      = errors.New("tool name is required")
	ErrToolAlreadyRegistered = errors.New("tool is already registered")
	ErrToolNotFound          = errors.New("tool not found")
)

// Registry 保存当前可供 AI 调用的工具。
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具，并拒绝名称为空或重复的工具。
func (registry *Registry) Register(tool Tool) error {
	if tool == nil {
		return ErrToolRequired
	}

	name := tool.Definition().Function.Name
	if name == "" {
		return ErrToolNameRequired
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if _, exists := registry.tools[name]; exists {
		return fmt.Errorf("%w: %s", ErrToolAlreadyRegistered, name)
	}

	registry.tools[name] = tool
	return nil
}

func (registry *Registry) Get(name string) (Tool, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	tool, exists := registry.tools[name]
	return tool, exists
}

func (registry *Registry) Has(name string) bool {
	_, exists := registry.Get(name)
	return exists
}

// Definitions 按工具名称排序，保证发送给 AI Provider 的顺序稳定。
func (registry *Registry) Definitions() []Definition {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	names := make([]string, 0, len(registry.tools))
	for name := range registry.tools {
		names = append(names, name)
	}
	sort.Strings(names)

	definitions := make([]Definition, 0, len(names))
	for _, name := range names {
		definitions = append(definitions, registry.tools[name].Definition())
	}

	return definitions
}

// Execute 查找并执行模型请求的工具，同时补齐调用 ID 和工具名称。
func (registry *Registry) Execute(
	ctx context.Context,
	call Call,
) (Result, error) {
	tool, exists := registry.Get(call.Function.Name)
	if !exists {
		return Result{}, fmt.Errorf(
			"%w: %s",
			ErrToolNotFound,
			call.Function.Name,
		)
	}

	result, err := tool.Execute(ctx, []byte(call.Function.Arguments))
	result.ToolCallID = call.ID
	result.Name = call.Function.Name

	if err != nil {
		result.Success = false
		return result, fmt.Errorf(
			"execute tool %q: %w",
			call.Function.Name,
			err,
		)
	}

	result.Success = true
	return result, nil
}
