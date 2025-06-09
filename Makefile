# Chrome自动化登录爆破工具 Makefile

# 变量定义
BINARY_NAME=chrome_auto_login
MAIN_FILE=cmd/main.go
BUILD_DIR=build
CONFIG_FILE=config/config.yaml

# 默认目标
.PHONY: all
all: clean deps build

# 安装依赖
.PHONY: deps
deps:
	@echo "正在下载依赖..."
	go mod tidy
	go mod download

# 编译程序
.PHONY: build
build:
	@echo "正在编译程序..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "编译完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 编译所有平台版本
.PHONY: build-all
build-all: clean deps
	@echo "正在编译所有平台版本..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	
	@echo "所有平台编译完成"

# 运行程序（需要提供URL参数）
.PHONY: run
run: build
	@if [ -z "$(URL)" ]; then \
		echo "请提供URL参数: make run URL=http://example.com/login"; \
		exit 1; \
	fi
	$(BUILD_DIR)/$(BINARY_NAME) -url "$(URL)"

# 运行调试模式（显示浏览器窗口）
.PHONY: debug
debug: build
	@if [ -z "$(URL)" ]; then \
		echo "请提供URL参数: make debug URL=http://example.com/login"; \
		exit 1; \
	fi
	$(BUILD_DIR)/$(BINARY_NAME) -url "$(URL)" -debug

# 运行页面分析模式
.PHONY: analyze
analyze: build
	@if [ -z "$(URL)" ]; then \
		echo "请提供URL参数: make analyze URL=http://example.com/login"; \
		exit 1; \
	fi
	$(BUILD_DIR)/$(BINARY_NAME) -url "$(URL)" -analyze

# 运行测试
.PHONY: test
test:
	@echo "正在运行测试..."
	go test -v ./test/

# 运行快速测试（跳过网络测试）
.PHONY: test-short
test-short:
	@echo "正在运行快速测试..."
	go test -v -short ./test/

# 运行Chrome连接测试
.PHONY: test-chrome
test-chrome:
	@echo "正在测试Chrome连接..."
	cd test && go test -v -run TestChromeConnection

# 运行网页访问测试
.PHONY: test-web
test-web:
	@echo "正在测试网页访问..."
	cd test && go test -v -run TestBasicWebAccess

# 运行验证码检测测试
.PHONY: test-captcha
test-captcha:
	@echo "正在测试验证码检测..."
	cd test && go test -v -run TestCaptcha

# 代码格式化
.PHONY: fmt
fmt:
	@echo "正在格式化代码..."
	go fmt ./...

# 代码检查
.PHONY: lint
lint:
	@echo "正在进行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过代码检查"; \
		echo "安装方法: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 创建必要的目录
.PHONY: init
init:
	@echo "正在初始化项目目录..."
	@mkdir -p logs
	@mkdir -p $(BUILD_DIR)
	@echo "目录初始化完成"

# 清理编译文件
.PHONY: clean
clean:
	@echo "正在清理编译文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f *.png
	@rm -f logs/*.log
	@echo "清理完成"

# 安装到系统路径
.PHONY: install
install: build
	@echo "正在安装到系统路径..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "安装完成，现在可以在任何地方使用 $(BINARY_NAME) 命令"

# 卸载
.PHONY: uninstall
uninstall:
	@echo "正在卸载..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "卸载完成"

# 显示帮助信息
.PHONY: help
help:
	@echo "Chrome自动化登录爆破工具 - 可用命令:"
	@echo ""
	@echo "  make deps        - 下载项目依赖"
	@echo "  make build       - 编译程序"
	@echo "  make build-all   - 编译所有平台版本"
	@echo "  make run URL=<url> - 运行爆破攻击"
	@echo "  make debug URL=<url> - 调试模式运行（显示浏览器）"
	@echo "  make analyze URL=<url> - 运行页面分析"
	@echo "  make test        - 运行所有测试"
	@echo "  make test-short  - 运行快速测试"
	@echo "  make test-chrome - 测试Chrome连接"
	@echo "  make test-web    - 测试网页访问"
	@echo "  make test-captcha - 测试验证码检测"
	@echo "  make fmt         - 格式化代码"
	@echo "  make lint        - 代码检查"
	@echo "  make init        - 初始化项目目录"
	@echo "  make clean       - 清理编译文件"
	@echo "  make install     - 安装到系统路径"
	@echo "  make uninstall   - 从系统卸载"
	@echo "  make help        - 显示此帮助信息"
	@echo ""
	@echo "示例:"
	@echo "  make run URL=http://example.com/login"
	@echo "  make debug URL=http://example.com/login"
	@echo "  make analyze URL=http://example.com/admin"

# 检查配置文件
.PHONY: check-config
check-config:
	@if [ ! -f $(CONFIG_FILE) ]; then \
		echo "错误: 配置文件 $(CONFIG_FILE) 不存在"; \
		exit 1; \
	fi
	@echo "配置文件检查通过"

# 开发模式（实时编译运行）
.PHONY: dev
dev:
	@if [ -z "$(URL)" ]; then \
		echo "请提供URL参数: make dev URL=http://example.com/login"; \
		exit 1; \
	fi
	@echo "开发模式启动，直接运行源码..."
	go run $(MAIN_FILE) -url "$(URL)" 