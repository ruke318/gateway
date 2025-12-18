package utils

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/url"
	"strings"
)

// Base64Encode Base64 编码
func Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// Base64Decode Base64 解码
func Base64Decode(data string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// HexEncode Hex 编码
func HexEncode(data string) string {
	return hex.EncodeToString([]byte(data))
}

// HexDecode Hex 解码
func HexDecode(data string) (string, error) {
	decoded, err := hex.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// JSONEncode JSON 序列化
func JSONEncode(data interface{}) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// JSONDecode JSON 反序列化
func JSONDecode(data string) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// XMLEncode XML 序列化
func XMLEncode(data interface{}) (string, error) {
	xmlBytes, err := xml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(xmlBytes), nil
}

// XMLDecode XML 反序列化
func XMLDecode(data string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	decoder := xml.NewDecoder(strings.NewReader(data))
	var stack []string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := token.(type) {
		case xml.StartElement:
			stack = append(stack, t.Name.Local)
		case xml.CharData:
			if len(stack) > 0 {
				key := strings.Join(stack, ".")
				result[key] = strings.TrimSpace(string(t))
			}
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return result, nil
}

// URLEncode URL 编码
func URLEncode(data string) string {
	return url.QueryEscape(data)
}

// URLDecode URL 解码
func URLDecode(data string) (string, error) {
	return url.QueryUnescape(data)
}
