package proxy

import (
	"errors"
	"sync"
	"time"

	"github.com/ruke318/gateway/config"
)

// ErrCircuitOpen 熔断器打开错误
var ErrCircuitOpen = errors.New("circuit breaker is open")

// State 熔断器状态
type State int

const (
	StateClosed   State = iota // 关闭（正常）
	StateOpen                  // 打开（熔断中）
	StateHalfOpen              // 半开（探测中）
)

// CircuitBreaker 单个 host 的熔断器
type CircuitBreaker struct {
	state           State
	failures        int       // 连续失败次数
	successes       int       // 半开状态连续成功次数
	lastFailureTime time.Time // 最后失败时间
	config          *config.CircuitBreakerConfig
	mu              sync.RWMutex
}

// CircuitBreakerManager 熔断器管理器，按 host 管理
type CircuitBreakerManager struct {
	breakers sync.Map // host -> *CircuitBreaker
	config   *config.CircuitBreakerConfig
}

// NewCircuitBreakerManager 创建熔断器管理器
func NewCircuitBreakerManager(cfg *config.CircuitBreakerConfig) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		config: cfg,
	}
}

// getOrCreate 获取或创建指定 host 的熔断器
func (m *CircuitBreakerManager) getOrCreate(host string) *CircuitBreaker {
	if cb, ok := m.breakers.Load(host); ok {
		return cb.(*CircuitBreaker)
	}
	cb := &CircuitBreaker{
		state:  StateClosed,
		config: m.config,
	}
	m.breakers.Store(host, cb)
	return cb
}

// Allow 判断是否允许请求通过
func (m *CircuitBreakerManager) Allow(host string) bool {
	if !m.config.Enabled {
		return true
	}
	return m.getOrCreate(host).Allow()
}

// RecordSuccess 记录请求成功
func (m *CircuitBreakerManager) RecordSuccess(host string) {
	if !m.config.Enabled {
		return
	}
	m.getOrCreate(host).RecordSuccess()
}

// RecordFailure 记录请求失败
func (m *CircuitBreakerManager) RecordFailure(host string) {
	if !m.config.Enabled {
		return
	}
	m.getOrCreate(host).RecordFailure()
}

// Allow 判断是否允许请求
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// 检查是否超过熔断时间，进入半开状态
		if time.Since(cb.lastFailureTime) > time.Duration(cb.config.Timeout)*time.Second {
			cb.state = StateHalfOpen
			cb.successes = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return true
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failures = 0
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
		}
	}
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.config.FailureThreshold {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		cb.state = StateOpen
		cb.successes = 0
	}
}
