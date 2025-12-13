package util

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// PathBuilder 路径构建器
type PathBuilder struct {
	placeholderRegex *regexp.Regexp
}

// NewPathBuilder 创建 PathBuilder 实例
func NewPathBuilder() *PathBuilder {
	return &PathBuilder{
		placeholderRegex: regexp.MustCompile(`\{(\w+)\}`),
	}
}

// BuildPath 解析后端路径中的 {key} 占位符
// 支持路径参数模板，例如：
// 路径模板："/api/orders/{order_id}/pay"
// 请求数据：{"order_id": "12345"}
// 解析结果："/api/orders/12345/pay"
// 占位符值会自动 URL 转义
func (pb *PathBuilder) BuildPath(pathTemplate string, reqBody []byte) string {
	if !strings.Contains(pathTemplate, "{") {
		return pathTemplate
	}

	// 解析请求体为 map
	var data map[string]interface{}
	if err := json.Unmarshal(reqBody, &data); err != nil {
		return pathTemplate
	}

	// 匹配 {key} 占位符并替换
	result := pb.placeholderRegex.ReplaceAllStringFunc(pathTemplate, func(match string) string {
		key := match[1 : len(match)-1] // 去掉 { 和 }
		if val, ok := data[key]; ok {
			return url.QueryEscape(fmt.Sprintf("%v", val))
		}
		return match
	})

	return result
}
