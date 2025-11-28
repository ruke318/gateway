// 请求/响应转换钩子示例

// 请求转换 - 修改请求体
if (context.requestBody) {
    var req = JSON.parse(context.requestBody);
    req.timestamp = Date.now();
    req.gateway = "v1.0";
    context.requestBody = JSON.stringify(req);
}

// 响应转换 - 修改响应体
if (context.responseBody) {
    var resp = JSON.parse(context.responseBody);
    resp.processedBy = "gateway";
    context.responseBody = JSON.stringify(resp);
}

// 添加自定义响应头
context.responseHeaders["X-Gateway-Version"] = "1.0";
