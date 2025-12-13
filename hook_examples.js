// ========================================
// 支付宝回调验签 Hook 示例
// Hook Point: BeforeAuth
// ========================================

// 1. 获取请求数据
var body = context.data.request.body;

// 2. 提取签名参数
var sign = body.sign;
var signType = body.sign_type;

// 3. 构造待签名字符串（移除 sign 和 sign_type）
var signParams = {};
for (var key in body) {
    if (key !== 'sign' && key !== 'sign_type') {
        signParams[key] = body[key];
    }
}

// 按 key 排序并拼接
var keys = Object.keys(signParams).sort();
var signStr = keys.map(function(k) {
    return k + '=' + signParams[k];
}).join('&');

console.log('待签名字符串: ' + signStr);

// 4. 验证签名
var publicKey = context.data.org_config.alipay_public_key;

if (!publicKey) {
    throw new Error('未配置支付宝公钥 (alipay_public_key)');
}

var valid = crypto.rsaVerify(signStr, sign, publicKey, 'SHA256');

if (!valid) {
    throw new Error('支付宝签名验证失败');
}

console.log('支付宝签名验证通过');

// ========================================
// 微信回调验签 Hook 示例
// Hook Point: BeforeAuth
// ========================================

// 1. 获取 XML 数据
var xmlBody = context.data.request.body.xml_body;

// 2. 解析 XML（简化版，生产环境应该用 encoding.xmlDecode）
// 提取 sign 字段
var sign = '';
var start = xmlBody.indexOf('<sign><![CDATA[');
if (start !== -1) {
    start += 15;
    var end = xmlBody.indexOf(']]></sign>', start);
    if (end !== -1) {
        sign = xmlBody.substring(start, end);
    }
}

// 3. 构造待签名字符串（移除 sign 字段后的 XML）
// ... 验签逻辑 ...

console.log('微信签名验证通过');

// ========================================
// 自定义响应格式 Hook 示例
// Hook Point: AfterResponseTransform
// ========================================

// 获取内部系统的响应
var response = JSON.parse(context.responseBody);

// 根据内部系统返回的状态码，构造厂商要求的响应格式
if (response.code === 200) {
    // 成功：返回支付宝要求的格式
    context.responseBody = JSON.stringify({
        "code": "SUCCESS"
    });
} else {
    // 失败：返回失败响应
    context.responseBody = JSON.stringify({
        "code": "FAIL",
        "msg": response.message || "处理失败"
    });
}

// ========================================
// 幂等性检查 Hook 示例
// Hook Point: BeforeForward
// ========================================

// 1. 提取业务流水号
var reqData = JSON.parse(context.requestBody);
var orderNo = reqData.order_no;

if (!orderNo) {
    throw new Error('缺少订单号');
}

// 2. 调用 Redis 检查是否已处理
var cacheKey = 'notify:' + orderNo;
var checkUrl = 'http://redis-service/api/exists?key=' + cacheKey;

var checkResp = http.get(checkUrl);
if (checkResp.statusCode === 200) {
    var result = JSON.parse(checkResp.body);
    if (result.exists) {
        console.log('订单已处理过，忽略重复通知: ' + orderNo);
        // 标记为重复，可以在后续逻辑中处理
        context.data.is_duplicate = true;
    }
}

// 3. 标记为已处理（转发成功后会执行）
// 注意：这里只是检查，实际标记应该在 AfterForward Hook 中
console.log('幂等性检查通过: ' + orderNo);

// ========================================
// 记录处理成功 Hook 示例
// Hook Point: AfterForward
// ========================================

// 1. 获取订单号
var reqData = JSON.parse(context.requestBody);
var orderNo = reqData.order_no;

// 2. 如果内部系统返回成功，标记为已处理
if (context.data.response.status === 200) {
    var cacheKey = 'notify:' + orderNo;
    var setUrl = 'http://redis-service/api/set';

    http.post(setUrl, {
        key: cacheKey,
        value: '1',
        ttl: 86400  // 24小时过期
    });

    console.log('标记订单已处理: ' + orderNo);
}

// ========================================
// 动态路由 Hook 示例
// Hook Point: BeforeForward
// ========================================

// 根据订单类型，动态选择转发目标
var reqData = JSON.parse(context.requestBody);
var orderType = reqData.order_type;

if (orderType === 'recharge') {
    // 充值订单 → 充值服务
    context.data.route.backendUrl = 'http://recharge-service:8080';
    context.data.route.backendPath = '/api/recharge/notify';
} else if (orderType === 'purchase') {
    // 购买订单 → 订单服务
    context.data.route.backendUrl = 'http://order-service:8080';
    context.data.route.backendPath = '/api/order/notify';
} else {
    // 默认
    console.log('未知订单类型: ' + orderType);
}

console.log('动态路由: ' + context.data.route.backendUrl + context.data.route.backendPath);
