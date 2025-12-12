// Package handler 提供请求处理器（统一调用和管理后台）
package handler

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/ruke318/gateway/database"
	"github.com/ruke318/gateway/hook"
	"github.com/ruke318/gateway/logger"
	"github.com/ruke318/gateway/model"
	"github.com/ruke318/gateway/proxy"
	"github.com/ruke318/gateway/transform"
	"github.com/ruke318/gateway/util"
	"github.com/savsgio/atreugo/v11"
	"go.uber.org/zap"
)

// InvokeHandler 统一调用入口处理器
// 负责处理 /gateway/v1/invoke 接口的所有请求
// 核心流程：
// 1. 解析请求参数（com_id, unit_id, service_id, biz_no, req）
// 2. 加载接口配置（Service + Vendor + Organization）
// 3. 构建 Hook 上下文（包含请求、路由、机构配置等）
// 4. 执行认证 Hook（BeforeAuth、AfterAuth）
// 5. 请求转换（BeforeRequestTransform → DSL → AfterRequestTransform）
// 6. 转发请求（BeforeForward → HTTP转发 → AfterForward）
// 7. 响应转换（BeforeResponseTransform → DSL → AfterResponseTransform）
// 8. 返回响应（统一格式）
type InvokeHandler struct {
	forwarder      *proxy.Forwarder           // HTTP 代理转发器
	dslTransformer *transform.DSLTransformer  // DSL 转换引擎
}

// NewInvokeHandler 创建 InvokeHandler 实例
// forwarder: HTTP 代理转发器，用于转发请求到外部厂商
// dslTransformer: DSL 转换引擎，用于执行声明式字段映射
func NewInvokeHandler(forwarder *proxy.Forwarder, dslTransformer *transform.DSLTransformer) *InvokeHandler {
	return &InvokeHandler{
		forwarder:      forwarder,
		dslTransformer: dslTransformer,
	}
}

// invokeContext 调用上下文，贯穿整个请求生命周期
// 封装了请求处理过程中所有需要共享的数据
type invokeContext struct {
	ctx     *atreugo.RequestCtx        // HTTP 请求上下文（atreugo框架）
	req     *model.InvokeRequest       // 解析后的请求参数
	svc     *model.Service             // 接口配置（包含关联的 Vendor、Organization）
	hookCtx *hook.HookContext          // Hook 执行上下文（用于 JS 脚本）
	hooks   map[string][]string        // Hook 脚本映射（执行点 → 脚本列表）
	logID   string                     // 日志追踪ID（基于 biz_no 生成）
}

// Invoke 统一调用入口
// 这是整个网关的核心方法，处理所有外部接口调用请求
// 请求格式：POST /gateway/v1/invoke
// {
//   "com_id": "alipay",          // 厂商编码
//   "unit_id": "org001",         // 机构编码
//   "service_id": "pay",         // 接口标识
//   "biz_no": "BIZ20231201001",  // 业务流水号
//   "req": {...}                 // 业务参数（统一格式）
// }
func (h *InvokeHandler) Invoke(ctx *atreugo.RequestCtx) error {
	ic := &invokeContext{ctx: ctx}

	// 1. 解析请求
	// 从 HTTP Body 中解析 JSON 参数，验证必填字段
	if err := h.parseRequest(ic); err != nil {
		return err
	}

	// 生成 LogID（用于日志追踪）
	// 基于 biz_no 生成唯一标识，方便排查问题
	ic.logID = logger.GenerateLogID(ic.req.BizNo)

	logger.Info(ic.logID, "Invoke", "请求开始",
		zap.String("unit_id", ic.req.UnitID),
		zap.String("service_id", ic.req.ServiceID),
		zap.String("com_id", ic.req.ComID),
		zap.String("biz_no", ic.req.BizNo),
	)

	// 2. 加载接口配置
	// 根据 unit_id、service_id、com_id 从数据库加载接口配置
	// 包括关联的 Vendor（厂商）、Organization（机构）、Hook 脚本等
	if err := h.loadServiceConfig(ic); err != nil {
		return err
	}

	// 3. 构建 HookContext
	// 创建 Hook 脚本执行上下文，包含请求、路由、机构配置等数据
	// 这些数据在 JS 脚本中可以通过 ctx.data.request、ctx.data.route 等访问
	h.buildHookContext(ic)

	// 4. 执行认证 Hook
	// 执行 BeforeAuth 和 AfterAuth Hook
	// 用于自定义认证逻辑（如验证签名、Token 等）
	if err := h.executeAuthHooks(ic); err != nil {
		return err
	}

	// 5. 请求转换
	// BeforeRequestTransform → DSL映射 → AfterRequestTransform
	// 将内部统一格式转换为厂商要求的格式
	if err := h.transformRequest(ic); err != nil {
		return err
	}

	// 6. 转发请求
	// BeforeForward → HTTP转发到厂商 → AfterForward
	// 支持添加签名、Token、修改请求头等
	if err := h.forwardRequest(ic); err != nil {
		return err
	}

	// 7. 响应转换
	// BeforeResponseTransform → DSL映射 → AfterResponseTransform
	// 将厂商响应格式转换回内部统一格式
	if err := h.transformResponse(ic); err != nil {
		return err
	}

	// 8. 返回响应
	// 设置响应头、状态码、Body，返回给调用方
	return h.sendResponse(ic)
}

// parseRequest 解析请求参数
// 验证必填字段：unit_id, service_id, com_id
func (h *InvokeHandler) parseRequest(ic *invokeContext) error {
	var req model.InvokeRequest
	if err := util.BindJSON(ic.ctx, &req); err != nil {
		return h.errorResponse(ic.ctx, "", 400, "invalid request body")
	}

	if req.UnitID == "" || req.ServiceID == "" || req.ComID == "" {
		return h.errorResponse(ic.ctx, "", 400, "unit_id, service_id, com_id are required")
	}

	ic.req = &req
	return nil
}

// loadServiceConfig 加载接口配置
// 根据 unit_id（机构）、service_id（接口）、com_id（厂商）联合查询接口配置
// 关联查询：Service → Organization（机构） → Vendor（厂商） → ServiceHooks（Hook脚本）
// 查询结果包含完整的三层配置数据
func (h *InvokeHandler) loadServiceConfig(ic *invokeContext) error {
	logger.Info(ic.logID, "LoadConfig", "加载接口配置")

	svc, err := database.GetServiceConfig(ic.req.UnitID, ic.req.ServiceID, ic.req.ComID)
	if err != nil {
		logger.Error(ic.logID, "LoadConfig", "接口配置加载失败", zap.Error(err))
		return h.errorResponse(ic.ctx, ic.logID, 404, err.Error())
	}

	ic.svc = svc
	ic.hooks = svc.GetHooksMap()
	return nil
}

// buildHookContext 构建 HookContext
// 创建 Hook 脚本执行上下文，包含以下数据：
// - ctx.request.headers: HTTP 请求头
// - ctx.request.body: 原始请求 Body（JSON字符串）
// - ctx.data.request: 请求元数据（method, path, query, host, header, body）
// - ctx.data.route: 路由信息（service_id, backendUrl, backendPath, backendMethod）
// - ctx.data.org_config: 机构配置（appId, secret等敏感信息）
// 这些数据在 JS Hook 脚本中可以访问和修改
func (h *InvokeHandler) buildHookContext(ic *invokeContext) {
	// 构建完整的请求数据，包含所有参数
	fullRequest := map[string]interface{}{
		"com_id":     ic.req.ComID,
		"unit_id":    ic.req.UnitID,
		"service_id": ic.req.ServiceID,
		"biz_no":     ic.req.BizNo,
		"req":        ic.req.Req,
	}
	reqBody, _ := json.Marshal(fullRequest)

	hookCtx := &hook.HookContext{
		LogID:           ic.logID,
		RequestHeaders:  make(map[string]string),
		ResponseHeaders: make(map[string]string),
		Data:            make(map[string]interface{}),
		RequestBody:     reqBody,
	}

	// 复制请求头
	ic.ctx.Request.Header.VisitAll(func(key, value []byte) {
		hookCtx.RequestHeaders[string(key)] = string(value)
	})

	// 解析请求体到 Data
	var reqData interface{}
	json.Unmarshal(reqBody, &reqData)

	hookCtx.Data["request"] = map[string]interface{}{
		"method": string(ic.ctx.Method()),
		"path":   string(ic.ctx.Path()),
		"query":  string(ic.ctx.QueryArgs().QueryString()),
		"host":   string(ic.ctx.Host()),
		"header": hookCtx.RequestHeaders,
		"body":   reqData,
	}

	hookCtx.Data["route"] = map[string]interface{}{
		"service_id":    ic.svc.ServiceID,
		"backendUrl":    ic.svc.GetBackendURL(),
		"backendPath":   ic.svc.BackendPath,
		"backendMethod": ic.svc.BackendMethod,
	}

	// 机构配置
	var orgConfig interface{}
	if ic.svc.Organization != nil && len(ic.svc.Organization.Config) > 0 {
		json.Unmarshal(ic.svc.Organization.Config, &orgConfig)
		// 如果解析后是字符串，说明可能是双重编码，再解析一次
		if orgConfigStr, ok := orgConfig.(string); ok {
			json.Unmarshal([]byte(orgConfigStr), &orgConfig)
		}
	}
	hookCtx.Data["org_config"] = orgConfig

	ic.hookCtx = hookCtx

	logger.Info(ic.logID, "BuildContext", "构建请求上下文",
		zap.String("request_body", string(reqBody)),
	)
}

// executeAuthHooks 执行认证相关 Hook
// Hook 执行点：
// - BeforeAuth: 认证前处理（如解密、参数预处理）
// - AfterAuth: 认证后处理（如Token验证）
// 认证失败返回 401，其他错误返回 500
func (h *InvokeHandler) executeAuthHooks(ic *invokeContext) error {
	// BeforeAuth
	if err := h.executeHooks(ic.hooks, model.HookBeforeAuth, ic.hookCtx); err != nil {
		logger.Error(ic.logID, "BeforeAuth", "执行失败", zap.Error(err))
		return h.errorResponse(ic.ctx, ic.logID, 401, "BeforeAuth error: "+err.Error())
	}

	// AfterAuth
	if err := h.executeHooks(ic.hooks, model.HookAfterAuth, ic.hookCtx); err != nil {
		logger.Error(ic.logID, "AfterAuth", "执行失败", zap.Error(err))
		return h.errorResponse(ic.ctx, ic.logID, 500, "AfterAuth error: "+err.Error())
	}

	return nil
}

// transformRequest 请求转换
// 转换流程：
// 1. BeforeRequestTransform Hook（前置处理）
// 2. DSL 映射转换（声明式字段映射，如：{"out_trade_no": "$.req.order_no"}）
// 3. AfterRequestTransform Hook（后置处理）
// 将内部统一格式转换为厂商要求的格式
func (h *InvokeHandler) transformRequest(ic *invokeContext) error {
	// BeforeRequestTransform Hook
	logger.Info(ic.logID, "BeforeRequestTransform", "开始执行",
		zap.String("body_before", string(ic.hookCtx.RequestBody)),
	)
	if err := h.executeHooks(ic.hooks, model.HookBeforeRequestTransform, ic.hookCtx); err != nil {
		logger.Error(ic.logID, "BeforeRequestTransform", "执行失败", zap.Error(err))
		return h.errorResponse(ic.ctx, ic.logID, 500, "BeforeRequestTransform error: "+err.Error())
	}
	logger.Info(ic.logID, "BeforeRequestTransform", "执行完成",
		zap.String("body_after", string(ic.hookCtx.RequestBody)),
	)

	// DSL 转换
	requestTransform, err := ic.svc.GetRequestTransformMap()
	logger.Info(ic.logID, "RequestTransform", "检查DSL配置",
		zap.Any("raw", string(ic.svc.RequestTransform)),
		zap.Any("parsed", requestTransform),
		zap.Error(err),
	)
	if len(requestTransform) > 0 {
		logger.Info(ic.logID, "RequestTransform", "DSL转换开始",
			zap.String("body_before", string(ic.hookCtx.RequestBody)),
		)
		transformed, err := h.dslTransformer.TransformWithContext(ic.hookCtx.RequestBody, requestTransform, ic.hookCtx.Data)
		if err != nil {
			logger.Error(ic.logID, "RequestTransform", "DSL转换失败", zap.Error(err))
			return h.errorResponse(ic.ctx, ic.logID, 500, "request transform error: "+err.Error())
		}
		ic.hookCtx.RequestBody = transformed
		logger.Info(ic.logID, "RequestTransform", "DSL转换完成",
			zap.String("body_after", string(ic.hookCtx.RequestBody)),
		)
	}

	// AfterRequestTransform Hook
	logger.Info(ic.logID, "AfterRequestTransform", "开始执行",
		zap.String("body_before", string(ic.hookCtx.RequestBody)),
	)
	if err := h.executeHooks(ic.hooks, model.HookAfterRequestTransform, ic.hookCtx); err != nil {
		logger.Error(ic.logID, "AfterRequestTransform", "执行失败", zap.Error(err))
		return h.errorResponse(ic.ctx, ic.logID, 500, "AfterRequestTransform error: "+err.Error())
	}
	logger.Info(ic.logID, "AfterRequestTransform", "执行完成",
		zap.String("body_after", string(ic.hookCtx.RequestBody)),
	)

	return nil
}

// forwardRequest 转发请求
// 转发流程：
// 1. BeforeForward Hook（前置处理，如添加签名、获取Token）
// 2. 构建后端请求（URL、Path、Method、Header、Body）
// 3. HTTP 转发到厂商后端
// 4. AfterForward Hook（后置处理，如解密响应）
// 支持在 Hook 中动态修改 backendUrl、backendPath、backendMethod
func (h *InvokeHandler) forwardRequest(ic *invokeContext) error {
	// BeforeForward Hook
	logger.Info(ic.logID, "BeforeForward", "开始执行",
		zap.String("body_before", string(ic.hookCtx.RequestBody)),
	)
	if err := h.executeHooks(ic.hooks, model.HookBeforeForward, ic.hookCtx); err != nil {
		logger.Error(ic.logID, "BeforeForward", "执行失败", zap.Error(err))
		return h.errorResponse(ic.ctx, ic.logID, 500, "BeforeForward error: "+err.Error())
	}
	logger.Info(ic.logID, "BeforeForward", "执行完成",
		zap.String("body_after", string(ic.hookCtx.RequestBody)),
	)

	// 构建后端请求（支持在 Hook 中修改）
	backendURL := ic.svc.GetBackendURL()
	backendPath := ic.svc.BackendPath
	backendMethod := ic.svc.BackendMethod
	if backendMethod == "" {
		backendMethod = "POST"
	}

	// 从 Hook 上下文中读取可能被修改的路由信息
	if routeData, ok := ic.hookCtx.Data["route"].(map[string]interface{}); ok {
		if url, ok := routeData["backendUrl"].(string); ok && url != "" {
			backendURL = url
		}
		if path, ok := routeData["backendPath"].(string); ok {
			backendPath = path
		}
		if method, ok := routeData["backendMethod"].(string); ok && method != "" {
			backendMethod = method
		}
	}

	// 构建完整的后端路径（支持模板变量）
	backendPath = h.buildBackendPath(backendPath, ic.hookCtx.RequestBody)

	// 构建请求体和 Content-Type
	body, contentType := h.buildRequestBody(ic.svc.BodyType, ic.hookCtx.RequestBody)

	// 构建请求头
	header := make(map[string][]string)
	ic.ctx.Request.Header.VisitAll(func(key, value []byte) {
		header[string(key)] = []string{string(value)}
	})
	header["Content-Type"] = []string{contentType}

	logger.Info(ic.logID, "Forward", "转发请求",
		zap.String("method", backendMethod),
		zap.String("url", backendURL),
		zap.String("path", backendPath),
		zap.String("content_type", contentType),
		zap.String("body", string(body)),
	)

	// 转发
	resp, respBody, err := h.forwarder.ForwardWithOptions(backendMethod, backendURL, backendPath, body, header)
	if err != nil {
		logger.Error(ic.logID, "Forward", "转发失败", zap.Error(err))
		h.executeHooks(ic.hooks, model.HookOnError, ic.hookCtx)
		return h.errorResponse(ic.ctx, ic.logID, 502, "forward error: "+err.Error())
	}

	logger.Info(ic.logID, "Forward", "转发成功",
		zap.Int("status", resp.StatusCode),
		zap.String("response_body", string(respBody)),
	)

	ic.hookCtx.Response = resp
	ic.hookCtx.ResponseBody = respBody

	// 更新响应数据到 HookContext
	responseHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			responseHeaders[k] = v[0]
		}
	}
	ic.hookCtx.Data["response"] = map[string]interface{}{
		"status": resp.StatusCode,
		"header": responseHeaders,
	}

	// AfterForward Hook
	logger.Info(ic.logID, "AfterForward", "开始执行",
		zap.String("body_before", string(ic.hookCtx.ResponseBody)),
	)
	if err := h.executeHooks(ic.hooks, model.HookAfterForward, ic.hookCtx); err != nil {
		logger.Error(ic.logID, "AfterForward", "执行失败", zap.Error(err))
		return h.errorResponse(ic.ctx, ic.logID, 500, "AfterForward error: "+err.Error())
	}
	logger.Info(ic.logID, "AfterForward", "执行完成",
		zap.String("body_after", string(ic.hookCtx.ResponseBody)),
	)

	return nil
}

// transformResponse 响应转换
// 转换流程：
// 1. BeforeResponseTransform Hook（前置处理）
// 2. DSL 映射转换（将厂商响应转换为统一格式）
// 3. AfterResponseTransform Hook（后置处理）
func (h *InvokeHandler) transformResponse(ic *invokeContext) error {
	// BeforeResponseTransform Hook
	logger.Info(ic.logID, "BeforeResponseTransform", "开始执行",
		zap.String("body_before", string(ic.hookCtx.ResponseBody)),
	)
	if err := h.executeHooks(ic.hooks, model.HookBeforeResponseTransform, ic.hookCtx); err != nil {
		logger.Error(ic.logID, "BeforeResponseTransform", "执行失败", zap.Error(err))
		return h.errorResponse(ic.ctx, ic.logID, 500, "BeforeResponseTransform error: "+err.Error())
	}
	logger.Info(ic.logID, "BeforeResponseTransform", "执行完成",
		zap.String("body_after", string(ic.hookCtx.ResponseBody)),
	)

	// DSL 转换
	responseTransform, _ := ic.svc.GetResponseTransformMap()
	if len(responseTransform) > 0 {
		logger.Info(ic.logID, "ResponseTransform", "DSL转换开始",
			zap.String("body_before", string(ic.hookCtx.ResponseBody)),
		)
		transformed, err := h.dslTransformer.TransformWithContext(ic.hookCtx.ResponseBody, responseTransform, ic.hookCtx.Data)
		if err != nil {
			logger.Error(ic.logID, "ResponseTransform", "DSL转换失败", zap.Error(err))
			return h.errorResponse(ic.ctx, ic.logID, 500, "response transform error: "+err.Error())
		}
		ic.hookCtx.ResponseBody = transformed
		logger.Info(ic.logID, "ResponseTransform", "DSL转换完成",
			zap.String("body_after", string(ic.hookCtx.ResponseBody)),
		)
	}

	// AfterResponseTransform Hook
	logger.Info(ic.logID, "AfterResponseTransform", "开始执行",
		zap.String("body_before", string(ic.hookCtx.ResponseBody)),
	)
	if err := h.executeHooks(ic.hooks, model.HookAfterResponseTransform, ic.hookCtx); err != nil {
		logger.Error(ic.logID, "AfterResponseTransform", "执行失败", zap.Error(err))
		return h.errorResponse(ic.ctx, ic.logID, 500, "AfterResponseTransform error: "+err.Error())
	}
	logger.Info(ic.logID, "AfterResponseTransform", "执行完成",
		zap.String("body_after", string(ic.hookCtx.ResponseBody)),
	)

	return nil
}

// sendResponse 返回响应
// 设置响应头、状态码、Body，返回给调用方
func (h *InvokeHandler) sendResponse(ic *invokeContext) error {
	for k, v := range ic.hookCtx.ResponseHeaders {
		ic.ctx.Response.Header.Set(k, v)
	}
	ic.ctx.Response.Header.Set("Content-Type", "application/json")
	ic.ctx.SetStatusCode(ic.hookCtx.Response.StatusCode)
	ic.ctx.SetBody(ic.hookCtx.ResponseBody)

	logger.Info(ic.logID, "Response", "请求完成",
		zap.Int("status", ic.hookCtx.Response.StatusCode),
		zap.String("body", string(ic.hookCtx.ResponseBody)),
	)

	return nil
}

// buildBackendPath 解析后端路径中的 {key} 占位符
// 支持路径参数模板，例如：
// 路径模板："/api/orders/{order_id}/pay"
// 请求数据：{"order_id": "12345"}
// 解析结果："/api/orders/12345/pay"
// 占位符值会自动 URL 转义
func (h *InvokeHandler) buildBackendPath(pathTemplate string, reqBody []byte) string {
	if !strings.Contains(pathTemplate, "{") {
		return pathTemplate
	}

	// 解析请求体为 map
	var data map[string]interface{}
	if err := json.Unmarshal(reqBody, &data); err != nil {
		return pathTemplate
	}

	// 匹配 {key} 占位符
	re := regexp.MustCompile(`\{(\w+)\}`)
	result := re.ReplaceAllStringFunc(pathTemplate, func(match string) string {
		key := match[1 : len(match)-1] // 去掉 { 和 }
		if val, ok := data[key]; ok {
			return url.QueryEscape(fmt.Sprintf("%v", val))
		}
		return match
	})

	return result
}

// buildRequestBody 根据 body_type 构建请求体
// 支持三种格式：
// - json: application/json（默认）
// - form: application/x-www-form-urlencoded（表单提交）
// - xml: application/xml（XML格式）
// 返回值：(body []byte, contentType string)
func (h *InvokeHandler) buildRequestBody(bodyType string, reqBody []byte) ([]byte, string) {
	switch bodyType {
	case "form":
		return h.jsonToForm(reqBody), "application/x-www-form-urlencoded"
	case "xml":
		return h.jsonToXML(reqBody), "application/xml"
	default:
		return reqBody, "application/json"
	}
}

// jsonToForm 将 JSON 转换为 form 编码格式
// 示例：{"name": "Alice", "age": 30} → "name=Alice&age=30"
// 适用于传统的表单提交场景
func (h *InvokeHandler) jsonToForm(jsonData []byte) []byte {
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

// jsonToXML 将 JSON 转换为 XML 格式
// 示例：
// JSON: {"_xml_root": "order", "id": "123", "amount": 100}
// XML:
// <?xml version="1.0" encoding="UTF-8"?>
// <order>
//   <id>123</id>
//   <amount>100</amount>
// </order>
// 特殊字段 _xml_root 用于指定根节点名称，默认为 "request"
func (h *InvokeHandler) jsonToXML(jsonData []byte) []byte {
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
	h.buildXMLNode(&sb, data, 1)
	sb.WriteString("</" + rootName + ">")

	return []byte(sb.String())
}

// buildXMLNode 递归构建 XML 节点
// 支持嵌套对象和数组
// - map[string]interface{}: 转换为嵌套的 XML 标签
// - []interface{}: 转换为多个 <item> 标签
// - 基本类型: 直接作为标签内容
func (h *InvokeHandler) buildXMLNode(sb *strings.Builder, data interface{}, indent int) {
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
				h.buildXMLNode(sb, val, indent+1)
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
				h.buildXMLNode(sb, item, indent+1)
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

// executeHooks 执行指定节点的所有 Hook
// hookPoint: Hook 执行点（BeforeAuth, AfterAuth, BeforeForward 等）
// ctx: Hook 上下文，包含请求数据、响应数据、自定义数据
// 按顺序执行该执行点配置的所有 Hook 脚本
// 任何一个脚本执行失败都会中断后续执行并返回错误
func (h *InvokeHandler) executeHooks(hooks map[string][]string, hookPoint string, ctx *hook.HookContext) error {
	scripts, ok := hooks[hookPoint]
	if !ok || len(scripts) == 0 {
		return nil
	}

	for _, script := range scripts {
		executor := hook.NewJSExecutor(script)
		if err := executor.Execute(ctx); err != nil {
			return fmt.Errorf("hook %s execute error: %w", hookPoint, err)
		}
	}
	return nil
}

// errorResponse 返回错误响应
// 记录错误日志并返回统一格式的错误响应
// 响应格式：{"code": xxx, "message": "xxx", "log_id": "xxx"}
func (h *InvokeHandler) errorResponse(ctx *atreugo.RequestCtx, logID string, code int, message string) error {
	if logID != "" {
		logger.Error(logID, "Error", message, zap.Int("code", code))
	}
	return ctx.JSONResponse(model.InvokeResponse{
		Code:    code,
		Message: message,
		LogID:   logID,
	}, code)
}
