# Chrome自动化登录爆破工具

一个基于Chrome浏览器的智能登录页面识别和自动化爆破工具，使用Golang开发。

## 功能特性

- 🔍 **智能页面识别**: 基于正则表达式自动识别登录页面
- 🎯 **自动元素定位**: 智能识别用户名、密码框、验证码和登录按钮
- ⚙️ **高度可配置**: 支持自定义识别规则和选择器策略
- 📝 **详细日志**: 完整的操作日志记录和分析报告
- 📊 **实时进度显示**: 清晰的进度条和状态反馈
- 💾 **结果自动保存**: 成功结果自动保存到文件
- ⚡ **智能停止机制**: 爆破成功后自动停止
- 📄 **日志轮转管理**: 自动管理日志文件大小和数量
- 🔧 **灵活扩展**: 模块化设计，易于扩展和自定义

## 项目结构

```
chrome_auto_login/
├── cmd/                    # 主程序入口
│   └── main.go
├── pkg/                    # 核心功能包
│   ├── browser/           # Chrome浏览器控制
│   ├── config/            # 配置管理
│   ├── detector/          # 页面和元素检测
│   └── bruteforce/        # 爆破引擎
├── util/                   # 工具函数
│   └── logger.go
├── config/                 # 配置文件
│   └── config.yaml
├── go.mod                  # Go模块定义
└── README.md              # 项目说明
```

## 安装和编译

### 前置要求

- Go 1.21 或更高版本
- Chrome浏览器
- 操作系统: Windows/Linux/macOS

### 编译

```bash
# 克隆项目
git clone https://github.com/cyberspacesec/chrome_auto_login
cd chrome_auto_login

# 下载依赖
go mod tidy

# 编译
go build -o chrome_auto_login cmd/main.go
```

## 使用方法

### 基本用法

```bash
# 对目标URL进行爆破
./chrome_auto_login -url "http://example.com/login"

# 仅分析页面，不执行爆破
./chrome_auto_login -url "http://example.com/login" -analyze

# 使用自定义配置文件
./chrome_auto_login -url "http://example.com/login" -config "my_config.yaml"

# 显示帮助信息
./chrome_auto_login -help
```

### 命令行选项

- `-url`: 目标登录页面URL (必需)
- `-config`: 配置文件路径 (默认: config/config.yaml)
- `-analyze`: 仅分析页面，不执行爆破
- `-debug`: 调试模式，显示浏览器窗口和详细操作过程
- `-help`: 显示帮助信息

### 使用Makefile快捷命令

```bash
# 运行爆破
make run URL=http://example.com/login

# 调试模式（显示浏览器窗口）
make debug URL=http://example.com/login

# 页面分析
make analyze URL=http://example.com/login

# 运行测试
make test-chrome                      # 测试Chrome连接
make test-web                         # 测试网页访问

# 编译相关
make build                            # 编译程序
make build-all                        # 编译所有平台版本
```

## 调试模式

使用 `-debug` 参数或 `make debug` 命令可以启用调试模式：

```bash
# 命令行调试
./chrome_auto_login -url "http://example.com/login" -debug

# Makefile调试
make debug URL=http://example.com/login
```

调试模式特性：
- 🖥️ **显示浏览器窗口**: 可以看到Chrome浏览器的实际操作过程
- 🔍 **详细日志输出**: 显示每个操作步骤的详细信息
- 🖊️ **表单操作可视化**: 可以看到用户名、密码的填充过程
- 🖱️ **点击操作可视化**: 可以看到按钮点击的实际效果

调试模式适用于：
- 开发和调试新的识别规则
- 排查页面元素定位问题
- 观察登录流程的实际执行过程
- 学习和理解工具的工作原理

## 验证码检测功能

🛡️ **智能验证码检测系统**，支持多种验证码类型的自动识别和分析：

### 支持的验证码类型

| 类型 | 描述 | 检测方式 | 处理策略 |
|------|------|----------|----------|
| 🔤 文字验证码 | 传统图片验证码 | 选择器匹配 | OCR识别或人工输入 |
| 🖼️ 图片验证码 | 图片形式验证码 | 图片元素检测 | OCR识别或人工输入 |
| 🎯 滑块验证码 | 拖拽滑块验证 | 特征元素检测 | 模拟拖拽操作 |
| 👆 点击验证码 | 点击选择验证 | 关键词检测 | 图像识别和点击 |
| 🤖 Google reCAPTCHA | Google验证码 | 框架检测 | 需要专门处理 |
| 🔒 hCaptcha | hCaptcha验证码 | 框架检测 | 需要专门处理 |
| 🧠 行为验证 | 智能行为分析 | 关键词检测 | 模拟正常行为 |

### 验证码检测示例

```bash
# 分析页面的验证码情况
./chrome_auto_login -url "http://example.com/login" -analyze

# 输出示例：
# 验证码检测结果:
#   🎯 类型: 滑块验证码
#   📊 置信度: 0.90
#   📋 处理策略: 需要模拟滑块拖拽操作
#   🎯 选择器: .geetest_slider
```

## 配置文件说明

配置文件 `config/config.yaml` 包含以下主要部分：

### 1. 浏览器配置

```yaml
browser:
  headless: false    # 是否无头模式
  timeout: 30        # 页面加载超时时间(秒)
  width: 1920        # 浏览器窗口宽度
  height: 1080       # 浏览器窗口高度
```

### 2. 登录页面识别规则

```yaml
login_page_detection:
  title_patterns:     # 页面标题正则表达式
    - "(?i).*login.*"
    - "(?i).*登录.*"
  url_patterns:       # 页面URL正则表达式
    - "(?i).*login.*"
    - "(?i).*signin.*"
  content_keywords:   # 页面内容关键词
    - "用户名"
    - "密码"
    - "登录"
```

### 3. 表单元素识别规则

```yaml
form_elements:
  username_selectors:  # 用户名输入框选择器
    - 'input[type="text"][name*="user"]'
    - 'input[type="text"][placeholder*="用户名"]'
  password_selectors:  # 密码输入框选择器
    - 'input[type="password"]'
    - 'input[name*="pass"]'
  submit_selectors:    # 提交按钮选择器
    - 'input[type="submit"]'
    - 'button[type="submit"]'
```

### 4. 爆破配置

```yaml
bruteforce:
  usernames:          # 用户名字典
    - "admin"
    - "administrator"
  passwords:          # 密码字典
    - "admin"
    - "password"
    - "123456"
  delay: 2            # 爆破间隔时间(秒)
  max_retries: 3      # 最大重试次数
```

### 5. 验证码处理配置

```yaml
captcha:
  # 验证码检测配置
  detection:
    enabled: true                    # 是否启用验证码检测
    timeout: 10                      # 检测超时时间(秒)
    confidence_threshold: 0.7        # 检测置信度阈值
    
  # 验证码处理策略
  handling:
    skip_on_detection: true          # 检测到验证码时是否跳过
    manual_input: false              # 是否支持人工输入
    ocr_enabled: false               # 是否启用OCR识别
    
  # 自定义验证码选择器
  custom_selectors:
    image_selectors:                 # 验证码图片选择器
      - "img[src*='captcha']"
      - "img[alt*='验证码']"
    input_selectors:                 # 验证码输入框选择器  
      - "input[placeholder*='验证码']"
      - "input[name*='captcha']"
    slider_selectors:                # 滑块选择器
      - ".slider-captcha"
      - ".geetest_slider"
```

## 自定义识别规则

### 添加新的选择器

1. 编辑 `config/config.yaml`
2. 在相应的 `*_selectors` 数组中添加新的CSS选择器
3. 选择器支持属性匹配、类名匹配、ID匹配等

示例：
```yaml
username_selectors:
  - 'input[name="loginname"]'          # 精确匹配name属性
  - 'input[id^="user"]'                # ID以"user"开头
  - 'input.username-input'             # 类名匹配
  - '#username-field'                  # ID选择器
```

### 添加页面识别规则

```yaml
login_page_detection:
  title_patterns:
    - "(?i).*管理员.*"                 # 管理员页面
    - "(?i).*system.*login.*"          # 系统登录页面
  url_patterns:
    - "(?i).*/admin/.*"                # 管理员路径
    - "(?i).*/backend/.*"              # 后台路径
```

### 自定义用户名密码字典

```yaml
bruteforce:
  usernames:
    - "admin"
    - "root"
    - "administrator"
    - "sa"
    - "guest"
  passwords:
    - "admin"
    - "password"
    - "123456"
    - "admin123"
    - "root"
```

## 使用示例

### 示例1: 分析页面

```bash
./chrome_auto_login -url "http://example.com/admin/login.php" -analyze
```

输出结果：
```
=== 页面分析结果 ===
页面标题: 管理员登录
页面URL: http://example.com/admin/login.php
是否为登录页面: true
检测到的表单元素:
  username: input[name="username"]
  password: input[type="password"]
  submit: button[type="submit"]
页面特征:
  • 包含用户名字段
  • 包含密码字段
  • 包含登录文本
```

### 示例2: 执行爆破

```bash
./chrome_auto_login -url "http://example.com/admin/login.php"
```

成功情况下的输出：
```
=== 开始自动化爆破 ===
🎉 爆破成功!
用户名: admin
密码: admin123
目标URL: http://example.com/admin/dashboard.php
成功截图已保存: success_screenshot.png
```

## 技术实现

### 核心技术栈

- **Chrome DevTools Protocol**: 通过chromedp库控制Chrome浏览器
- **正则表达式**: 用于页面内容模式匹配
- **CSS选择器**: 用于DOM元素定位
- **YAML配置**: 灵活的配置管理

### 关键算法

1. **页面识别算法**: 综合标题、URL、内容多维度判断
2. **元素定位算法**: 优先级匹配，从最精确到最通用
3. **登录成功判断**: URL跳转 + 页面内容分析

### 安全特性

- 支持爆破间隔设置，避免被WAF检测
- 自动检测验证码，跳过无法处理的情况
- 详细的日志记录，便于审计

## 注意事项

⚠️ **重要提醒**:
- 本工具仅用于**授权的安全测试**
- 请勿用于非法攻击他人系统
- 使用前请确保获得目标系统所有者的明确授权
- 遵守当地法律法规和职业道德

## 常见问题

### Q: 工具无法识别登录页面？
A: 检查配置文件中的识别规则，添加更多的页面特征模式。

### Q: 找不到登录元素？
A: 使用浏览器开发者工具查看实际的HTML结构，添加对应的CSS选择器。

### Q: 爆破速度太快被封IP？
A: 增加配置文件中的delay值，添加更长的间隔时间。

### Q: Chrome浏览器启动失败？
A: 确保已正确安装Chrome浏览器，检查系统权限。

## 扩展开发

### 添加新的识别策略

1. 修改 `pkg/detector/detector.go`
2. 在 `IsLoginPage` 方法中添加新的检测逻辑
3. 更新配置文件结构

### 添加新的爆破策略

1. 修改 `pkg/bruteforce/bruteforce.go`
2. 实现新的登录成功判断逻辑
3. 添加新的错误处理机制

## 日志和结果管理

### 日志文件管理

工具支持自动日志管理，配置在 `config.yaml` 中：

```yaml
logging:
  file_management:
    save_to_file: true                    # 是否保存到文件
    log_dir: "logs"                       # 日志目录
    filename_format: "2006-01-02.log"     # 文件名格式（按日期）
    max_files: 0                          # 保留文件数量（0=全部保留）
    max_size: 100                         # 单文件最大大小(MB)
    rotate_by_date: true                  # 是否按日期轮转
```

### 结果自动保存

成功的爆破结果会自动保存到 `result` 目录：

```yaml
results:
  save_dir: "result"                            # 结果保存目录
  success_filename_format: "2006-01-02_success.txt"  # 成功结果文件名
  format: "url:username:password"               # 保存格式
  realtime_save: true                           # 实时保存
```

### 文件结构示例

```
chrome_auto_login/
├── logs/                          # 日志文件
│   ├── 2025-01-15.log            # 按日期的日志文件
│   └── 2025-01-16.log
└── result/                        # 结果文件
    ├── 2025-01-15_success.txt     # 成功结果
    └── 2025-01-16_success.txt
```

### 成功结果格式

```
http://example.com/login:admin:password123
http://test.com/admin:root:admin
```

## 进度条和状态显示

### 实时进度显示
- 🔓 爆破进度条始终显示在最底部
- 📊 实时更新尝试次数、成功率、剩余时间
- 🎯 当前正在尝试的用户名/密码组合

### 智能停止机制
- ⚡ 爆破成功后立即停止，不继续尝试其他组合
- 💾 成功结果立即保存到文件
- 📈 显示详细的爆破摘要报告

### 输出示例

```
🔓 爆破进度 [██████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 12/96 (12.5%) 剩余: 14m32s - 尝试 admin:test123

🎉 [成功] admin/admin123 - 登录成功！

📊 爆破摘要报告
总耗时: 2m45s
总尝试: 12 次
成功: 1 次
失败: 11 次
成功率: 8.3%
平均速度: 0.1 次/秒
```

## 版本更新

- v1.0.0: 初始版本，基础功能实现
- v1.1.0: 新增进度条优化、日志管理、结果保存功能
- 更多功能正在开发中...

## 贡献指南

欢迎提交Issue和Pull Request来改进此工具。

## 许可证

本项目仅供学习和授权测试使用。 