.PHONY: help build run test clean docker-build docker-up docker-down lint fmt

# 帮助信息
help:
	@echo "Tomato Study Room - Go Backend"
	@echo ""
	@echo "可用命令:"
	@echo "  make build          - 编译应用"
	@echo "  make run            - 运行应用（开发模式）"
	@echo "  make test           - 运行测试"
	@echo "  make test-coverage  - 运行测试并生成覆盖率报告"
	@echo "  make clean          - 清理编译产物"
	@echo "  make lint           - 代码静态检查"
	@echo "  make fmt            - 代码格式化"
	@echo "  make docker-build   - 构建Docker镜像"
	@echo "  make docker-up      - 启动Docker容器"
	@echo "  make docker-down    - 停止Docker容器"
	@echo "  make deps           - 下载依赖"

# 编译应用
build:
	@echo "编译应用..."
	go build -o bin/tomato-backend ./cmd

# 运行应用
run: build
	@echo "运行应用..."
	./bin/tomato-backend

# 运行开发模式（自动重启）
run-dev:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	@echo "启动开发模式..."
	air

# 运行测试
test:
	@echo "运行测试..."
	go test -v ./...

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "运行测试并生成覆盖率报告..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 清理编译产物
clean:
	@echo "清理编译产物..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# 代码静态检查
lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "运行代码静态检查..."
	golangci-lint run ./...

# 代码格式化
fmt:
	@echo "代码格式化..."
	go fmt ./...

# 下载依赖
deps:
	@echo "下载依赖..."
	go mod download
	go mod tidy

# 构建Docker镜像
docker-build:
	@echo "构建Docker镜像..."
	docker build -t tomato-backend:latest .

# 启动Docker容器
docker-up:
	@echo "启动Docker容器..."
	docker-compose up -d

# 停止Docker容器
docker-down:
	@echo "停止Docker容器..."
	docker-compose down

# 查看日志
docker-logs:
	docker-compose logs -f backend

# 数据库迁移
db-migrate:
	@echo "执行数据库迁移..."
	go run ./cmd migrate

# 生成mock
mock:
	@which mockgen > /dev/null || go install github.com/golang/mock/mockgen@latest
	@echo "生成mock文件..."
	mockgen -source=internal/repository/repository.go -destination=tests/mocks/mock_repository.go -package=mocks
	mockgen -source=internal/service/auth.go -destination=tests/mocks/mock_service.go -package=mocks

# 代码审查
review: fmt lint test

# 完整的开发流程
dev: clean deps fmt lint test build run-dev

# 所有依赖的版本
versions:
	@echo "Go版本:"
	@go version
	@echo ""
	@echo "主要依赖:"
	@grep -E "github.com/(gin-gonic|gorm|golang-jwt|spf13)" go.mod
