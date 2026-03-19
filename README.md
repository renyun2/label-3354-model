# Go API Starter

> 基于 Gin + GORM + Redis 的 Golang RESTful API 启动套件，内置微信小程序开发支持。

## 技术栈

| 组件 | 说明 |
|------|------|
| [Gin](https://github.com/gin-gonic/gin) | HTTP 框架 |
| [GORM](https://gorm.io/) | ORM 框架（MySQL） |
| [go-redis](https://github.com/redis/go-redis) | Redis 客户端 |
| [Viper](https://github.com/spf13/viper) | 配置管理 |
| [Zap](https://go.uber.org/zap) | 高性能日志 |
| [golang-jwt](https://github.com/golang-jwt/jwt) | JWT 认证 |
| [golang.org/x/time](https://pkg.go.dev/golang.org/x/time/rate) | 请求限流 |

## 项目结构

```
.
├── cmd/
│   └── server/main.go          # 程序入口，依赖注入
├── internal/
│   ├── config/config.go        # 配置结构体定义
│   ├── handler/                # HTTP 处理器层
│   │   ├── user.go             # 用户相关接口
│   │   ├── article.go          # 文章 CRUD 接口（示例）
│   │   └── wechat.go           # 微信消息接口
│   ├── middleware/             # 中间件
│   │   ├── auth.go             # JWT 鉴权
│   │   ├── cors.go             # CORS 跨域
│   │   ├── logger.go           # 请求日志 + TraceID
│   │   └── ratelimit.go        # IP 速率限制
│   ├── model/                  # 数据模型（GORM）
│   │   ├── base.go             # 基础模型（ID/时间/软删除）
│   │   ├── user.go             # 用户模型
│   │   └── article.go          # 文章模型
│   ├── repository/             # 数据访问层
│   │   ├── base.go             # 泛型基础仓库
│   │   ├── user.go             # 用户仓库
│   │   └── article.go          # 文章仓库
│   ├── router/router.go        # 路由注册
│   └── service/                # 业务逻辑层
│       ├── user.go             # 用户服务（含微信登录）
│       ├── article.go          # 文章服务
│       └── wechat.go           # 微信服务（消息推送）
├── pkg/
│   ├── cache/redis.go          # Redis 封装
│   ├── database/database.go    # GORM 初始化
│   ├── jwt/jwt.go              # JWT 工具
│   ├── logger/logger.go        # Zap 日志封装
│   ├── response/response.go    # 统一响应格式
│   └── wechat/                 # 微信 SDK
│       ├── auth.go             # 登录、手机号
│       ├── message.go          # 订阅消息、客服消息
│       └── client.go           # 客户端聚合
├── configs/
│   ├── config.yaml             # 运行配置
│   └── config.example.yaml     # 配置示例
├── migrations/init.sql         # 数据库初始化 SQL
└── .air.toml                   # 热重载配置
```

---

## 快速开始

### 本地开发运行

**前置条件**：Go 1.22+、MySQL 8.0+、Redis 7+

```bash
# 1. 安装依赖
go mod download

# 2. 复制并修改配置
cp configs/config.example.yaml configs/config.yaml
# 编辑 config.yaml，修改数据库/Redis连接信息和微信AppID等

# 3. 运行
go run ./cmd/server -config ./configs/config.yaml

# 或编译后运行
make build && make run

# 热重载开发（需安装 air）
go install github.com/air-verse/air@latest
make dev
```

---

## 配置说明

编辑 `configs/config.yaml`，关键配置项：

```yaml
wechat:
  mini_program:
    app_id: "wx_your_app_id"       # 微信小程序 AppID
    app_secret: "your_app_secret"  # 微信小程序 AppSecret

jwt:
  secret: "your-secret-key"   # 请在生产环境使用强随机密钥
  expire_hours: 72            # Token 有效期（小时）
```

> 环境变量可覆盖配置（格式：`APP_DATABASE_PASSWORD=xxx`）。

---

## API 文档

### 统一响应格式

```json
{
  "code": 0,
  "message": "成功",
  "data": {},
  "trace_id": "1709123456789-a1b2c3d4"
}
```

| code | 含义 |
|------|------|
| 0 | 成功 |
| 400 | 参数错误 |
| 401 | 未授权 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
| 10001 | 微信授权失败 |
| 20001 | 用户不存在 |

---

## curl 测试示例

### 1. 健康检查

```bash
curl http://localhost:8080/health
```

响应：
```json
{"status":"ok","version":"1.0.0","service":"go-api-starter"}
```

---

### 2. 微信小程序登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/wechat-login \
  -H "Content-Type: application/json" \
  -d '{
    "code": "微信小程序wx.login()返回的code",
    "nickname": "测试用户",
    "avatar_url": "https://example.com/avatar.jpg"
  }'
```

响应：
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "user": {
      "id": 1,
      "open_id": "oXxxx...",
      "nickname": "测试用户",
      "role": "user",
      "created_at": "2024-01-01T12:00:00+08:00"
    },
    "token": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "expires_in": 1709382400
    }
  }
}
```

---

### 3. 刷新 Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}'
```

---

### 4. 获取当前用户信息

```bash
# 将 TOKEN 替换为登录返回的 access_token
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN"
```

---

### 5. 更新用户资料

```bash
curl -X PUT http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nickname": "新昵称",
    "avatar_url": "https://example.com/new-avatar.jpg",
    "gender": 1
  }'
```

---

### 6. 绑定手机号

```bash
curl -X POST http://localhost:8080/api/v1/user/bind-phone \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000"}'
```

---

### 7. 文章列表（公开接口）

```bash
curl "http://localhost:8080/api/v1/articles?page=1&page_size=10"
```

响应：
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "list": [],
    "total": 0,
    "page": 1,
    "page_size": 10
  }
}
```

---

### 8. 创建文章

```bash
curl -X POST http://localhost:8080/api/v1/articles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "我的第一篇文章",
    "content": "Hello World！这是文章内容。",
    "status": 1
  }'
```

---

### 9. 获取文章详情

```bash
curl http://localhost:8080/api/v1/articles/1
```

---

### 10. 更新文章

```bash
curl -X PUT http://localhost:8080/api/v1/articles/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "更新后的标题", "status": 1}'
```

---

### 11. 删除文章

```bash
curl -X DELETE http://localhost:8080/api/v1/articles/1 \
  -H "Authorization: Bearer $TOKEN"
```

---

### 12. 发送微信订阅消息（管理员权限）

```bash
# 需要 role=admin 的 Token
ADMIN_TOKEN="..."

curl -X POST http://localhost:8080/api/v1/wechat/message/subscribe \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "open_id": "oXxxx...",
    "template_id": "your-template-id",
    "page": "pages/index/index",
    "data": {
      "thing1": {"value": "您的订单已发货"},
      "time2":  {"value": "2024-01-01 12:00:00"},
      "thing3": {"value": "顺丰快递"}
    }
  }'
```

---

### 13. 发送客服消息（管理员权限）

```bash
curl -X POST http://localhost:8080/api/v1/wechat/message/customer \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "open_id": "oXxxx...",
    "content": "您好，您的问题已处理完毕。"
  }'
```

---

## 微信小程序前端对接

### 登录流程

```javascript
// 1. 调用 wx.login 获取 code
wx.login({
  success(res) {
    // 2. 将 code 发送到后端换取 token
    wx.request({
      url: 'https://your-api.com/api/v1/auth/wechat-login',
      method: 'POST',
      data: {
        code: res.code,
        nickname: '用户昵称',
        avatar_url: '头像URL'
      },
      success(loginRes) {
        const { access_token, refresh_token } = loginRes.data.data.token;
        wx.setStorageSync('access_token', access_token);
        wx.setStorageSync('refresh_token', refresh_token);
      }
    });
  }
});

// 3. 后续请求携带 Token
wx.request({
  url: 'https://your-api.com/api/v1/user/profile',
  header: {
    'Authorization': `Bearer ${wx.getStorageSync('access_token')}`
  },
  success(res) {
    console.log(res.data);
  }
});
```

---

## 扩展指南

### 添加新的业务模块

1. **定义模型** `internal/model/xxx.go`
2. **创建仓库** `internal/repository/xxx.go`（继承 `BaseRepository[T]`）
3. **编写服务** `internal/service/xxx.go`
4. **编写处理器** `internal/handler/xxx.go`
5. **注册路由** 在 `internal/router/router.go` 中添加路由
6. **注入依赖** 在 `cmd/server/main.go` 中完成初始化

### 环境变量覆盖配置

```bash
# 格式：APP_<配置路径大写，以_分隔>
export APP_DATABASE_PASSWORD=prod_password
export APP_JWT_SECRET=prod_jwt_secret
export APP_WECHAT_MINI_PROGRAM_APP_SECRET=prod_wechat_secret
```

---

## License

MIT
