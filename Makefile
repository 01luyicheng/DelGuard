# DelGuard 跨平台构建 Makefile

# 项目信息
PROJECT_NAME := delguard
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 构建参数
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -s -w"
GCFLAGS := -gcflags "all=-trimpath=$(PWD)"
ASMFLAGS := -asmflags "all=-trimpath=$(PWD)"

# 构建目录
BUILD_DIR := build
DIST_DIR := dist

# 支持的平台
PLATFORMS := \
	windows/amd64 \
	windows/386 \
	windows/arm64 \
	linux/amd64 \
	linux/386 \
	linux/arm64 \
	linux/arm \
	darwin/amd64 \
	darwin/arm64

.PHONY: all build clean test lint fmt vet deps help install-tools cross-compile package release

# 默认目标
all: clean test build

# 帮助信息
help:
	@echo "DelGuard 构建系统"
	@echo ""
	@echo "可用目标:"
	@echo "  all           - 清理、测试、构建"
	@echo "  build         - 构建当前平台版本"
	@echo "  cross-compile - 交叉编译所有平台"
	@echo "  test          - 运行测试"
	@echo "  lint          - 运行代码检查"
	@echo "  fmt           - 格式化代码"
	@echo "  vet           - 运行 go vet"
	@echo "  clean         - 清理构建文件"
	@echo "  deps          - 下载依赖"
	@echo "  install-tools - 安装开发工具"
	@echo "  package       - 打包发布文件"
	@echo "  release       - 创建发布版本"

# 安装开发工具
install-tools:
	@echo "安装开发工具..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/goreleaser/goreleaser@latest

# 下载依赖
deps:
	@echo "下载依赖..."
	go mod download
	go mod tidy

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 运行 go vet
vet:
	@echo "运行 go vet..."
	go vet ./...

# 代码检查
lint: install-tools
	@echo "运行代码检查..."
	golangci-lint run

# 运行测试
test:
	@echo "运行测试..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 构建当前平台
build: deps fmt vet
	@echo "构建 $(PROJECT_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME) .

# 交叉编译所有平台
cross-compile: deps fmt vet
	@echo "交叉编译所有平台..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GARCH=$${platform#*/} $(MAKE) build-platform PLATFORM=$$platform; \
	done

# 构建特定平台
build-platform:
	@echo "构建 $(PLATFORM)..."
	@GOOS=$(word 1,$(subst /, ,$(PLATFORM))) \
	 GARCH=$(word 2,$(subst /, ,$(PLATFORM))) \
	 OUTPUT_NAME=$(PROJECT_NAME)-$(VERSION)-$(PLATFORM) \
	 $(MAKE) build-single-platform

build-single-platform:
	@if [ "$(GOOS)" = "windows" ]; then \
		OUTPUT_NAME="$(OUTPUT_NAME).exe"; \
	fi; \
	echo "  -> $(GOOS)/$(GARCH): $$OUTPUT_NAME"; \
	CGO_ENABLED=0 GOOS=$(GOOS) GARCH=$(GARCH) \
	go build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) \
	-o $(DIST_DIR)/$$OUTPUT_NAME .

# 打包发布文件
package: cross-compile
	@echo "打包发布文件..."
	@cd $(DIST_DIR) && \
	for file in $(PROJECT_NAME)-$(VERSION)-*; do \
		if [[ $$file == *.exe ]]; then \
			platform=$${file#$(PROJECT_NAME)-$(VERSION)-}; \
			platform=$${platform%.exe}; \
			zip -q $$file.zip $$file; \
		else \
			platform=$${file#$(PROJECT_NAME)-$(VERSION)-}; \
			tar -czf $$file.tar.gz $$file; \
		fi; \
		echo "  -> 已打包: $$file"; \
	done

# 创建发布版本
release: package
	@echo "创建发布版本 $(VERSION)..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --rm-dist; \
	else \
		echo "goreleaser 未安装，跳过自动发布"; \
		echo "手动发布文件位于 $(DIST_DIR)/ 目录"; \
	fi

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@rm -f coverage.out coverage.html

# 安装到系统
install: build
	@echo "安装 $(PROJECT_NAME) 到系统..."
	@if [ "$(shell uname)" = "Darwin" ] || [ "$(shell uname)" = "Linux" ]; then \
		sudo cp $(BUILD_DIR)/$(PROJECT_NAME) /usr/local/bin/; \
		echo "已安装到 /usr/local/bin/$(PROJECT_NAME)"; \
	elif [ "$(OS)" = "Windows_NT" ]; then \
		echo "Windows 安装请运行安装脚本"; \
	fi

# 卸载
uninstall:
	@echo "卸载 $(PROJECT_NAME)..."
	@if [ "$(shell uname)" = "Darwin" ] || [ "$(shell uname)" = "Linux" ]; then \
		sudo rm -f /usr/local/bin/$(PROJECT_NAME); \
		echo "已从 /usr/local/bin/ 卸载"; \
	fi

# 显示版本信息
version:
	@echo "项目: $(PROJECT_NAME)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git提交: $(GIT_COMMIT)"

# 开发模式运行
dev: build
	@echo "开发模式运行..."
	./$(BUILD_DIR)/$(PROJECT_NAME) --help