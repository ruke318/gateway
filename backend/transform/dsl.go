package transform

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/oliveagle/jsonpath"
	"github.com/ruke318/gateway/hook"
)

type DSLTransformer struct{}

func NewDSLTransformer() *DSLTransformer {
	return &DSLTransformer{}
}

func (t *DSLTransformer) Transform(data []byte, template map[string]interface{}) ([]byte, error) {
	return t.TransformWithContext(data, template, nil)
}

func (t *DSLTransformer) TransformWithContext(data []byte, template map[string]interface{}, contextData map[string]interface{}) ([]byte, error) {
	if len(template) == 0 {
		return data, nil
	}

	var sourceData interface{}
	if err := json.Unmarshal(data, &sourceData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	result, err := t.processTemplate(sourceData, template, contextData)
	if err != nil {
		return nil, err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return output, nil
}

func (t *DSLTransformer) processTemplate(sourceData interface{}, template map[string]interface{}, contextData map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range template {
		processed, err := t.processValue(sourceData, value, contextData)
		if err != nil {
			return nil, fmt.Errorf("failed to process key %s: %w", key, err)
		}
		result[key] = processed
	}

	return result, nil
}

func (t *DSLTransformer) processValue(sourceData interface{}, value interface{}, contextData map[string]interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return t.processString(sourceData, v, contextData)
	case map[string]interface{}:
		if jsonPath, ok := v["json.path"].(string); ok {
			return t.processArray(sourceData, jsonPath, v, contextData)
		}
		return t.processTemplate(sourceData, v, contextData)
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			processed, err := t.processValue(sourceData, item, contextData)
			if err != nil {
				return nil, err
			}
			result[i] = processed
		}
		return result, nil
	default:
		return v, nil
	}
}

func (t *DSLTransformer) processString(sourceData interface{}, value string, contextData map[string]interface{}) (interface{}, error) {
	// 1. 处理函数调用: @fn.dict.get("payment_method", "ALIPAY")
	if strings.HasPrefix(value, "@fn.") {
		return t.processFunctionCall(sourceData, value, contextData)
	}

	// 2. 处理 Context 注入: @ctx.org_config.appId
	if strings.HasPrefix(value, "@ctx.") {
		if contextData == nil {
			return nil, nil
		}
		ctxPath := strings.TrimPrefix(value, "@ctx.")
		return t.getContextValue(contextData, ctxPath), nil
	}

	// 3. 处理 JSONPath 查询: $.req.order_no
	if !strings.HasPrefix(value, "$.") {
		return value, nil
	}

	if value == "$." {
		return sourceData, nil
	}

	result, err := jsonpath.JsonPathLookup(sourceData, value)
	if err != nil {
		return nil, nil
	}

	return result, nil
}

func (t *DSLTransformer) getContextValue(contextData map[string]interface{}, path string) interface{} {
	keys := strings.Split(path, ".")
	var current interface{} = contextData

	for _, key := range keys {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[key]
		} else {
			return nil
		}
	}

	return current
}

func (t *DSLTransformer) processArray(sourceData interface{}, arrayPath string, itemTemplate map[string]interface{}, contextData map[string]interface{}) (interface{}, error) {
	arrayData, err := jsonpath.JsonPathLookup(sourceData, arrayPath)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup array path %s: %w", arrayPath, err)
	}

	arraySlice, ok := arrayData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("path %s does not point to an array", arrayPath)
	}

	result := make([]interface{}, 0, len(arraySlice))

	templateCopy := make(map[string]interface{})
	for k, v := range itemTemplate {
		if k != "json.path" {
			templateCopy[k] = v
		}
	}

	for _, item := range arraySlice {
		processedItem, err := t.processTemplate(item, templateCopy, contextData)
		if err != nil {
			return nil, fmt.Errorf("failed to process array item: %w", err)
		}
		result = append(result, processedItem)
	}

	return result, nil
}

// ============ 函数调用支持 ============

// functionCall 表示一个函数调用
type functionCall struct {
	FunctionPath string   // 函数路径，如 "dict.get" 或 "lib.common.buildSign"
	Args         []string // 参数列表（原始字符串）
}

// processFunctionCall 处理函数调用表达式
// 支持的语法：@fn.dict.get("payment_method", "ALIPAY")
// 参数可以是：JSONPath（$.req.field）、Context（@ctx.field）、字面量（"string", 123, true）
func (t *DSLTransformer) processFunctionCall(sourceData interface{}, expr string, contextData map[string]interface{}) (interface{}, error) {
	// 1. 解析函数表达式
	funcCall, err := t.parseFunctionExpr(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse function expression: %w", err)
	}

	// 2. 解析参数（解析 $. 和 @ctx. 变量）
	args := make([]interface{}, len(funcCall.Args))
	for i, argStr := range funcCall.Args {
		resolved, err := t.resolveArgument(sourceData, argStr, contextData)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve argument %d: %w", i, err)
		}
		args[i] = resolved
	}

	// 3. 执行函数
	result, err := t.executeFunction(funcCall.FunctionPath, args, sourceData, contextData)
	if err != nil {
		return nil, fmt.Errorf("failed to execute function %s: %w", funcCall.FunctionPath, err)
	}

	return result, nil
}

// parseFunctionExpr 解析函数表达式
// 输入: @fn.dict.get("payment_method", "ALIPAY")
// 输出: {FunctionPath: "dict.get", Args: ["\"payment_method\"", "\"ALIPAY\""]}
func (t *DSLTransformer) parseFunctionExpr(expr string) (*functionCall, error) {
	// 去掉 @fn. 前缀
	expr = strings.TrimPrefix(expr, "@fn.")

	// 使用正则表达式匹配: functionPath(arg1, arg2, ...)
	re := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_.]*)\((.*)\)$`)
	matches := re.FindStringSubmatch(expr)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid function expression: %s", expr)
	}

	funcPath := matches[1]
	argsStr := matches[2]

	// 解析参数列表（支持嵌套括号和引号）
	args := t.parseArguments(argsStr)

	return &functionCall{
		FunctionPath: funcPath,
		Args:         args,
	}, nil
}

// parseArguments 解析参数列表
// 输入: "payment_method", "ALIPAY", $.req.amount, @ctx.secret
// 输出: ["\"payment_method\"", "\"ALIPAY\"", "$.req.amount", "@ctx.secret"]
func (t *DSLTransformer) parseArguments(argsStr string) []string {
	if strings.TrimSpace(argsStr) == "" {
		return []string{}
	}

	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	depth := 0 // 括号深度

	for _, ch := range argsStr {
		switch ch {
		case '"', '\'':
			if inQuotes {
				if ch == quoteChar {
					inQuotes = false
					quoteChar = 0
				}
			} else {
				inQuotes = true
				quoteChar = ch
			}
			current.WriteRune(ch)

		case '(':
			if !inQuotes {
				depth++
			}
			current.WriteRune(ch)

		case ')':
			if !inQuotes {
				depth--
			}
			current.WriteRune(ch)

		case ',':
			if !inQuotes && depth == 0 {
				// 遇到分隔符，保存当前参数
				arg := strings.TrimSpace(current.String())
				if arg != "" {
					args = append(args, arg)
				}
				current.Reset()
			} else {
				current.WriteRune(ch)
			}

		default:
			current.WriteRune(ch)
		}
	}

	// 保存最后一个参数
	arg := strings.TrimSpace(current.String())
	if arg != "" {
		args = append(args, arg)
	}

	return args
}

// resolveArgument 解析参数值
// 支持：JSONPath（$.req.field）、Context（@ctx.field）、字面量（"string", 123, true, null）
func (t *DSLTransformer) resolveArgument(sourceData interface{}, argStr string, contextData map[string]interface{}) (interface{}, error) {
	argStr = strings.TrimSpace(argStr)

	// 1. JSONPath 查询
	if strings.HasPrefix(argStr, "$.") {
		if argStr == "$." {
			return sourceData, nil
		}
		result, err := jsonpath.JsonPathLookup(sourceData, argStr)
		if err != nil {
			return nil, fmt.Errorf("JSONPath lookup failed: %w", err)
		}
		return result, nil
	}

	// 2. Context 注入
	if strings.HasPrefix(argStr, "@ctx.") {
		if contextData == nil {
			return nil, nil
		}
		ctxPath := strings.TrimPrefix(argStr, "@ctx.")
		return t.getContextValue(contextData, ctxPath), nil
	}

	// 3. 字面量：字符串、数字、布尔、null
	return t.parseLiteral(argStr)
}

// parseLiteral 解析字面量
// 支持: "string", 'string', 123, 123.45, true, false, null
func (t *DSLTransformer) parseLiteral(str string) (interface{}, error) {
	str = strings.TrimSpace(str)

	// null
	if str == "null" {
		return nil, nil
	}

	// boolean
	if str == "true" {
		return true, nil
	}
	if str == "false" {
		return false, nil
	}

	// 字符串（带引号）
	if (strings.HasPrefix(str, "\"") && strings.HasSuffix(str, "\"")) ||
		(strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'")) {
		return str[1 : len(str)-1], nil
	}

	// 数字（尝试解析为 JSON）
	var num interface{}
	if err := json.Unmarshal([]byte(str), &num); err == nil {
		return num, nil
	}

	// 默认返回原字符串
	return str, nil
}

// executeFunction 执行函数
// 使用 Goja VM 执行函数调用
func (t *DSLTransformer) executeFunction(funcPath string, args []interface{}, sourceData interface{}, contextData map[string]interface{}) (interface{}, error) {
	// 从 VM 池获取 VM
	pool := hook.GetVMPool()
	vm := pool.Get()
	defer pool.Put(vm)

	// 构建 context 对象（模拟 Hook 上下文）
	// 注意：这里需要构建一个简化的 context，包含 data 字段
	// 以便字典函数能够获取 unit_id
	ctxData := map[string]interface{}{
		"data": contextData,
	}

	// 设置 context 到 VM
	vm.Set("context", ctxData)

	// 构建函数调用代码
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal arguments: %w", err)
	}

	// JavaScript 代码：调用函数并返回结果
	script := fmt.Sprintf("(%s).apply(null, %s)", funcPath, string(argsJSON))

	// 执行脚本
	result, err := vm.RunString(script)
	if err != nil {
		return nil, fmt.Errorf("JS execution error: %w", err)
	}

	// 导出结果（转换为 Go 类型）
	return result.Export(), nil
}

