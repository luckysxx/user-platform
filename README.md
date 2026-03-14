# user-platform

一个基于 Go 的 User 中台服务，提供用户注册、登录、Token 刷新能力，同时暴露 HTTP 和 gRPC 两种协议。

## 功能特性

- 用户注册（用户名 + 密码）
- 用户登录（签发 Access Token + Refresh Token）
- Refresh Token 轮换刷新
- HTTP API（Gin）
- gRPC API（grpc-go）
- PostgreSQL（Ent）持久化
- Redis（Refresh Token 存储）
- Docker Compose 一键启动（基础设施 + 服务）

## 技术栈

- Go 1.25
- Gin
- gRPC + Protobuf
- Ent ORM
- PostgreSQL
- Redis
- Zap 日志
- Docker / Docker Compose

## 项目结构（核心）

- cmd/http: HTTP 服务入口
- cmd/grpc: gRPC 服务入口
- internal/service: 业务逻辑
- internal/repository: 数据访问
- internal/transport/http: HTTP 路由与 Handler
- internal/transport/grpc: gRPC Server
- internal/platform/config: 配置加载
- proto/user: gRPC 协议定义

## 快速开始

### 1. 环境准备

- Go >= 1.25
- Docker + Docker Compose
- protoc（如果需要重新生成 protobuf）

### 2. 配置环境变量

复制配置文件：

cp .env.example .env

建议将 .env 中数据库连接改为：

DB_SOURCE=postgres://luckys:123456@localhost:5432/user_platform?sslmode=disable

默认常用配置项：

- SERVER_PORT=8081
- USER_GRPC_PORT=9091（仅 gRPC 入口读取）
- REDIS_ADDR=localhost:6379
- REDIS_PASSWORD=123456
- JWT_SECRET=自定义强随机密钥

### 3. 启动基础设施（本地开发）

make local-infra-up

这会启动：

- PostgreSQL（5432）
- Redis（6379）

### 4. 启动服务

启动 HTTP：

make local-run-http

启动 gRPC：

make local-run-grpc

也可以一键启动全部容器（基础设施 + 服务）：

make docker-up

停止：

make docker-down

## HTTP API

Base URL:

http://localhost:8081/api/v1

接口列表：

- POST /users/register
- POST /users/login
- POST /users/refresh

### 注册示例

curl -X POST 'http://localhost:8081/api/v1/users/register' \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "alice123",
    "password": "Password123"
  }'

### 登录示例

注意：当前代码中的登录 DTO 将 access_token、refresh_token 也标记为了必填字段，建议先传占位值。

curl -X POST 'http://localhost:8081/api/v1/users/login' \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "alice123",
    "password": "Password123",
    "access_token": "placeholder",
    "refresh_token": "placeholder"
  }'

### 刷新 Token 示例

curl -X POST 'http://localhost:8081/api/v1/users/refresh' \
  -H 'Content-Type: application/json' \
  -d '{
    "token": "<refresh_token>"
  }'

### 响应格式

成功响应：

{
  "code": 200,
  "msg": "success",
  "data": { ... }
}

业务错误通常也是 HTTP 200，具体失败码在 body.code 中。

## gRPC API

地址：

localhost:9091

服务定义：

- user.UserService/Register
- user.AuthService/Login
- user.AuthService/RefreshToken

使用 grpcurl 示例（明文）：

grpcurl -plaintext -d '{"username":"alice123","password":"Password123"}' \
  localhost:9091 user.UserService/Register

grpcurl -plaintext -d '{"username":"alice123","password":"Password123"}' \
  localhost:9091 user.AuthService/Login

grpcurl -plaintext -d '{"token":"<refresh_token>"}' \
  localhost:9091 user.AuthService/RefreshToken

## 常用 Make 命令

- make local-infra-up: 启动 postgres + redis
- make local-infra-down: 停止 postgres + redis
- make local-run-http: 本地启动 HTTP
- make local-run-grpc: 本地启动 gRPC
- make local-test: 运行测试
- make proto-gen: 生成 protobuf 代码
- make docker-up: 启动全部容器
- make docker-down: 停止并清理容器
- make docker-logs: 查看服务日志
- make health: 查看健康状态

## 数据库说明

当前服务启动流程未自动执行 Ent Schema Migration。若你使用全新数据库，请先确保存在 users 表。

可参考最小建表 SQL：

CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR(32) UNIQUE NOT NULL,
  password TEXT NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'active',
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

## 可观测组件（Docker）

基础设施编排中包含：

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000（默认账号/密码见 docker-compose-infra.yaml）
- Loki: http://localhost:3100

## 已知注意项

- HTTP 登录 DTO 当前要求 access_token/refresh_token 必填（即使业务层未使用），后续可在 DTO 校验标签中移除。

## License

仅用于学习与内部开发。