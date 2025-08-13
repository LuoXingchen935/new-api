# 手动设置渠道余额功能设计文档

## 功能需求
实现管理员可以手动设置渠道余额的功能，而不是只能通过API自动更新。

## 后端实现

### 1. API路由设计
- 路由路径：`PUT /api/channel/set_balance`
- 请求方法：PUT
- 请求体：
  ```json
  {
    "id": 1,
    "balance": 100.0
  }
  ```
- 响应格式：
  ```json
  {
    "success": true,
    "message": "余额更新成功",
    "data": {
      "id": 1,
      "balance": 100.0
    }
  }
  ```

### 2. 控制器函数设计
在 `controller/channel.go` 中添加新的 `SetChannelBalance` 函数：

```go
func SetChannelBalance(c *gin.Context) {
    var req struct {
        ID      int     `json:"id"`
        Balance float64 `json:"balance"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        common.ApiError(c, err)
        return
    }
    
    // 获取渠道
    channel, err := model.GetChannelById(req.ID, true)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    
    // 更新余额
    channel.UpdateBalance(req.Balance)
    
    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "余额更新成功",
        "data": map[string]interface{}{
            "id":      channel.Id,
            "balance": channel.Balance,
        },
    })
}
```

### 3. 路由注册
在 `router/api-router.go` 中添加新的路由：

```go
channelRoute.PUT("/set_balance", controller.SetChannelBalance)
```

## 前端实现

### 1. 在渠道表格中添加"设置余额"按钮
在 `web/src/components/table/ChannelsTable.js` 中修改余额显示部分，添加设置余额的功能。

### 2. 在编辑渠道页面添加余额输入字段
在 `web/src/pages/Channel/EditChannel.js` 中添加余额输入字段。

## 安全考虑
- 只有管理员可以访问此功能
- 需要验证输入参数的有效性
- 需要确保余额不能设置为负数（如果业务逻辑不允许）

## 测试计划
1. 测试API端点是否正确注册
2. 测试控制器函数是否正确处理请求
3. 测试余额是否正确更新到数据库
4. 测试前端UI是否正确显示和交互