# Trust-Receive

Trust-Receive 是一个高性能的缓存文件一致性校验系统，旨在作为全网权威中心，校验边缘节点（CDN 边缘、缓存服务器等）上报的缓存文件信息是否与全网状态一致。当检测到不一致时，系统可触发实时告警，并为后续全网推送删除错误资源提供决策依据。

## 🚀 核心功能

- **实时一致性校验**：接收边缘节点上报的 URL、文件 Hash、文件大小（Content-Length）及最后修改时间（Last-Modified），并与权威记录进行比对。
- **高性能去重优化**：内置 **布隆过滤器 (Bloom Filter)**，在海量上报数据中快速过滤已知记录，大幅减轻 Redis 的查询压力。
- **多维度一致性判定**：结合 `URL + Last-Modified` 作为版本标识，针对同一版本的 Hash 和文件大小进行精确比对。
- **告警与联动**：检测到数据不一致（如 Hash 冲突）时，立即触发告警逻辑，支持扩展对接钉钉、飞书、邮件或自定义清理脚本。
- **双协议支持**：同时提供 **gRPC** 和 **HTTP** 接口，方便不同环境的客户端接入。

## 🛠 技术栈

- **框架**: [Kratos v2](https://go-kratos.dev/) (Go 微服务框架)
- **存储**: Redis (用于持久化文件指纹信息)
- **算法**: Bloom Filter (用于高性能查重)
- **通信**: gRPC, HTTP/JSON (Protobuf 定义)
- **构建**: Wire (依赖注入), Docker, Makefile

## 📁 项目结构

```text
.
├── api/                # 接口定义 (Protobuf)
├── cmd/                # 程序入口
├── configs/            # 配置文件
├── internal/
│   ├── biz/           # 业务逻辑层 (一致性验证、布隆过滤器)
│   ├── data/          # 数据访问层 (Redis 读写)
│   ├── server/        # HTTP/gRPC 服务配置
│   └── service/       # API 接口实现
├── Dockerfile          # 容器化部署脚本
└── Makefile            # 自动化构建工具
```

## 🚥 快速开始

### 环境依赖
- Go 1.19+
- Redis 6.x+
- Protoc & Kratos 工具链

### 安装
```bash
# 克隆项目
git clone https://github.com/omalloc/trust-receive.git
cd trust-receive

# 安装依赖工具 (首次运行)
make init

# 生成代码 (Protobuf, Wire)
make all
```

### 配置
编辑 `configs/config.yaml`，配置您的 Redis 地址及服务端口：
```yaml
server:
  http:
    addr: 0.0.0.0:8000
  grpc:
    addr: 0.0.0.0:9000
data:
  redis:
    addrs: ["127.0.0.1:6379"]
    password: "your_password"
```

### 运行
```bash
# 本地启动
kratos run

# Docker 部署
docker build -t trust-receive .
docker run -p 8000:8000 -p 9000:9000 -v $(pwd)/configs:/data/conf trust-receive
```

## 📖 API 说明

### 1. 文件信息上报与校验
- **路径**: `POST /receive`
- **协议**: HTTP / gRPC
- **请求参数**:
  - `url`: 文件资源 URL
  - `hash`: 文件 HASH 值
  - `lm`: Last-Modified (最后修改时间)
  - `cl`: Content-Length (文件大小)

- **逻辑说明**:
  1. 系统根据 `url` + `lm` 生成唯一版本 Key。
  2. 若该版本首次上报，则将其 `hash` 和 `cl` 存入 Redis。
  3. 若该版本已存在，则比对当前上报的 `hash` 和 `cl` 是否与存储值一致。
  4. 不一致则返回错误并触发告警。

## 🔔 告警与扩展
业务逻辑位于 `internal/biz/alerter.go`。您可以根据需求在此处集成：
- 钉钉/飞书 Webhook 告警。
- 调用控制台 API 全网删除/清理错误缓存资源。
- 写入 Prometheus 指标进行监控盘展示。

## 📄 License
[MIT](LICENSE)

