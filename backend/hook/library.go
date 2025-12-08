package hook

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ruke318/gateway/database"
)

// ScriptLibrary 公共函数库管理器
type ScriptLibrary struct {
	mu      sync.RWMutex
	scripts map[string]map[string]string // namespace -> name -> script
}

var globalLibrary = &ScriptLibrary{
	scripts: make(map[string]map[string]string),
}

// LoadLibrary 从数据库加载公共函数库
func LoadLibrary() error {
	library, err := database.GetScriptLibrary()
	if err != nil {
		return err
	}

	globalLibrary.mu.Lock()
	defer globalLibrary.mu.Unlock()
	globalLibrary.scripts = library
	return nil
}

// GetLibrary 获取全局函数库
func GetLibrary() *ScriptLibrary {
	return globalLibrary
}

// GenerateLibraryJS 生成注入到 JS 的函数库代码
func (l *ScriptLibrary) GenerateLibraryJS() string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.scripts) == 0 {
		return "var lib = {};"
	}

	var sb strings.Builder
	sb.WriteString("var lib = {};\n")

	for namespace, funcs := range l.scripts {
		sb.WriteString(fmt.Sprintf("lib.%s = {};\n", namespace))
		for name, script := range funcs {
			// 将函数包装到命名空间
			sb.WriteString(fmt.Sprintf("lib.%s.%s = %s;\n", namespace, name, script))
		}
	}

	return sb.String()
}

// ReloadLibrary 重新加载函数库
func ReloadLibrary() error {
	if err := LoadLibrary(); err != nil {
		return err
	}
	// 清空 VM 池，让新请求使用新的函数库
	if vmPool != nil {
		vmPool.Clear()
	}
	return nil
}
