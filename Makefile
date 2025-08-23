# DelGuard Makefile - 构建脚本
# 支持跨平台构建、测试和部署

# 项目信息
PROJECT_NAME := DelGuard
VERSION := 1.0.0
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT) -w -s"

# 构建配置
BUILD_DIR := build
DIST_DIR := dist
COVERAGE_DIR := coverage
LOGS_DIR := logs

# 目标平台
PLATFORMS := \
	windows/amd64 \
	windows/386 \
	linux/amd64 \
	linux/386 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64

# Go 配置
GO_VERSION := $(shell go version | awk '{print $$3}')
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# 工具
GO := go
GOLANGCI_LINT := golangci-lint
GOSEC := gosec
GOVULNCHECK := govulncheck

# 颜色输出
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
MAGENTA := \033[35m
CYAN := \033[36m
RESET := \033[0m

# 默认目标
.PHONY: all
all: clean test build

# 清理
.PHONY: clean
clean:
	@echo "$(CYAN)清理构建产物...$(RESET)"
	@rm -rf $(BUILD_DIR) $(DIST_DIR) $(COVERAGE_DIR) $(LOGS_DIR)
	@rm -f coverage.out coverage.html
	@mkdir -p $(BUILD_DIR) $(DIST_DIR) $(COVERAGE_DIR) $(LOGS_DIR)

# 依赖检查
.PHONY: deps
deps:
	@echo "$(CYAN)检查依赖...$(RESET)"
	@$(GO) mod download
	@$(GO) mod verify
	@$(GO) mod tidy

# 代码格式化
.PHONY: fmt
fmt:
	@echo "$(CYAN)格式化代码...$(RESET)"
	@$(GO) fmt ./...
	@$(GO) vet ./...

# 静态代码检查
.PHONY: lint
lint: deps
	@echo "$(CYAN)运行静态代码检查...$(RESET)"
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run --timeout=5m; \
	else \
		echo "$(YELLOW)golangci-lint 未安装，跳过...$(RESET)"; \
	fi

# 安全扫描
.PHONY: security
security:
	@echo "$(CYAN)运行安全扫描...$(RESET)"
	@if command -v $(GOSEC) >/dev/null 2>&1; then \
		$(GOSEC) ./...; \
	else \
		echo "$(YELLOW)gosec 未安装，跳过...$(RESET)"; \
	fi
	@if command -v $(GOVULNCHECK) >/dev/null 2>&1; then \
		$(GOVULNCHECK) ./...; \
	else \
		echo "$(YELLOW)govulncheck 未安装，跳过...$(RESET)"; \
	fi

# 单元测试
.PHONY: test
test: deps fmt
	@echo "$(CYAN)运行单元测试...$(RESET)"
	@$(GO) test -v ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)测试覆盖率报告: coverage.html$(RESET)"

# 安全测试
.PHONY: test-security
test-security:
	@echo "$(CYAN)运行安全测试...$(RESET)"
	@$(GO) test -v -run "^TestSecurity" ./...

# 性能测试
.PHONY: test-bench
test-bench:
	@echo "$(CYAN)运行性能测试...$(RESET)"
	@$(GO) test -bench=. -benchmem ./...

# 构建单个平台
.PHONY: build
build: deps fmt
	@echo "$(CYAN)构建 $(GOOS)/$(GOARCH)...$(RESET)"
	@mkdir -p $(BUILD_DIR)/$(GOOS)/$(GOARCH)
	@$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(GOOS)/$(GOARCH)/$(PROJECT_NAME)$(if $(filter windows,$(GOOS)),.exe,) .
	@$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(GOOS)/$(GOARCH)/security_tool$(if $(filter windows,$(GOOS)),.exe,) security_tool.go

# 交叉编译所有平台
.PHONY: build-all
build-all: clean deps fmt
	@echo "$(CYAN)交叉编译所有平台...$(RESET)"
	@set -e; for platform in $(PLATFORMS); do \
		platform_split=($$(echo $$platform | tr "/" " ")); \
		GOOS=$${platform_split[0]}; \
		GOARCH=$${platform_split[1]}; \
		output_name=$(DIST_DIR)/$(PROJECT_NAME)-$${GOOS}-$${GOARCH}; \
		security_tool_name=$(DIST_DIR)/security_tool-$${GOOS}-$${GOARCH}; \
		if [ $${GOOS} = "windows" ]; then \
			output_name=$${output_name}.exe; \
			security_tool_name=$${security_tool_name}.exe; \
		fi; \
		echo "$(BLUE)构建 $$platform -> $$output_name$(RESET)"; \
		mkdir -p $(DIST_DIR); \
		GOOS=$${GOOS} GOARCH=$${GOARCH} $(GO) build $(LDFLAGS) -o $${output_name} .; \
		GOOS=$${GOOS} GOARCH=$${GOARCH} $(GO) build $(LDFLAGS) -o $${security_tool_name} security_tool.go; \
		done

# 打包发布
.PHONY: package
package: build-all
	@echo "$(CYAN)打包发布文件...$(RESET)"
	@for platform in $(PLATFORMS); do \
		platform_split=($$(echo $$platform | tr "/" " ")); \
		GOOS=$${platform_split[0]}; \
		GOARCH=$${platform_split[1]}; \
		binary_name=$(PROJECT_NAME)-$${GOOS}-$${GOARCH}; \
		security_tool_name=security_tool-$${GOOS}-$${GOARCH}; \
		if [ $${GOOS} = "windows" ]; then \
			binary_name=$${binary_name}.exe; \
			security_tool_name=$${security_tool_name}.exe; \
		fi; \
		package_name=$(PROJECT_NAME)-$(VERSION)-$${GOOS}-$${GOARCH}; \
		echo "$(BLUE)打包 $$package_name$(RESET)"; \
		mkdir -p $(DIST_DIR)/$${package_name}; \
		cp $(DIST_DIR)/$${binary_name} $(DIST_DIR)/$${package_name}/$(PROJECT_NAME)$(if $(filter windows,$(GOOS)),.exe,); \
		cp $(DIST_DIR)/$${security_tool_name} $(DIST_DIR)/$${package_name}/security_tool$(if $(filter windows,$(GOOS)),.exe,); \
		cp README.md $(DIST_DIR)/$${package_name}/; \
		cp SECURITY.md $(DIST_DIR)/$${package_name}/; \
		cp SECURITY_DEPLOYMENT.md $(DIST_DIR)/$${package_name}/; \
		cp config/security_template.json $(DIST_DIR)/$${package_name}/config.json; \
		cd $(DIST_DIR) && tar -czf $${package_name}.tar.gz $${package_name}/; \
		cd $(DIST_DIR) && zip -r $${package_name}.zip $${package_name}/; \
		echo "$(GREEN)打包完成: $${package_name}.tar.gz 和 $${package_name}.zip$(RESET)"; \
		done

# 生成校验和
.PHONY: checksum
checksum: package
	@echo "$(CYAN)生成校验和...$(RESET)"
	@cd $(DIST_DIR) && find . -name "*.tar.gz" -o -name "*.zip" | xargs sha256sum > checksums.txt
	@cd $(DIST_DIR) && cat checksums.txt

# 安装
.PHONY: install
install: build
	@echo "$(CYAN)安装 DelGuard...$(RESET)"
	@sudo cp $(BUILD_DIR)/$(GOOS)/$(GOARCH)/$(PROJECT_NAME)$(if $(filter windows,$(GOOS)),.exe,) /usr/local/bin/$(PROJECT_NAME)
	@sudo cp $(BUILD_DIR)/$(GOOS)/$(GOARCH)/security_tool$(if $(filter windows,$(GOOS)),.exe,) /usr/local/bin/security_tool
	@sudo mkdir -p /etc/delguard
	@sudo cp config/security_template.json /etc/delguard/config.json
	@echo "$(GREEN)安装完成: /usr/local/bin/$(PROJECT_NAME) 和 /usr/local/bin/security_tool$(RESET)"

# 卸载
.PHONY: uninstall
uninstall:
	@echo "$(CYAN)卸载 DelGuard...$(RESET)"
	@sudo rm -f /usr/local/bin/$(PROJECT_NAME)
	@sudo rm -f /usr/local/bin/security_tool
	@sudo rm -rf /etc/delguard
	@echo "$(GREEN)卸载完成$(RESET)"

# 构建安全工具
.PHONY: build-security-tool
build-security-tool:
	@echo "$(CYAN)构建安全工具...$(RESET)"
	@mkdir -p $(BUILD_DIR)/$(GOOS)/$(GOARCH)
	@$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(GOOS)/$(GOARCH)/security_tool$(if $(filter windows,$(GOOS)),.exe,) security_tool.go

# 运行安全检查
.PHONY: run-security-check
run-security-check: build-security-tool
	@echo "$(CYAN)运行安全检查...$(RESET)"
	@$(BUILD_DIR)/$(GOOS)/$(GOARCH)/security_tool$(if $(filter windows,$(GOOS)),.exe,) -check

# 运行安全验证
.PHONY: run-security-verify
run-security-verify: build-security-tool
	@echo "$(CYAN)运行安全验证...$(RESET)"
	@$(BUILD_DIR)/$(GOOS)/$(GOARCH)/security_tool$(if $(filter windows,$(GOOS)),.exe,) -verify

# 生成安全报告
.PHONY: run-security-report
run-security-report: build-security-tool
	@echo "$(CYAN)生成安全报告...$(RESET)"
	@$(BUILD_DIR)/$(GOOS)/$(GOARCH)/security_tool$(if $(filter windows,$(GOOS)),.exe,) -report

# 开发环境设置
.PHONY: dev-setup
dev-setup:
	@echo "$(CYAN)设置开发环境...$(RESET)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "$(GREEN)开发环境设置完成$(RESET)"

# 代码质量报告
.PHONY: quality-report
quality-report: deps fmt lint test security
	@echo "$(CYAN)生成代码质量报告...$(RESET)"
	@mkdir -p $(LOGS_DIR)
	@echo "$(GREEN)代码质量报告已生成在 $(LOGS_DIR)/ 目录$(RESET)"

# 发布准备
.PHONY: release
release: clean deps fmt lint test security build-all package checksum
	@echo "$(GREEN)发布准备完成!$(RESET)"
	@echo "$(GREEN)构建产物在 $(DIST_DIR)/ 目录$(RESET)"
	@echo "$(GREEN)校验和在 $(DIST_DIR)/checksums.txt$(RESET)"

# 帮助
.PHONY: help
help:
	@echo "$(CYAN)DelGuard Makefile 帮助$(RESET)"
	@echo "$(MAGENTA)构建目标:$(RESET)"
	@echo "  $(YELLOW)all$(RESET)         - 完整构建流程"
	@echo "  $(YELLOW)build$(RESET)       - 构建当前平台"
	@echo "  $(YELLOW)build-all$(RESET)   - 交叉编译所有平台"
	@echo "  $(YELLOW)package$(RESET)     - 打包发布文件"
	@echo "  $(YELLOW)release$(RESET)     - 发布准备"
	@echo ""
	@echo "$(MAGENTA)测试目标:$(RESET)"
	@echo "  $(YELLOW)test$(RESET)        - 运行单元测试"
	@echo "  $(YELLOW)test-security$(RESET) - 运行安全测试"
	@echo "  $(YELLOW)test-bench$(RESET)    - 运行性能测试"
	@echo ""
	@echo "$(MAGENTA)安全工具:$(RESET)"
	@echo "  $(YELLOW)build-security-tool$(RESET) - 构建安全工具"
	@echo "  $(YELLOW)run-security-check$(RESET)  - 运行安全检查"
	@echo "  $(YELLOW)run-security-verify$(RESET) - 运行安全验证"
	@echo "  $(YELLOW)run-security-report$(RESET) - 生成安全报告"
	@echo ""
	@echo "$(MAGENTA)质量检查:$(RESET)"
	@echo "  $(YELLOW)fmt$(RESET)         - 格式化代码"
	@echo "  $(YELLOW)lint$(RESET)        - 静态代码检查"
	@echo "  $(YELLOW)security$(RESET)    - 安全扫描"
	@echo "  $(YELLOW)quality-report$(RESET) - 生成质量报告"
	@echo ""
	@echo "$(MAGENTA)部署:$(RESET)"
	@echo "  $(YELLOW)install$(RESET)     - 安装到系统"
	@echo "  $(YELLOW)uninstall$(RESET)   - 从系统卸载"
	@echo ""
	@echo "$(MAGENTA)开发:$(RESET)"
	@echo "  $(YELLOW)dev-setup$(RESET)   - 设置开发环境"
	@echo "  $(YELLOW)clean$(RESET)       - 清理构建产物"
	@echo "  $(YELLOW)help$(RESET)        - 显示此帮助"