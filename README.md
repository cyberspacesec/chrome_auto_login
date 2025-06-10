# Chrome自动化登录爆破工具

一个基于Chrome浏览器的智能登录页面识别和自动化爆破工具，使用Golang开发。支持批量URL处理、多种验证码检测、智能页面分析等高级功能。

## ✨ 功能特性

### 🎯 核心功能
- 🔍 **智能页面检测**: 基于多维度置信度算法的登录页面识别
- 🎯 **自动元素定位**: 智能识别用户名、密码框、验证码和登录按钮
- 📊 **批量处理**: 支持从文件读取URL列表、用户名和密码字典
- ⚙️ **高度可配置**: 支持自定义识别规则和选择器策略
- 📝 **详细日志**: 完整的操作日志记录和分析报告
- 💾 **结果自动保存**: 成功结果自动保存到文件

### 🛡️ 验证码检测系统
- 🔤 **文字验证码**: 传统图片验证码识别
- 🖼️ **图片验证码**: 基于图像的验证码检测
- 🎯 **滑块验证码**: 拖拽式验证码支持
- 👆 **点击验证码**: 点击选择类验证码
- 🤖 **reCAPTCHA**: Google验证码检测
- 🔒 **hCaptcha**: hCaptcha验证码检测
- 🧠 **行为验证**: 智能行为分析验证码
- 📊 **置信度评分**: 0.0-1.0评分系统确保检测准确性

### 🔧 技术特性
- 🌐 **自动编码检测**: 支持UTF-8、GBK、Big5等多种编码
- 📄 **页面源码输出**: analyze模式下完整页面源码展示
- ⏱️ **智能超时控制**: 各检测阶段独立超时管理
- 🖥️ **可视化调试**: debug模式下浏览器窗口可见
- 🔄 **网络监听**: 获取响应头和网络状态信息
- 📊 **实时进度显示**: 清晰的进度条和状态反馈

### 🚀 OCR配置支持
- 🔧 **多提供商**: Tesseract、百度、阿里云、腾讯云
- 🖼️ **图像预处理**: 缩放、灰度化、降噪、对比度增强
- 🔄 **故障转移**: 主要提供商失败时自动切换备用方案
- ⚙️ **灵活配置**: 支持自定义OCR参数和模型路径

## 📁 项目结构

```
chrome_auto_login/
├── cmd/                    # 主程序入口
│   └── main.go            # 命令行工具主程序
├── pkg/                    # 核心功能包
│   ├── browser/           # Chrome浏览器控制
│   │   └── browser.go     # 浏览器自动化操作
│   ├── config/            # 配置管理
│   │   └── config.go      # 配置文件解析和管理
│   ├── detector/          # 页面和元素检测
│   │   ├── detector.go    # 页面分析和检测器
│   │   └── captcha.go     # 验证码检测引擎
│   └── bruteforce/        # 爆破引擎
│       └── bruteforce.go  # 登录爆破逻辑
├── util/                   # 工具函数
│   └── logger.go          # 日志系统
├── config/                 # 配置文件
│   └── config.yaml        # 主配置文件
├── test/                   # 测试文件
│   ├── test_urls.txt      # URL测试列表
│   ├── test_usernames.txt # 用户名测试字典
│   ├── test_passwords.txt # 密码测试字典
│   ├── test_captcha.html  # 验证码测试页面
│   └── test_no_captcha.html # 无验证码测试页面
├── logs/                   # 日志目录（自动创建）
├── result/                 # 结果输出目录（自动创建）
├── go.mod                  # Go模块定义
├── go.sum                  # 依赖校验文件
├── Makefile               # 构建脚本
└── README.md              # 项目说明
```

## 🚀 安装和编译

### 前置要求

- Go 1.21 或更高版本
- Chrome浏览器
- 操作系统: Windows/Linux/macOS

### 快速开始

```bash
# 克隆项目
git clone https://github.com/cyberspacesec/chrome_auto_login
cd chrome_auto_login

# 下载依赖
go mod tidy

# 编译
go build -o chrome_auto_login cmd/main.go

# 查看帮助
./chrome_auto_login -help
```

### 编译所有平台版本

```bash
# 使用Makefile编译多平台版本
make build-all

# 手动编译不同平台
# Windows 64位
GOOS=windows GOARCH=amd64 go build -o chrome_auto_login.exe cmd/main.go

# Linux 64位
GOOS=linux GOARCH=amd64 go build -o chrome_auto_login_linux cmd/main.go

# macOS 64位
GOOS=darwin GOARCH=amd64 go build -o chrome_auto_login_macos cmd/main.go
```

## 📖 使用方法

### 命令行参数

```bash
用法:
  ./chrome_auto_login -url <目标URL> [选项]
  ./chrome_auto_login -f <URL文件> [选项]

选项:
  -url string        目标登录页面URL (与-f/-file二选一)
  -f string          从文件读取URL列表，一行一个URL (与-url二选一)
  -file string       -f的别名，从文件读取URL列表
  -username string   从文件读取用户名列表，一行一个用户名
  -password string   从文件读取密码列表，一行一个密码
  -path string       Chrome浏览器可执行文件路径（可选，不指定则自动检测）
  -config string     配置文件路径 (默认: config/config.yaml)
  -analyze           仅分析页面，不执行爆破
  -debug             调试模式，显示浏览器窗口和详细操作过程
  -help              显示此帮助信息
```

### 基本使用示例

```bash
# 基本爆破
./chrome_auto_login -url "http://example.com/login"

# 批量URL处理
./chrome_auto_login -f urls.txt

# 使用自定义字典
./chrome_auto_login -url "http://example.com/login" \
  -username users.txt -password passwords.txt

# 页面分析（不执行爆破）
./chrome_auto_login -url "http://example.com/login" -analyze

# 调试模式（显示浏览器窗口）
./chrome_auto_login -url "http://example.com/login" -debug

# 指定Chrome路径
./chrome_auto_login -url "http://example.com/login" \
  -path "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"

# 使用自定义配置文件
./chrome_auto_login -url "http://example.com/login" -config my_config.yaml
```

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

## 📋 文件格式说明

### URL文件格式 (urls.txt)
```
# 支持注释行
http://example1.com/login
https://example2.com/admin/login
https://example3.com/user/signin

# 空行会被自动忽略
https://example4.com/auth
```

### 用户名文件格式 (usernames.txt)
```
# 常用用户名
admin
administrator
root
user
test
guest
# 可以添加更多用户名
manager
operator
```

### 密码文件格式 (passwords.txt)
```
# 常用密码
admin
password
123456
admin123
root
# 可以添加更多密码
password123
admin@123
```

**注意事项:**
- 支持 `#` 和 `//` 开头的注释行
- 自动忽略空行
- 每行一个条目
- 文件编码建议使用UTF-8

## 🔧 调试模式详解

使用 `-debug` 参数可以启用可视化调试模式：

```bash
./chrome_auto_login -url "http://example.com/login" -debug
```

### 调试模式特性
- 🖥️ **浏览器窗口可见**: 观察Chrome的实际操作过程
- 🔍 **详细日志输出**: 显示每个步骤的详细信息
- 🖊️ **表单操作可视化**: 观察用户名、密码填充过程
- 🖱️ **点击操作可视化**: 观察按钮点击效果
- 📊 **实时状态显示**: 显示检测过程和结果
- 🕒 **操作延迟**: 便于观察每个操作步骤

### 适用场景
- 🔍 排查页面元素定位问题
- 📝 调试识别规则和选择器
- 📚 学习工具工作原理
- 🛠️ 开发新功能和规则
- 🎯 验证爆破流程的正确性

## 🛡️ 验证码检测详解

### 检测流程

工具采用**两阶段智能检测机制**：

#### 第一阶段：快速预检测（3秒内）
- 🔍 快速扫描常见验证码选择器
- 📝 检测页面关键词：验证码、captcha、verify等
- ⚡ 无验证码页面几乎瞬间完成

#### 第二阶段：详细类型分析（最多10.5秒）
每种类型检测时间分配（各1.5秒）：
1. 🤖 **reCAPTCHA检测** - Google验证码框架
2. 🔒 **hCaptcha检测** - hCaptcha验证码框架  
3. 🎯 **滑块验证码** - 拖拽滑块类验证码
4. 🖼️ **图片验证码** - 图像选择类验证码
5. 🔤 **文字验证码** - 传统输入框验证码
6. 👆 **点击验证码** - 点击操作类验证码
7. 🧠 **行为验证码** - 智能行为分析类

### 检测结果示例

```bash
🎯 验证码类型分析完成: 文字验证码 (置信度: 0.80)
📋 验证码描述: 文字验证码 - 需要输入验证码字符
🛠️ 处理策略: 可通过OCR识别自动处理
🎯 选择器: input[name*="captcha"]
```

### 支持的验证码类型

| 类型 | 检测特征 | 置信度评估 | 处理策略 |
|------|----------|------------|----------|
| 🔤 文字验证码 | 输入框+图片 | 基于元素匹配 | OCR识别 |
| 🖼️ 图片验证码 | 图片选择 | 关键词+元素 | 图像识别 |
| 🎯 滑块验证码 | 滑块元素 | 特征元素检测 | 拖拽模拟 |
| 👆 点击验证码 | 点击指令 | 文本分析 | 坐标定位 |
| 🤖 reCAPTCHA | iframe框架 | 框架检测 | 专门处理 |
| 🔒 hCaptcha | 特定类名 | 类名匹配 | 专门处理 |
| 🧠 行为验证 | 行为关键词 | 关键词权重 | 行为模拟 |

## ⚙️ 配置文件详解

### 主要配置项

#### 浏览器配置
```yaml
browser:
  headless: true           # 无头模式
  timeout: 60             # 页面加载超时（秒）
  width: 1920             # 浏览器窗口宽度
  height: 1080            # 浏览器窗口高度
  chrome_path: ""         # Chrome路径（可选）
```

#### 验证码检测配置
```yaml
captcha:
  detection:
    enabled: true                    # 启用验证码检测
    timeout: 10                      # 检测超时时间
    confidence_threshold: 0.7        # 置信度阈值
    verbose_output: true             # 详细输出
  
  handling:
    skip_on_detection: true          # 检测到验证码时跳过
    manual_input: false              # 人工输入支持
    ocr_enabled: false               # OCR识别启用
```

#### OCR配置
```yaml
ocr:
  primary_provider: "tesseract"      # 主要OCR提供商
  fallback_providers:               # 备用提供商
    - "local"
    - "tesseract"
  
  preprocessing:                    # 图像预处理
    enabled: true
    resize_enabled: true
    grayscale: true
    noise_reduction: true
    contrast_enhancement: true
```

### 自定义识别规则

#### 用户名输入框选择器
```yaml
form_elements:
  username_selectors:
    - 'input[type="text"][name*="user"]'
    - 'input[type="text"][placeholder*="用户名"]'
    - 'input[type="email"]'
    # 可添加更多自定义选择器
```

#### 验证码元素选择器
```yaml
captcha:
  custom_selectors:
    image_selectors:
      - "img[src*='captcha']"
      - "img[alt*='验证码']"
    input_selectors:
      - "input[placeholder*='验证码']"
      - "input[name*='captcha']"
```

## 📊 页面分析功能

### analyze模式特性

使用 `-analyze` 参数进行页面分析：

```bash
./chrome_auto_login -url "http://example.com/login" -analyze
```

### 分析输出内容

1. **基本页面信息**
   - 页面标题和URL
   - 登录页面置信度（0.0-1.0）
   - 页面编码（UTF-8、GBK等）
   - 分析用时统计

2. **检测到的页面特征**
   - 用户名输入框
   - 密码输入框
   - 验证码相关元素
   - 提交按钮

3. **表单元素详情**
   - 精确的CSS选择器
   - 元素检测置信度
   - 元素可见性状态

4. **验证码检测结果**
   - 验证码类型和描述
   - 检测置信度评分
   - 推荐处理策略
   - 相关选择器信息

5. **完整页面源码**
   - HTML源码完整输出
   - 便于深入分析和调试

6. **网络信息**
   - HTTP响应头
   - 状态码信息
   - 重定向跟踪

### 分析结果示例

```
=== 页面分析结果 ===
页面标题: 管理员登录 - 系统后台
页面URL: http://example.com/admin/login
是否为登录页面: true (置信度: 0.85)
页面编码: utf-8
分析用时: 2.5s

检测到的页面特征:
  • 用户名输入框
  • 密码输入框
  • 文字验证码
  • 提交按钮

检测到的表单元素:
  用户名输入框: input[name="username"]
  密码输入框: input[type="password"]
  验证码输入框: input[name="captcha"]
  提交按钮: button[type="submit"]

验证码检测结果:
  🎯 类型: 文字验证码
  📊 置信度: 0.80
  📋 处理策略: 可通过OCR识别自动处理
  🎯 选择器: input[name="captcha"]
```

## 📂 输出文件说明

### 日志文件
- **位置**: `logs/` 目录
- **格式**: `YYYY-MM-DD.log`
- **内容**: 详细的操作日志和错误信息
- **轮转**: 自动按日期轮转和大小管理

### 结果文件
- **成功结果**: `result/YYYY-MM-DD_success.txt`
- **格式**: `URL:用户名:密码`
- **实时保存**: 成功即保存，避免数据丢失

### 截图文件
- **成功截图**: `success_screenshot.png`
- **内容**: 登录成功后的页面截图
- **用途**: 验证登录成功状态

## 🔒 安全警告

### ⚠️ 重要声明

**此工具仅供授权的安全测试使用，严禁用于非法用途！**

### 合法使用场景
- ✅ 授权的渗透测试
- ✅ 安全研究和学习
- ✅ 企业内部安全评估
- ✅ 漏洞赏金计划
- ✅ 安全意识培训

### 禁止使用场景
- ❌ 未经授权的系统攻击
- ❌ 恶意破坏和入侵
- ❌ 商业机密窃取
- ❌ 个人隐私侵犯
- ❌ 任何违法犯罪活动

### 使用责任
- 使用者需承担所有法律责任
- 确保获得明确的书面授权
- 遵守当地法律法规
- 遵循道德和职业准则

## 🚀 高级功能

### 批量处理能力
- **多URL并发**: 支持批量URL处理
- **自动跳过**: 智能跳过非登录页面
- **进度追踪**: 实时显示处理进度
- **错误处理**: 单个失败不影响整体进程

### 智能检测算法
- **多维度评分**: 综合页面特征评估置信度
- **编码自适应**: 自动检测和转换页面编码
- **超时控制**: 各阶段独立超时防止死锁
- **容错机制**: 网络异常自动重试

### 扩展性设计
- **模块化架构**: 易于添加新功能
- **配置驱动**: 无需修改代码即可调整行为
- **插件友好**: 支持自定义检测规则
- **API接口**: 便于集成到其他工具

## 🛠️ 故障排除

### 常见问题

**Q: Chrome浏览器启动失败**
```bash
# 检查Chrome安装
which google-chrome-stable
# 或指定Chrome路径
./chrome_auto_login -url "..." -path "/path/to/chrome"
```

**Q: 页面加载超时**
```bash
# 增加超时时间（修改config.yaml）
browser:
  timeout: 120  # 增加到120秒
```

**Q: 元素定位失败**
```bash
# 使用调试模式观察
./chrome_auto_login -url "..." -debug
# 使用分析模式查看页面结构
./chrome_auto_login -url "..." -analyze
```

**Q: 验证码检测不准确**
```bash
# 调整置信度阈值（config.yaml）
captcha:
  detection:
    confidence_threshold: 0.5  # 降低阈值
```

**Q: 编码问题导致乱码**
```bash
# 工具会自动检测编码，如仍有问题，检查页面meta标签：
<meta charset="UTF-8">
```

### 调试技巧

1. **使用debug模式**: 观察浏览器实际操作
2. **查看分析结果**: 了解页面检测情况
3. **检查日志文件**: 查看详细错误信息
4. **调整配置参数**: 修改超时和阈值设置
5. **验证选择器**: 使用浏览器开发者工具

## 🤝 贡献指南

### 参与贡献

欢迎提交Issue和Pull Request！

1. **Fork项目** 到你的GitHub
2. **创建特性分支** (`git checkout -b feature/AmazingFeature`)
3. **提交更改** (`git commit -m 'Add some AmazingFeature'`)
4. **推送分支** (`git push origin feature/AmazingFeature`)
5. **创建Pull Request**

### 开发建议

- 遵循Go语言编码规范
- 添加适当的注释和文档
- 编写单元测试
- 更新README文档

### 功能建议

- 新的验证码类型支持
- 更多OCR提供商集成
- 页面识别规则优化
- 性能和稳定性改进

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 📞 联系信息

- **作者**: zhizhuo
- **邮箱**: zhizhuo@cyberspacesec.com
- **组织**: CyberspaceSec
- **项目地址**: https://github.com/cyberspacesec/chrome_auto_login

---

**⚡ 快速开始**

```bash
git clone https://github.com/cyberspacesec/chrome_auto_login
cd chrome_auto_login
go build -o chrome_auto_login cmd/main.go
./chrome_auto_login -url "http://example.com/login" -analyze
```

**🔒 安全提醒**: 仅限授权渗透测试使用，使用者承担所有法律责任。 