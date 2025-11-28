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

func (m *Manager) Register(point HookPoint, hook Hook) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks[point] = append(m.hooks[point], hook)
	return nil
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
