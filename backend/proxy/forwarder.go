package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/ruke318/gateway/config"
)

// Forwarder HTTP 转发器，按厂商（host）管理连接池和熔断器
type Forwarder struct {
	clients sync.Map               // host -> *http.Client
	config  *config.HTTPPoolConfig // 连接池配置
	breaker *CircuitBreakerManager // 熔断器管理器
}

// NewForwarder 创建 Forwarder 实例
func NewForwarder(poolCfg *config.HTTPPoolConfig, cbCfg *config.CircuitBreakerConfig) *Forwarder {
	return &Forwarder{
		config:  poolCfg,
		breaker: NewCircuitBreakerManager(cbCfg),
	}
}

// getOrCreateClient 获取或创建指定 host 的 HTTP Client
func (f *Forwarder) getOrCreateClient(host string) *http.Client {
	// 尝试从缓存获取
	if client, ok := f.clients.Load(host); ok {
		return client.(*http.Client)
	}

	// 创建新的 client
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: f.config.MaxIdleConnsPerHost,
		MaxConnsPerHost:     f.config.MaxConnsPerHost,
		IdleConnTimeout:     time.Duration(f.config.IdleConnTimeout) * time.Second,
		DisableKeepAlives:   false,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(f.config.RequestTimeout) * time.Second,
	}

	// 存储到缓存（可能有并发，但没关系，多创建一个也无妨）
	f.clients.Store(host, client)
	return client
}

// extractHost 从 URL 中提取 host
func extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Host
}

// Forward 转发请求（兼容旧接口）
func (f *Forwarder) Forward(req *http.Request, body []byte) (*http.Response, []byte, error) {
	return f.ForwardWithOptions(req.Method, req.URL.String(), "", body, req.Header)
}

// ForwardWithOptions 转发请求到指定后端
func (f *Forwarder) ForwardWithOptions(method, backendURL, path string, body []byte, headers http.Header) (*http.Response, []byte, error) {
	fullURL := backendURL + path
	host := extractHost(backendURL)

	// 检查熔断状态
	if !f.breaker.Allow(host) {
		return nil, nil, ErrCircuitOpen
	}

	// 获取对应 host 的 client
	client := f.getOrCreateClient(host)

	// 创建请求
	proxyReq, err := http.NewRequest(method, fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	// 复制请求头
	for k, v := range headers {
		proxyReq.Header[k] = v
	}

	// 发起请求
	resp, err := client.Do(proxyReq)
	if err != nil {
		f.breaker.RecordFailure(host)
		return nil, nil, err
	}

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		f.breaker.RecordFailure(host)
		return nil, nil, err
	}

	// 根据状态码判断成功/失败（5xx 视为失败）
	if resp.StatusCode >= 500 {
		f.breaker.RecordFailure(host)
	} else {
		f.breaker.RecordSuccess(host)
	}

	return resp, respBody, nil
}
