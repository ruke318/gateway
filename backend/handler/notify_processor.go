// Package handler 提供回调处理器（将厂商回调转换为 InvokeRequest）
package handler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ruke318/gateway/model"
	"github.com/savsgio/atreugo/v11"
)

// NotifyProcessor 渠道回调处理器接口
// 负责将厂商回调请求转换为标准的 InvokeRequest，然后复用 invoke 流程
type NotifyProcessor interface {
	// Process 转换回调请求为 InvokeRequest
	// ctx: HTTP 请求上下文
	// unitID: 机构ID（从路由参数中获取）
	// serviceID: 服务标识（如 payNotify, refundNotify）
	// channel: 渠道标识（数字或字符串）
	Process(ctx *atreugo.RequestCtx, unitID, serviceID, channel string) (*model.InvokeRequest, error)
}

// 全局渠道处理器注册表
// 可以通过 RegisterNotifyProcessor 动态注册新的渠道处理器
var notifyProcessors = map[string]NotifyProcessor{
	"alipay": &AlipayNotifyProcessor{},
	"wechat": &WechatNotifyProcessor{},
	// 更多渠道可以在这里注册...
}

// RegisterNotifyProcessor 注册自定义渠道处理器
func RegisterNotifyProcessor(channel string, processor NotifyProcessor) {
	notifyProcessors[channel] = processor
}

// getNotifyProcessor 获取渠道处理器
// 规则：
// 1. 如果 channel 是数字，使用默认处理器
// 2. 如果 channel 是字符串，查找是否有专属处理器
// 3. 找不到则使用默认处理器兜底
func getNotifyProcessor(channel string) NotifyProcessor {
	// 判断是否是数字（默认逻辑）
	if _, err := strconv.Atoi(channel); err == nil {
		return &DefaultNotifyProcessor{}
	}

	// 查找是否有特定渠道的实现
	if processor, ok := notifyProcessors[channel]; ok {
		return processor
	}

	// 兜底：使用默认处理器
	return &DefaultNotifyProcessor{}
}

// DefaultNotifyProcessor 默认回调处理器
// 适用于数字渠道或没有特殊要求的渠道
// 特点：不验签、不解密，直接转发原始数据
type DefaultNotifyProcessor struct{}

// Process 转换请求（默认逻辑）
func (p *DefaultNotifyProcessor) Process(ctx *atreugo.RequestCtx, unitID, serviceID, channel string) (*model.InvokeRequest, error) {
	// 1. 解析请求体（支持 JSON、Form、GET 参数）
	reqData := make(map[string]interface{})

	// 判断请求类型
	method := string(ctx.Method())
	contentType := string(ctx.Request.Header.ContentType())

	if method == "GET" || contentType == "application/x-www-form-urlencoded" {
		// Form 格式 或 GET 请求：从 QueryArgs 和 PostArgs 提取参数
		ctx.QueryArgs().VisitAll(func(key, value []byte) {
			reqData[string(key)] = string(value)
		})
		ctx.PostArgs().VisitAll(func(key, value []byte) {
			reqData[string(key)] = string(value)
		})
	} else {
		// JSON 格式：解析 Body
		if len(ctx.PostBody()) > 0 {
			if err := json.Unmarshal(ctx.PostBody(), &reqData); err != nil {
				return nil, fmt.Errorf("invalid request body: %w", err)
			}
		}
	}

	// 2. 提取业务流水号（优先级：biz_no > order_no > out_trade_no > trade_no > 生成UUID）
	bizNo := extractBizNo(reqData)

	// 3. 构造 InvokeRequest
	return &model.InvokeRequest{
		ComID:     channel,   // 渠道作为 com_id
		UnitID:    unitID,    // 使用传入的机构 ID
		ServiceID: serviceID, // payNotify, refundNotify 等
		BizNo:     bizNo,     // 业务流水号
		Req:       reqData,   // 原始请求数据
	}, nil
}

// extractBizNo 从请求数据中提取业务流水号
// 按优先级尝试多个字段：biz_no > order_no > out_trade_no > trade_no
// 都没有则生成唯一 ID
func extractBizNo(data map[string]interface{}) string {
	// 优先级 1: biz_no
	if bizNo, ok := data["biz_no"].(string); ok && bizNo != "" {
		return bizNo
	}

	// 优先级 2: order_no
	if orderNo, ok := data["order_no"].(string); ok && orderNo != "" {
		return orderNo
	}

	// 优先级 3: out_trade_no（支付宝常用）
	if outTradeNo, ok := data["out_trade_no"].(string); ok && outTradeNo != "" {
		return outTradeNo
	}

	// 优先级 4: trade_no
	if tradeNo, ok := data["trade_no"].(string); ok && tradeNo != "" {
		return tradeNo
	}

	// 兜底：生成唯一 ID
	return fmt.Sprintf("notify_%d", time.Now().UnixNano())
}

// AlipayNotifyProcessor 支付宝回调处理器
// 处理支付宝的 Form 格式回调（签名验证可以在 Hook 中做）
type AlipayNotifyProcessor struct{}

// Process 转换支付宝回调
func (p *AlipayNotifyProcessor) Process(ctx *atreugo.RequestCtx, unitID, serviceID, channel string) (*model.InvokeRequest, error) {
	// 1. 提取 Form 参数（支付宝回调是 application/x-www-form-urlencoded）
	reqData := make(map[string]interface{})

	// 支持 POST Form 和 GET 参数
	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		reqData[string(key)] = string(value)
	})
	ctx.PostArgs().VisitAll(func(key, value []byte) {
		reqData[string(key)] = string(value)
	})

	// 2. 提取关键字段
	outTradeNo := ""
	if val, ok := reqData["out_trade_no"].(string); ok {
		outTradeNo = val
	}

	// 3. 构造 InvokeRequest
	return &model.InvokeRequest{
		ComID:     "alipay",   // 固定为 alipay（会查询 vendor.code='alipay' 的配置）
		UnitID:    unitID,     // 使用传入的机构 ID
		ServiceID: serviceID,  // payNotify, refundNotify 等
		BizNo:     outTradeNo, // 商户订单号作为流水号
		Req:       reqData,    // 完整回调数据（包含签名等）
	}, nil
}

// WechatNotifyProcessor 微信支付回调处理器
// 处理微信的 XML 格式回调（签名验证可以在 Hook 中做）
type WechatNotifyProcessor struct{}

// Process 转换微信回调
func (p *WechatNotifyProcessor) Process(ctx *atreugo.RequestCtx, unitID, serviceID, channel string) (*model.InvokeRequest, error) {
	// 微信回调是 XML 格式，这里简化处理，实际应该解析 XML
	// 也可以在 Hook 中用 encoding.xmlDecode() 解析

	// 暂时将 XML body 作为原始数据传递，在 Hook 中解析
	reqData := map[string]interface{}{
		"xml_body": string(ctx.PostBody()),
	}

	// 尝试从 XML 中提取 out_trade_no（简化版）
	// 实际应该使用 XML 解析库
	outTradeNo := ""
	body := string(ctx.PostBody())
	// 简单提取（生产环境应该用正规 XML 解析）
	// <out_trade_no><![CDATA[ORDER001]]></out_trade_no>
	if start := indexOf(body, "<out_trade_no>"); start != -1 {
		start += len("<out_trade_no>")
		if start2 := indexOf(body[start:], "<![CDATA["); start2 != -1 {
			start += start2 + len("<![CDATA[")
			if end := indexOf(body[start:], "]]>"); end != -1 {
				outTradeNo = body[start : start+end]
			}
		}
	}

	return &model.InvokeRequest{
		ComID:     "wechat",   // 固定为 wechat
		UnitID:    unitID,     // 使用传入的机构 ID
		ServiceID: serviceID,  // payNotify
		BizNo:     outTradeNo, // 商户订单号
		Req:       reqData,    // 原始 XML 数据
	}, nil
}

// indexOf 简单的字符串查找（Go 标准库 strings.Index）
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
