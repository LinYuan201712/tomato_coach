# 项目运行说明（Tomato Study Room）

本文档涵盖如何在本地或容器中运行后端（Go）和前端（Vue），以及常见配置、测试与部署指引。

**注意**：仓库中不应包含真实密钥。请使用 `backend-go/.env.example` 作为模板，创建本地 `backend-go/.env` 并填入私密值。项目会自动加载此文件中的环境变量。

---

## 先决条件

- Go 1.21+
- Node.js 18+ / npm 或 pnpm
- MySQL 8.0+
- 可选：Docker & Docker Compose

---

## 配置（推荐）

1. 复制并编辑环境变量：

```bash
cd backend-go
cp .env.example .env
# 编辑 .env 填入数据库密码、第三方 API Key 等
```

2. 配置文件说明：
- 项目会优先加载 `.env` 中的环境变量（如 `TOMATO_DB_PASSWORD`）。
- 基础非敏感配置仍保留在 `backend-go/config/config.yaml` 中。

---

## 本地运行后端（开发模式）

1. 进入后端目录：

```bash
cd backend-go
```

2. 下载依赖并整理模块：

```bash
go mod download
go mod tidy
```

3. 运行测试（注意：部分测试可能依赖外部服务或 DB，若本地没有 DB，请有选择性运行）：

```bash
go test ./...    # 可加 -run 指定测试
```

4. 本地运行：

```bash
# 直接运行（程序会自动加载 .env 中的环境变量）
go run ./cmd/main.go

# 或构建二进制并运行
go build -o tomato-backend ./cmd/main.go
./tomato-backend
```

5. 常见检查点：
- 确认 `backend-go/.env` 环境变量配置正确（DB 地址、账号、密码）
- 若使用 Ark / Aliyun / Langfuse 等服务，请先在 `.env` 中填入正确的 Key

---

## 数据库迁移

项目使用 GORM 自动迁移（在 `cmd/main.go` 的 `autoMigrate` 中），默认在应用启动时执行自动迁移。

如果需要手动迁移或初始化数据：

- 在本地 MySQL 中创建数据库，例如 `tomato_study_room`
- 启动应用，应用会创建表

---

## 本地运行前端（开发模式）

1. 进入前端目录：

```bash
cd tamato-frontend-main
```

2. 安装依赖并启动开发服务器：

```bash
npm ci
npm run dev
# 或使用 pnpm / yarn
# pnpm install
# pnpm dev
```

3. 生成生产构建：

```bash
npm run build
# 构建结果在 dist/ 或配置指定目录
```

---

## 使用 Docker 与 Docker Compose

仓库包含 `backend-go/Dockerfile` 与 `docker-compose.yml`（检查根或 `backend-go` 下）。通过 Docker Compose 可以快速启动后端与依赖服务（例如 MySQL、Elasticsearch）：

```bash
# 在项目根或 docker-compose.yml 所在目录
docker compose up --build
# 若使用旧版 docker-compose
docker-compose up --build
```

注意：Docker Compose 会自动读取 `backend-go/.env` 文件并将其注入到容器中。确保你已经创建了该文件。

---

## 生产部署建议

- 禁止在仓库中提交 `backend-go/.env`（真实密钥），使用 `.env.example` 作为模板。
- 使用环境变量或密钥管理服务（AWS Secrets Manager / Azure Key Vault / GitHub Secrets）来管理敏感信息。
- 在 CI 中执行 `go test`、`go vet`、`golangci-lint`（如适用），并构建镜像推到私有镜像仓库。

---

## 常见问题

Q: 启动报错无法连接数据库？

A: 请确认 `backend-go/.env` 中数据库地址、用户名、密码和数据库名正确，且本地 MySQL 已运行并允许远程连接（或容器网络已连通）。

Q: 我不想把密钥写到文件里，怎么办？

A: 使用环境变量或在运行命令前导出：

```bash
export TOMATO_DB_PASSWORD=realpassword
# Windows PowerShell
$env:TOMATO_DB_PASSWORD = "realpassword"
```

并在应用加载配置时优先读取环境变量（项目已支持通过 Viper 加载多种来源）。

---

## 常用命令汇总

```bash
# 后端
cd backend-go
go mod tidy
go test ./...
go run ./cmd/main.go
# 构建
go build -o tomato-backend ./cmd/main.go

# 前端
cd tamato-frontend-main
npm ci
npm run dev
npm run build

# Docker
docker compose up --build
```

---
