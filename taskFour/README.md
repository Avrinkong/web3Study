# 个人博客系统后端

一个基于 Go、Gin 和 GORM 构建的个人博客系统后端，提供完整的文章管理、用户认证和评论功能。

## 项目特性

- ✅ 用户注册和登录（JWT认证）
- ✅ 文章的完整 CRUD 操作
- ✅ 评论功能
- ✅ 权限控制（用户只能操作自己的资源）
- ✅ Swagger API 文档
- ✅ 完整的错误处理和日志记录
- ✅ SQLite 数据库（无需额外安装数据库服务）

## 技术栈

- **后端框架**: Gin
- **ORM**: GORM
- **数据库**: SQLite
- **认证**: JWT
- **API 文档**: Swagger
- **密码加密**: bcrypt

## 项目结构

```
blog-system/
├── main.go                 # 应用入口文件
├── go.mod                 # Go 模块文件
├── go.sum                 # 依赖校验文件
├── blog.db                # SQLite 数据库文件（自动生成）
├── app.log                # 应用日志文件（自动生成）
├── docs/                  # Swagger 文档（自动生成）
├── config/                # 配置相关
│   ├── database.go
│   └── jwt.go
├── controllers/           # 控制器层
│   ├── auth.go
│   ├── post.go
│   └── comment.go
├── middleware/            # 中间件
│   ├── auth.go
│   ├── logger.go
│   └── error.go
├── models/                # 数据模型
│   ├── user.go
│   ├── post.go
│   └── comment.go
└── README.md              # 项目说明文档
```

## 快速开始

### 环境要求

- Go 1.21 或更高版本

### 安装步骤

1. **克隆项目**
   ```bash
   git clone <项目地址>
   cd blog-system
   ```

2. **安装依赖**
   ```bash
   go mod tidy
   ```

3. **安装 Swagger CLI 工具**
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

4. **生成 Swagger 文档**
   ```bash
   swag init
   ```

5. **运行项目**
   ```bash
   go run main.go
   ```

6. **访问应用**
   - API 服务: http://localhost:8080
   - Swagger 文档: http://localhost:8080/swagger/index.html

## API 接口

### 认证接口

#### 用户注册
- **URL**: `POST /api/auth/register`
- **Body**:
  ```json
  {
    "username": "testuser",
    "password": "password123",
    "email": "test@example.com"
  }
  ```

#### 用户登录
- **URL**: `POST /api/auth/login`
- **Body**:
  ```json
  {
    "username": "testuser",
    "password": "password123"
  }
  ```
- **响应**:
  ```json
  {
    "message": "Login successful",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com"
    }
  }
  ```

### 文章接口

#### 获取文章列表
- **URL**: `GET /api/posts?page=1&limit=10`

#### 获取单篇文章
- **URL**: `GET /api/posts/1`

#### 创建文章（需要认证）
- **URL**: `POST /api/posts`
- **Headers**: `Authorization: Bearer {token}`
- **Body**:
  ```json
  {
    "title": "我的第一篇文章",
    "content": "这是文章的内容..."
  }
  ```

#### 更新文章（需要认证）
- **URL**: `PUT /api/posts/1`
- **Headers**: `Authorization: Bearer {token}`
- **Body**:
  ```json
  {
    "title": "更新后的标题",
    "content": "更新后的内容..."
  }
  ```

#### 删除文章（需要认证）
- **URL**: `DELETE /api/posts/1`
- **Headers**: `Authorization: Bearer {token}`

### 评论接口

#### 创建评论（需要认证）
- **URL**: `POST /api/comments`
- **Headers**: `Authorization: Bearer {token}`
- **Body**:
  ```json
  {
    "content": "这是一条评论",
    "post_id": 1
  }
  ```

#### 获取文章评论
- **URL**: `GET /api/posts/1/comments`

## 测试用例

### 1. 用户注册
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123",
    "email": "test@example.com"
  }'
```

### 2. 用户登录
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

### 3. 创建文章
```bash
curl -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "我的第一篇文章",
    "content": "这是文章的内容..."
  }'
```

### 4. 获取文章列表
```bash
curl -X GET "http://localhost:8080/api/posts?page=1&limit=10"
```

### 5. 创建评论
```bash
curl -X POST http://localhost:8080/api/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "content": "这是一条评论",
    "post_id": 1
  }'
```

## 配置说明

### 环境变量
项目支持以下环境变量配置：

```bash
# JWT 密钥（生产环境必须修改）
export JWT_SECRET=your-super-secret-key-change-in-production

# 服务端口
export PORT=8080
```

### 数据库配置
项目使用 SQLite 数据库，数据库文件 `blog.db` 会在首次运行时自动创建。

## 数据库设计

### Users 表
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | uint | 主键 |
| username | string | 用户名，唯一 |
| password | string | 加密后的密码 |
| email | string | 邮箱，唯一 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

### Posts 表
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | uint | 主键 |
| title | string | 文章标题 |
| content | text | 文章内容 |
| user_id | uint | 用户ID，外键 |
| created_at | time | 创建时间 |
| updated_at | time | 更新时间 |

### Comments 表
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | uint | 主键 |
| content | text | 评论内容 |
| user_id | uint | 用户ID，外键 |
| post_id | uint | 文章ID，外键 |
| created_at | time | 创建时间 |

## 安全特性

- 密码使用 bcrypt 加密存储
- JWT token 认证
- 权限验证（用户只能操作自己的资源）
- 输入参数验证
- SQL 注入防护（使用 GORM）

## 日志系统

应用日志会输出到 `app.log` 文件，包含：
- 请求日志（方法、路径、状态码、响应时间）
- 错误日志
- 系统运行信息

## 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   # 使用其他端口
   export PORT=8081
   go run main.go
   ```

2. **数据库连接失败**
   - 检查当前目录是否有写权限
   - 确保没有其他进程占用数据库文件

3. **Swagger 文档无法访问**
   ```bash
   # 重新生成文档
   swag init
   ```

4. **依赖安装失败**
   ```bash
   # 清理缓存并重新安装
   go clean -modcache
   go mod tidy
   ```

## 许可证

MIT License

---

**开始使用**: 按照上面的安装步骤即可快速启动项目！