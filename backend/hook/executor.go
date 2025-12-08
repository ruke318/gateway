package hook

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
)

// VMPool JS虚拟机池
type VMPool struct {
	pool    chan *goja.Runtime
	maxSize int
	mu      sync.Mutex
}

var (
	vmPool     *VMPool
	vmPoolOnce sync.Once
)

// InitVMPool 初始化 VM 池（启动时调用）
func InitVMPool(size int) {
	vmPoolOnce.Do(func() {
		vmPool = NewVMPool(size)
	})
}

// GetVMPool 获取全局 VM 池
func GetVMPool() *VMPool {
	if vmPool == nil {
		InitVMPool(100) // 默认池大小 100
	}
	return vmPool
}

// NewVMPool 创建 VM 池
func NewVMPool(size int) *VMPool {
	return &VMPool{
		pool:    make(chan *goja.Runtime, size),
		maxSize: size,
	}
}

// Get 从池中获取 VM，没有则创建新的
func (p *VMPool) Get() *goja.Runtime {
	select {
	case vm := <-p.pool:
		return vm
	default:
		return p.createVM()
	}
}

// Put 归还 VM 到池中
func (p *VMPool) Put(vm *goja.Runtime) {
	// 清理 VM 状态
	p.resetVM(vm)

	select {
	case p.pool <- vm:
		// 成功放回池中
	default:
		// 池满了，丢弃这个 VM（让 GC 回收）
	}
}

// createVM 创建新的 VM 并初始化
func (p *VMPool) createVM() *goja.Runtime {
	vm := goja.New()

	// 注册内置模块（crypto, http, encoding, util, console）
	RegisterBuiltins(vm)

	// 注入公共函数库
	library := GetLibrary()
	libJS := library.GenerateLibraryJS()
	if libJS != "" {
		vm.RunString(libJS)
	}

	return vm
}

// resetVM 重置 VM 状态，清理上一次执行的残留
func (p *VMPool) resetVM(vm *goja.Runtime) {
	// 删除 context 变量，避免污染下次执行
	vm.Set("context", goja.Undefined())
}

// Size 返回当前池中可用的 VM 数量
func (p *VMPool) Size() int {
	return len(p.pool)
}

// Clear 清空池中所有 VM（函数库重载时调用）
func (p *VMPool) Clear() {
	for {
		select {
		case <-p.pool:
			// 丢弃，让 GC 回收
		default:
			return
		}
	}
}

// JSExecutor JS 执行器
type JSExecutor struct {
	script string
}

// NewJSExecutor 创建 JS 执行器
func NewJSExecutor(script string) *JSExecutor {
	return &JSExecutor{
		script: script,
	}
}

// Execute 执行脚本
func (e *JSExecutor) Execute(ctx *HookContext) error {
	// 从池中获取 VM
	pool := GetVMPool()
	vm := pool.Get()
	defer pool.Put(vm)

	// 设置上下文
	vm.Set("context", map[string]interface{}{
		"requestBody":     string(ctx.RequestBody),
		"responseBody":    string(ctx.ResponseBody),
		"requestHeaders":  ctx.RequestHeaders,
		"responseHeaders": ctx.ResponseHeaders,
		"data":            ctx.Data,
		"error":           ctx.Error,
	})

	// 执行脚本
	_, err := vm.RunString(e.script)
	if err != nil {
		return fmt.Errorf("JS execution error: %w", err)
	}

	// 提取结果
	result := vm.Get("context").Export()
	if resultMap, ok := result.(map[string]interface{}); ok {
		if reqBody, ok := resultMap["requestBody"].(string); ok {
			ctx.RequestBody = []byte(reqBody)
		}
		if respBody, ok := resultMap["responseBody"].(string); ok {
			ctx.ResponseBody = []byte(respBody)
		}
		if reqHeaders, ok := resultMap["requestHeaders"].(map[string]interface{}); ok {
			for k, v := range reqHeaders {
				if strVal, ok := v.(string); ok {
					ctx.RequestHeaders[k] = strVal
				}
			}
		}
		if respHeaders, ok := resultMap["responseHeaders"].(map[string]interface{}); ok {
			for k, v := range respHeaders {
				if strVal, ok := v.(string); ok {
					ctx.ResponseHeaders[k] = strVal
				}
			}
		}
		if data, ok := resultMap["data"].(map[string]interface{}); ok {
			ctx.Data = data
		}
	}

	return nil
}
