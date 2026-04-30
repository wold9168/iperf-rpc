# AGENTS.md - iperf-rpc

## 项目概述

iperf-rpc 是一个基于 Golang 后端的 iperf3 网络测速 RPC 项目。预期运行在 Kubernetes 集群中，用于测试集群间的网络互联效果（pod-to-pod 跨节点/跨集群吞吐量测试）。

核心能力：前端通过 HTTP/RPC 接口操纵后端执行任意 iperf3 指令。两个后端实例之间可通过 iperf3 client/server 模式互相测速。

## 技术栈

| 层 | 技术 | 说明 |
|---|------|------|
| 后端 | Go 1.22+ | gin + swaggo/swag (OpenAPI) |
| 前端 | Vue 3 + Vite | 使用 Axios 调用后端，pnpm 包管理 |
| 打包 | goreleaser | Go 二进制发版 |
| 容器 | Docker (multi-stage) | 最终镜像包含 Go 二进制 + iperf3 + 前端静态资源 |
| 注册表 | ghcr.io | GitHub Container Registry |
| 分支 | master (生产) / dev (开发) | 双分支管理 |

## 项目结构

```
iperf-rpc/
├── AGENTS.md                  # 本文件
├── .github/
│   └── workflows/
│       └── release.yml        # CI: goreleaser + docker build + push to ghcr.io
├── .goreleaser.yaml           # goreleaser 配置
├── Dockerfile                 # 多阶段构建 (go build + iperf3 + 前端)
├── docker-compose.yaml        # 本地开发/测试编排 (两个后端 + 前端)
├── go.mod
├── go.sum
├── main.go                    # 入口
├── cmd/
│   └── server/
│       └── main.go            # 启动逻辑
├── internal/
│   ├── api/
│   │   ├── router.go          # Gin 路由注册
│   │   ├── handler.go         # HTTP Handler
│   │   └── middleware.go      # 中间件 (CORS, logging)
│   ├── iperf/
│   │   ├── executor.go        # iperf3 命令执行器 (封装 os/exec)
│   │   └── executor_test.go
│   ├── model/
│   │   └── types.go           # 请求/响应结构体 + swag 注解
│   └── config/
│       └── config.go          # 配置 (端口, 日志级别等)
├── docs/
│   ├── swagger.json           # swag init 生成
│   └── swagger.yaml
├── web/                       # Vue 3 前端
│   ├── package.json
│   ├── vite.config.js
│   ├── index.html
│   └── src/
│       ├── App.vue
│       ├── main.js
│       ├── api/
│       │   └── index.js       # Axios 封装
│       └── components/
│           ├── IperfForm.vue  # iperf 指令构造与提交
│           ├── ResultView.vue # 测速结果展示
│           └── StatusBar.vue  # 后端连接状态
└── scripts/
    └── swag.sh                # swag init 脚本
```

## API 设计

所有 API 前缀: `/api/v1`

### POST /api/v1/iperf/run

执行一次 iperf3 指令。

Request:
```json
{
  "mode": "client",          // client | server
  "args": {
    "target": "10.0.0.2",    // 客户端模式必填: 目标服务器 IP
    "port": 5201,            // 端口 (默认 5201)
    "duration": 10,          // 测试时长秒数 (默认 10)
    "parallel": 1,           // 并行流数
    "bandwidth": "100M",     // 带宽限制 (如 100M, 1G)
    "protocol": "tcp",       // tcp | udp
    "reverse": false,        // 反向测试 (server 发送)
    "extra": ""              // 额外命令行参数
  }
}
```

Response:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "uuid",
    "command": "iperf3 -c 10.0.0.2 -p 5201 -t 10 -J",
    "output": "...",         // iperf3 JSON 输出原文
    "summary": {             // 解析后的摘要
      "bitrate_bps": 945000000,
      "jitter_ms": 0.5,
      "lost_percent": 0,
      "duration_sec": 10
    },
    "status": "completed",   // running | completed | error
    "started_at": "...",
    "finished_at": "..."
  }
}
```

### POST /api/v1/iperf/server/start

在当前后端启动 iperf3 服务端 (后台运行)。

Response:
```json
{
  "code": 0,
  "message": "server started",
  "data": {
    "port": 5201,
    "pid": 12345
  }
}
```

### POST /api/v1/iperf/server/stop

停止当前后端的 iperf3 服务端。

### GET /api/v1/iperf/status

获取当前后端 iperf3 服务端运行状态。

### GET /api/v1/health

健康检查端点。

### GET /swagger/*

Swagger UI (开发环境)。

## 工作流

### 开发环境 (dev 分支)

```bash
# 启动两个后端 (模拟跨集群两端)
docker compose -f docker-compose.dev.yaml up -d

# 前端开发
cd web && pnpm run dev
```

### 发版流程 (由操作者手动触发)

1. `dev` 分支开发完成并测试通过
2. 合并到 `master`
3. 打 tag (如 `v0.1.0`)
4. push tag 触发 GitHub Actions: goreleaser 构建二进制 + Docker 构建镜像 + push 到 ghcr.io

### goreleaser 配置要点

- 构建目标: `linux/amd64`, `linux/arm64`
- 输出二进制名: `iperf-rpc`
- ldflags 注入版本信息
- 不自动创建 release (需要时手动在 GitHub 创建)

## Docker 镜像

镜像名: `ghcr.io/<owner>/iperf-rpc:<tag>`

### 多阶段构建

1. **go-builder**: 编译 Go 后端
2. **web-builder**: `pnpm run build` 产出 dist
3. **runtime**: `alpine` 基础镜像 + `iperf3` + Go 二进制 + 前端静态文件

运行时 Go 二进制通过 `embed.FS` 内嵌前端 dist，或通过 Gin 静态文件服务挂载 `web/dist`。

## K8s 部署模型

```
# 集群 A
iperf-rpc-server (Deployment, port 5201 + 8080)
iperf-rpc-client (Deployment, port 8080)

# 集群 B  
iperf-rpc-server (Deployment, port 5201 + 8080)
```

- `iperf-rpc-server`: 兼做 API server + iperf3 server (暴露 5201 端口)
- `iperf-rpc-client`: 只做 API server (通过 Service 名连接集群 B 的 server)

前端可部署在任一实例上。

## 开发命令参考

```bash
# swag 文档生成
swag init -g cmd/server/main.go -o docs

# 本地运行
go run cmd/server/main.go

# 测试
go test ./...

# 前端构建
cd web && pnpm run build

# goreleaser 本地测试 (仅构建)
goreleaser build --snapshot --clean

# Docker 构建
docker build -t iperf-rpc:dev .
```

## Commit 规范

- 由操作者手动执行 commit
- 使用 conventional commits: `feat:`, `fix:`, `docs:`, `chore:`
- 分支: `master` (stable) / `dev` (development)
