// 错误处理钩子示例

if (context.error) {
    console.log("错误发生:", context.error);

    // 记录错误信息到data中
    context.data.errorTime = Date.now();
    context.data.errorMessage = String(context.error);

    // 可以在这里添加错误上报逻辑
    console.log("错误已记录，时间:", context.data.errorTime);
}
