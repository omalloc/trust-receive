# Trust-Receive

Trust-Receive is a high-performance cache file consistency verification system designed as a global authoritative center. It verifies whether cache file information reported by edge nodes (CDN edges, cache servers, etc.) is consistent with the authoritative global state. When inconsistency is detected, the system can trigger real-time alerts and provide a basis for decision-making for subsequent global push deletions of incorrect resources.

## 🚀 Core Features

- **Real-time Consistency Verification**: Receives URL, file Hash, file size (Content-Length), and last modified time (Last-Modified) reported by edge nodes, and compares them with authoritative records.
- **High-performance Deduplication Optimization**: Built-in **Bloom Filter** to quickly filter known records in massive reported data, significantly reducing Redis query pressure.
- **Multi-dimensional Consistency Determination**: Combines `URL + Last-Modified` as a version identifier to perform precise comparisons of Hash and file size for the same version.
- **Alerting & Linkage**: Immediately triggers alerting logic when data inconsistency (e.g., Hash conflict) is detected, supporting extended integration with DingTalk, Lark, Email, or custom cleanup scripts.
- **Dual Protocol Support**: Provides both **gRPC** and **HTTP** interfaces for easy access by clients in different environments.

## 🛠 Technology Stack

- **Framework**: [Kratos v2](https://go-kratos.dev/) (Go microservices framework)
- **Storage**: Redis (for persisting file fingerprint information)
- **Algorithm**: Bloom Filter (for high-performance deduplication)
- **Communication**: gRPC, HTTP/JSON (Protobuf definition)
- **Build**: Wire (dependency injection), Docker, Makefile

## 📁 Project Structure

```text
.
├── api/                # Interface definitions (Protobuf)
├── cmd/                # Program entry points
├── configs/            # Configuration files
├── internal/
│   ├── biz/           # Business logic layer (Consistency verification, Bloom Filter)
│   ├── data/          # Data access layer (Redis R/W)
│   ├── server/        # HTTP/gRPC server configuration
│   └── service/       # API interface implementation
├── Dockerfile          # Containerization deployment script
└── Makefile            # Automated build tool
```

## 🚥 Quick Start

### Prerequisites
- Go 1.19+
- Redis 6.x+
- Protoc & Kratos toolchain

### Installation
```bash
# Clone the project
git clone https://github.com/omalloc/trust-receive.git
cd trust-receive

# Install dependency tools (first time)
make init

# Generate code (Protobuf, Wire)
make all
```

### Configuration
Edit `configs/config.yaml` to configure your Redis address and service ports:
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

### Running
```bash
# Start locally
kratos run

# Docker deployment
docker build -t trust-receive .
docker run -p 8000:8000 -p 9000:9000 -v $(pwd)/configs:/data/conf trust-receive
```

## 📖 API Description

### 1. File Information Reporting and Verification
- **Path**: `POST /receive`
- **Protocol**: HTTP / gRPC
- **Request Parameters**:
  - `url`: File resource URL
  - `hash`: File HASH value
  - `lm`: Last-Modified time
  - `cl`: Content-Length (File size)

- **Logic Description**:
  1. The system generates a unique version Key based on `url` + `lm`.
  2. If it's the first time this version is reported, its `hash` and `cl` are stored in Redis.
  3. If the version already exists, the current reported `hash` and `cl` are compared with the stored values.
  4. If inconsistent, an error is returned and an alert is triggered.

## 🔔 Alerting and Extension
Business logic is located in `internal/biz/alerter.go`. You can integrate here according to your needs:
- DingTalk/Lark Webhook alerts.
- Call console APIs for global deletion/cleanup of incorrect cache resources.
- Write to Prometheus metrics for dashboard display.

## 📄 License
[MIT](LICENSE)

