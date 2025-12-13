#!/bin/bash
# 回调功能测试脚本

BASE_URL="http://localhost:8080"

echo "========================================="
echo "Gateway 回调功能测试"
echo "========================================="
echo ""

# 测试 1: 数字渠道 JSON 回调
echo "测试 1: 数字渠道 JSON 回调"
echo "-----------------------------------------"
curl -X POST "${BASE_URL}/gateway/v1/notify/payNotify/1" \
  -H "Content-Type: application/json" \
  -d '{
    "order_no": "ORDER001",
    "amount": 100,
    "status": "success"
  }'
echo -e "\n"

# 测试 2: 数字渠道 Form 回调
echo "测试 2: 数字渠道 Form 回调"
echo "-----------------------------------------"
curl -X POST "${BASE_URL}/gateway/v1/notify/payNotify/2" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "order_no=ORDER002&amount=200&status=success"
echo -e "\n"

# 测试 3: 数字渠道 GET 回调
echo "测试 3: 数字渠道 GET 回调"
echo "-----------------------------------------"
curl -X GET "${BASE_URL}/gateway/v1/notify/refundNotify/1?order_no=ORDER003&amount=50&status=success"
echo -e "\n"

# 测试 4: 支付宝回调（Form 格式）
echo "测试 4: 支付宝回调"
echo "-----------------------------------------"
curl -X POST "${BASE_URL}/gateway/v1/notify/payNotify/alipay" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "out_trade_no=ORDER004&trade_no=2024XXX&total_amount=100&trade_status=TRADE_SUCCESS&notify_time=2024-01-01%2012:00:00&sign=test_sign&sign_type=RSA2"
echo -e "\n"

# 测试 5: 微信回调（XML 格式模拟）
echo "测试 5: 微信回调"
echo "-----------------------------------------"
curl -X POST "${BASE_URL}/gateway/v1/notify/payNotify/wechat" \
  -H "Content-Type: application/xml" \
  -d '<xml>
<out_trade_no><![CDATA[ORDER005]]></out_trade_no>
<transaction_id><![CDATA[WX2024XXX]]></transaction_id>
<total_fee>100</total_fee>
<result_code><![CDATA[SUCCESS]]></result_code>
</xml>'
echo -e "\n"

echo "========================================="
echo "测试完成"
echo "========================================="
