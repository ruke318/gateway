package hook

import (
	"fmt"
	"log"

	"github.com/dop251/goja"
)

type JSExecutor struct {
	vm     *goja.Runtime
	script string
}

func NewJSExecutor(script string) *JSExecutor {
	vm := goja.New()

	// 注册console对象
	console := vm.NewObject()
	console.Set("log", func(args ...interface{}) { log.Println(args...) })
	console.Set("info", func(args ...interface{}) {
		log.Println(append([]interface{}{"[INFO]"}, args...)...)
	})
	console.Set("warn", func(args ...interface{}) {
		log.Println(append([]interface{}{"[WARN]"}, args...)...)
	})
	console.Set("error", func(args ...interface{}) {
		log.Println(append([]interface{}{"[ERROR]"}, args...)...)
	})
	vm.Set("console", console)

	// 注册全局函数
	vm.Set("setTimeout", func(fn func(), delay int) {})
	vm.Set("setInterval", func(fn func(), delay int) {})

	return &JSExecutor{
		vm:     vm,
		script: script,
	}
}

func (e *JSExecutor) Execute(ctx *HookContext) error {
	e.vm.Set("context", map[string]interface{}{
		"requestBody":     string(ctx.RequestBody),
		"responseBody":    string(ctx.ResponseBody),
		"requestHeaders":  ctx.RequestHeaders,
		"responseHeaders": ctx.ResponseHeaders,
		"data":            ctx.Data,
		"error":           ctx.Error,
	})

	_, err := e.vm.RunString(e.script)
	if err != nil {
		return fmt.Errorf("JS execution error: %w", err)
	}

	result := e.vm.Get("context").Export()
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
