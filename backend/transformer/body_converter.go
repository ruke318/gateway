// Package transformer 提供请求体格式转换功能
package transformer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// BodyConverter 请求体转换器
type BodyConverter struct{}

// NewBodyConverter 创建 BodyConverter 实例
func NewBodyConverter() *BodyConverter {
	return &BodyConverter{}
}

// Convert 根据 body_type 转换请求体
// 支持三种格式：
// - json: application/json（默认）
// - form: application/x-www-form-urlencoded（表单提交）
// - xml: application/xml（XML 格式）
// 返回值：(body []byte, contentType string)
func (c *BodyConverter) Convert(bodyType string, reqBody []byte) ([]byte, string) {
	switch bodyType {
	case "form":
		return c.JSONToForm(reqBody), "application/x-www-form-urlencoded"
	case "xml":
		return c.JSONToXML(reqBody), "application/xml"
	default:
		return reqBody, "application/json"
	}
}

// JSONToForm 将 JSON 转换为 form 编码格式
// 示例：{"name": "Alice", "age": 30} → "name=Alice&age=30"
// 适用于传统的表单提交场景
func (c *BodyConverter) JSONToForm(jsonData []byte) []byte {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return jsonData
	}

	values := url.Values{}
	for k, v := range data {
		values.Set(k, fmt.Sprintf("%v", v))
	}

	return []byte(values.Encode())
}

// JSONToXML 将 JSON 转换为 XML 格式
// 示例：
// JSON: {"_xml_root": "order", "id": "123", "amount": 100}
// XML:
// <?xml version="1.0" encoding="UTF-8"?>
// <order>
//   <id>123</id>
//   <amount>100</amount>
// </order>
// 特殊字段 _xml_root 用于指定根节点名称，默认为 "request"
func (c *BodyConverter) JSONToXML(jsonData []byte) []byte {
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return jsonData
	}

	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")

	// 如果是 map，提取自定义根节点（如果有 _xml_root 字段）
	rootName := "request"
	if dataMap, ok := data.(map[string]interface{}); ok {
		if root, exists := dataMap["_xml_root"]; exists {
			if rootStr, ok := root.(string); ok && rootStr != "" {
				rootName = rootStr
				delete(dataMap, "_xml_root") // 移除特殊字段
			}
		}
	}

	sb.WriteString("<" + rootName + ">\n")
	c.buildXMLNode(&sb, data, 1)
	sb.WriteString("</" + rootName + ">")

	return []byte(sb.String())
}

// buildXMLNode 递归构建 XML 节点
// 支持嵌套对象和数组
// - map[string]interface{}: 转换为嵌套的 XML 标签
// - []interface{}: 转换为多个 <item> 标签
// - 基本类型: 直接作为标签内容
func (c *BodyConverter) buildXMLNode(sb *strings.Builder, data interface{}, indent int) {
	indentStr := strings.Repeat("  ", indent)

	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			sb.WriteString(indentStr + "<" + key + ">")
			if isPrimitive(val) {
				sb.WriteString(fmt.Sprintf("%v", val))
				sb.WriteString("</" + key + ">\n")
			} else {
				sb.WriteString("\n")
				c.buildXMLNode(sb, val, indent+1)
				sb.WriteString(indentStr + "</" + key + ">\n")
			}
		}
	case []interface{}:
		for _, item := range v {
			sb.WriteString(indentStr + "<item>")
			if isPrimitive(item) {
				sb.WriteString(fmt.Sprintf("%v", item))
				sb.WriteString("</item>\n")
			} else {
				sb.WriteString("\n")
				c.buildXMLNode(sb, item, indent+1)
				sb.WriteString(indentStr + "</item>\n")
			}
		}
	default:
		sb.WriteString(indentStr + fmt.Sprintf("%v\n", v))
	}
}

// isPrimitive 判断是否是基本类型
// 基本类型：string, int, int64, float64, bool, nil
// 非基本类型（map, slice）需要递归展开
func isPrimitive(v interface{}) bool {
	switch v.(type) {
	case string, int, int64, float64, bool, nil:
		return true
	default:
		return false
	}
}
