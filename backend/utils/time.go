package utils

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

// UUID 生成 UUID
func UUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// Now 获取当前时间戳(秒)
func Now() int64 {
	return time.Now().Unix()
}

// FormatTime 格式化时间戳
func FormatTime(timestamp int64, layout string) string {
	layout = strings.ReplaceAll(layout, "YYYY", "2006")
	layout = strings.ReplaceAll(layout, "MM", "01")
	layout = strings.ReplaceAll(layout, "DD", "02")
	layout = strings.ReplaceAll(layout, "HH", "15")
	layout = strings.ReplaceAll(layout, "mm", "04")
	layout = strings.ReplaceAll(layout, "ss", "05")
	return time.Unix(timestamp, 0).Format(layout)
}

// ParseTime 解析时间字符串
func ParseTime(timeStr, layout string) (int64, error) {
	layout = strings.ReplaceAll(layout, "YYYY", "2006")
	layout = strings.ReplaceAll(layout, "MM", "01")
	layout = strings.ReplaceAll(layout, "DD", "02")
	layout = strings.ReplaceAll(layout, "HH", "15")
	layout = strings.ReplaceAll(layout, "mm", "04")
	layout = strings.ReplaceAll(layout, "ss", "05")
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// Sleep 休眠
func Sleep(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}
