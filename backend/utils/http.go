package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPGet 发送 GET 请求
func HTTPGet(urlStr string, headers map[string]string) (map[string]interface{}, error) {
	return HTTPRequest("GET", urlStr, nil, headers)
}

// HTTPPost 发送 POST 请求(原始body)
func HTTPPost(urlStr string, body string, headers map[string]string) (map[string]interface{}, error) {
	return HTTPRequest("POST", urlStr, body, headers)
}

// HTTPPostJSON 发送 POST JSON 请求
func HTTPPostJSON(urlStr string, data interface{}, headers map[string]string) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"
	return HTTPRequest("POST", urlStr, string(jsonBytes), headers)
}

// HTTPPostForm 发送 POST 表单请求
func HTTPPostForm(urlStr string, data map[string]string, headers map[string]string) (map[string]interface{}, error) {
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	return HTTPRequest("POST", urlStr, form.Encode(), headers)
}

// HTTPRequest 发送自定义 HTTP 请求
func HTTPRequest(method, urlStr string, body interface{}, headers map[string]string) (map[string]interface{}, error) {
	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		default:
			jsonBytes, _ := json.Marshal(v)
			bodyReader = bytes.NewReader(jsonBytes)
		}
	}

	req, err := http.NewRequest(method, urlStr, bodyReader)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jsonData interface{}
	if err := json.Unmarshal(respBody, &jsonData); err == nil {
		return map[string]interface{}{
			"status":  resp.StatusCode,
			"headers": resp.Header,
			"body":    string(respBody),
			"json":    jsonData,
		}, nil
	}

	return map[string]interface{}{
		"status":  resp.StatusCode,
		"headers": resp.Header,
		"body":    string(respBody),
		"json":    nil,
	}, nil
}
