package hook

import (
	"io/ioutil"
	"sync"
)

type Manager struct {
	hooks map[HookPoint][]Hook
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		hooks: make(map[HookPoint][]Hook),
	}
}

func (m *Manager) RegisterScript(point HookPoint, scriptPath string) error {
	script, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		return err
	}
	return m.Register(point, NewJSExecutor(string(script)))
}

// RegisterScriptString 直接注册字符串形式的 JavaScript 脚本
// 适用于从数据库或其他存储读取的脚本
func (m *Manager) RegisterScriptString(point HookPoint, scriptContent string) error {
	return m.Register(point, NewJSExecutor(scriptContent))
}

func (m *Manager) Register(point HookPoint, hook Hook) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks[point] = append(m.hooks[point], hook)
	return nil
}

// UpdateHook 更新指定 HookPoint 的所有 Hook（替换）
func (m *Manager) UpdateHook(point HookPoint, scriptContent string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 替换该 HookPoint 的所有 Hook
	m.hooks[point] = []Hook{NewJSExecutor(scriptContent)}
	return nil
}

// ClearHook 清空指定 HookPoint 的所有 Hook
func (m *Manager) ClearHook(point HookPoint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.hooks, point)
}

// GetHookCount 获取指定 HookPoint 的 Hook 数量
func (m *Manager) GetHookCount(point HookPoint) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.hooks[point])
}

func (m *Manager) Execute(point HookPoint, ctx *HookContext) error {
	m.mu.RLock()
	hooks := m.hooks[point]
	m.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook.Execute(ctx); err != nil {
			return err
		}
	}
	return nil
}
