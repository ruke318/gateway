// 认证钩子示例
// 在认证前后可以修改请求头或添加自定义逻辑

// 在 BeforeAuth 钩子中使用
context.data.authStartTime = Date.now();
console.log("认证开始:", context.requestHeaders);

// 在 AfterAuth 钩子中使用
if (context.data.authStartTime) {
    var duration = Date.now() - context.data.authStartTime;
    console.log("认证耗时:", duration, "ms");
}
