package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cyberspacesec/chrome_auto_login/pkg/browser"
	"github.com/cyberspacesec/chrome_auto_login/pkg/bruteforce"
	"github.com/cyberspacesec/chrome_auto_login/pkg/config"
	"github.com/cyberspacesec/chrome_auto_login/pkg/detector"
	"github.com/cyberspacesec/chrome_auto_login/util"
)

const (
	Version = "0.0.1"
	Author  = "zhizhuo"
)

// showWelcomeBanner 显示漂亮的欢迎页面
func showWelcomeBanner() {
	// ASCII艺术标题
	banner := `
╔═══════════════════════════════════════════════════════════════════════════════╗
║                                                                               ║
║    ╔═╗╦ ╦╦═╗╔═╗╔╦╗╔═╗  ╔═╗╦ ╦╔╦╗╔═╗  ╦  ╔═╗╔═╗╦╔╗╔                            ║
║    ║  ╠═╣╠╦╝║ ║║║║║╣   ╠═╣║ ║ ║ ║ ║  ║  ║ ║║ ╦║║║║                            ║
║    ╚═╝╩ ╩╩╚═╚═╝╩ ╩╚═╝  ╩ ╩╚═╝ ╩ ╚═╝  ╩═╝╚═╝╚═╝╩╝╚╝                            ║
║                                                                               ║
║    🔐 智能登录页面爆破工具 | 🌐 自动化安全测试平台                            ║
║                                                                               ║
║    ┌─────────────────────────────────────────────────────────────────────┐    ║
║    │  🎯 版本: %-10s    👨‍💻 作者: %-20s               │    ║
║    │  📅 时间: %-19s    🏢 组织: CyberspaceSec             │    ║
║    └─────────────────────────────────────────────────────────────────────┘    ║
║                                                                               ║
║    ✨ 功能特性:                                                               ║
║       🔍 智能页面检测    🎯 多种登录方式识别    🚀 高效爆破引擎               ║
║       📊 实时进度显示    💾 结果自动保存        🛡️ 验证码自动处理              ║
║       🔧 灵活配置系统    📝 详细日志记录        🌈 美观界面显示               ║
║                                                                               ║
║    ⚠️  免责声明: 此工具仅用于授权安全测试，请勿用于非法用途                    ║
║                                                                               ║
╚═══════════════════════════════════════════════════════════════════════════════╝`

	// 获取当前时间
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// 清屏并显示欢迎信息
	fmt.Print("\033[2J\033[H") // 清屏并移动到左上角

	// 打印带有动态信息的横幅
	formattedBanner := fmt.Sprintf(banner, Version, Author, currentTime)

	// 添加颜色效果
	fmt.Print("\033[1;36m") // 青色粗体
	fmt.Println(formattedBanner)
	fmt.Print("\033[0m") // 重置颜色

	// 添加加载动画效果
	fmt.Print("\n🚀 正在初始化系统")
	for i := 0; i < 3; i++ {
		time.Sleep(300 * time.Millisecond)
		fmt.Print(".")
	}
	fmt.Println(" 完成!")

	// 系统信息
	fmt.Println("\n" + strings.Repeat("─", 80))
	fmt.Printf("🔧 系统配置检查...\n")
	fmt.Printf("   ✅ Go运行环境正常\n")
	fmt.Printf("   ✅ Chrome浏览器支持\n")
	fmt.Printf("   ✅ 网络连接就绪\n")
	fmt.Printf("   ✅ 配置文件加载\n")
	fmt.Println(strings.Repeat("─", 80))
}

func main() {
	// 命令行参数
	var (
		configFile   = flag.String("config", "config/config.yaml", "配置文件路径")
		targetURL    = flag.String("url", "", "目标登录页面URL")
		urlFile      = flag.String("f", "", "从文件读取URL列表，一行一个URL")
		fileAlias    = flag.String("file", "", "从文件读取URL列表（-f的别名）")
		usernameFile = flag.String("username", "", "从文件读取用户名列表，一行一个用户名")
		passwordFile = flag.String("password", "", "从文件读取密码列表，一行一个密码")
		chromePath   = flag.String("path", "", "Chrome浏览器可执行文件路径（可选，不指定则自动检测）")
		analyze      = flag.Bool("analyze", false, "仅分析页面，不执行爆破")
		debug        = flag.Bool("debug", false, "调试模式，显示浏览器窗口和详细操作过程")
		help         = flag.Bool("help", false, "显示帮助信息")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// 处理file参数的别名
	if *fileAlias != "" && *urlFile == "" {
		*urlFile = *fileAlias
	}

	// 验证参数
	if *targetURL == "" && *urlFile == "" {
		fmt.Println("错误: 必须指定目标URL (-url) 或URL文件 (-f/-file)")
		fmt.Println("使用 -help 查看帮助信息")
		os.Exit(1)
	}

	// 加载配置
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
		os.Exit(1)
	}

	// 如果指定了Chrome路径，设置到配置中
	if *chromePath != "" {
		cfg.Browser.ChromePath = *chromePath
		fmt.Printf("✅ 使用指定的Chrome路径: %s\n", *chromePath)
	}

	// 从文件加载用户名和密码（如果指定）
	if *usernameFile != "" {
		usernames, err := readFileLines(*usernameFile)
		if err != nil {
			fmt.Printf("读取用户名文件失败: %v\n", err)
			os.Exit(1)
		}
		cfg.Bruteforce.Usernames = usernames
		fmt.Printf("✅ 从文件加载了 %d 个用户名\n", len(usernames))
	}

	if *passwordFile != "" {
		passwords, err := readFileLines(*passwordFile)
		if err != nil {
			fmt.Printf("读取密码文件失败: %v\n", err)
			os.Exit(1)
		}
		cfg.Bruteforce.Passwords = passwords
		fmt.Printf("✅ 从文件加载了 %d 个密码\n", len(passwords))
	}

	// 如果启用debug模式，强制使用可视化浏览器和详细日志
	if *debug {
		cfg.Browser.Headless = false
		cfg.Logging.Level = "debug"
		fmt.Println("🔍 调试模式已启用：浏览器窗口可见，显示详细操作过程")
	} else {
		// 非debug模式时，确保不显示debug信息
		if cfg.Logging.Level == "debug" {
			cfg.Logging.Level = "error"
		}
	}

	// 初始化日志器
	logConfig := util.LogConfig{
		Level:          cfg.Logging.Level,
		SaveToFile:     cfg.Logging.FileManagement.SaveToFile,
		LogDir:         cfg.Logging.FileManagement.LogDir,
		FilenameFormat: cfg.Logging.FileManagement.FilenameFormat,
		MaxFiles:       cfg.Logging.FileManagement.MaxFiles,
		MaxSize:        cfg.Logging.FileManagement.MaxSize,
		RotateByDate:   cfg.Logging.FileManagement.RotateByDate,
	}

	err = util.InitLogger(logConfig)
	if err != nil {
		fmt.Printf("初始化日志器失败: %v\n", err)
		os.Exit(1)
	}

	// 确保程序退出时关闭日志文件
	defer util.CloseLogger()

	// 显示欢迎页面
	showWelcomeBanner()

	// 创建浏览器实例
	browserInstance := browser.NewBrowser(cfg, util.Logger)

	// 启动浏览器
	util.LogInfo("启动Chrome浏览器...")
	if err := browserInstance.Start(); err != nil {
		util.LogError(fmt.Sprintf("启动浏览器失败: %v", err))
		os.Exit(1)
	}
	defer browserInstance.Close()

	// 创建页面检测器
	pageDetector := detector.NewPageDetector(browserInstance, cfg, util.Logger)

	// 获取URL列表
	urls := []string{}
	if *urlFile != "" {
		fileUrls, err := readFileLines(*urlFile)
		if err != nil {
			util.LogError(fmt.Sprintf("读取URL文件失败: %v", err))
			os.Exit(1)
		}
		urls = fileUrls
		fmt.Printf("✅ 从文件加载了 %d 个URL\n", len(urls))
	} else {
		urls = []string{*targetURL}
	}

	// 处理每个URL
	for i, url := range urls {
		if len(urls) > 1 {
			fmt.Printf("\n" + strings.Repeat("=", 70))
			fmt.Printf("\n🎯 处理第 %d/%d 个URL: %s\n", i+1, len(urls), url)
			fmt.Println(strings.Repeat("=", 70))
		}

		// 导航到目标URL
		if err := browserInstance.NavigateTo(url); err != nil {
			util.LogError(fmt.Sprintf("导航到目标URL失败: %v", err))
			continue
		}

		// 如果只是分析模式
		if *analyze {
			util.LogInfo("=== 页面分析模式 ===")

			// 分析页面
			analysis, err := pageDetector.AnalyzePage()
			if err != nil {
				util.LogError(fmt.Sprintf("页面分析失败: %v", err))
				continue
			}

			// 输出分析结果
			printAnalysisResult(analysis)
			continue
		}

		// 创建状态显示器和进度感知日志器
		statusDisplay := util.NewStatusDisplay()
		progressLogger := util.NewProgressAwareLogger(statusDisplay)

		// 创建爆破引擎
		bruteforceEngine := bruteforce.NewBruteForceEngine(browserInstance, pageDetector, cfg, progressLogger)

		// 显示爆破信息
		if len(urls) == 1 {
			fmt.Println("\n" + strings.Repeat("=", 70))
			fmt.Println("🎯 自动化登录爆破即将开始")
			fmt.Println(strings.Repeat("=", 70))
		}

		// 执行爆破
		result, err := bruteforceEngine.ExecuteBruteForce(url)
		if err != nil {
			util.LogError(fmt.Sprintf("爆破执行失败: %v", err))
			continue
		}

		// 输出结果
		printBruteForceResult(result)
	}
}

// readFileLines 从文件中读取行，去除空行和注释
func readFileLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释行
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "//") {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func showHelp() {
	// 打印带有版本信息的标题
	fmt.Printf("\n╔═══════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                  🔐 Chrome自动化登录爆破工具                           ║\n")
	fmt.Printf("║                     版本: %-10s | 作者: %-15s       ║\n", Version, Author)
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Println("用法:")
	fmt.Println("  ./chrome_auto_login -url <目标URL> [选项]")
	fmt.Println("  ./chrome_auto_login -f <URL文件> [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -url string        目标登录页面URL (与-f/-file二选一)")
	fmt.Println("  -f string          从文件读取URL列表，一行一个URL (与-url二选一)")
	fmt.Println("  -file string       -f的别名，从文件读取URL列表")
	fmt.Println("  -username string   从文件读取用户名列表，一行一个用户名")
	fmt.Println("  -password string   从文件读取密码列表，一行一个密码")
	fmt.Println("  -path string       Chrome浏览器可执行文件路径（可选，不指定则自动检测）")
	fmt.Println("  -config string     配置文件路径 (默认: config/config.yaml)")
	fmt.Println("  -analyze           仅分析页面，不执行爆破")
	fmt.Println("  -debug             调试模式，显示浏览器窗口和详细操作过程")
	fmt.Println("  -help              显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 基本用法")
	fmt.Println("  ./chrome_auto_login -url \"http://example.com/login\"")
	fmt.Println()
	fmt.Println("  # 从文件读取URL列表")
	fmt.Println("  ./chrome_auto_login -f urls.txt")
	fmt.Println()
	fmt.Println("  # 使用自定义用户名和密码字典")
	fmt.Println("  ./chrome_auto_login -url \"http://example.com/login\" -username users.txt -password passwords.txt")
	fmt.Println()
	fmt.Println("  # 调试模式（显示浏览器窗口）")
	fmt.Println("  ./chrome_auto_login -url \"http://example.com/login\" -debug")
	fmt.Println()
	fmt.Println("  # 使用自定义配置文件")
	fmt.Println("  ./chrome_auto_login -url \"http://example.com/login\" -config \"my_config.yaml\"")
	fmt.Println()
	fmt.Println("  # 指定Chrome浏览器路径")
	fmt.Println("  ./chrome_auto_login -url \"http://example.com/login\" -path \"/path/to/chrome\"")
	fmt.Println()
	fmt.Println("  # 仅分析页面")
	fmt.Println("  ./chrome_auto_login -url \"http://example.com/login\" -analyze")
	fmt.Println()
	fmt.Println("文件格式:")
	fmt.Println("  - URL文件: 每行一个URL")
	fmt.Println("  - 用户名文件: 每行一个用户名")
	fmt.Println("  - 密码文件: 每行一个密码")
	fmt.Println("  - 支持 # 和 // 开头的注释行")
	fmt.Println("  - 自动忽略空行")
	fmt.Println()
	fmt.Println("注意:")
	fmt.Println("  - 此工具仅用于授权测试，请勿用于非法用途")
	fmt.Println("  - 确保已安装Chrome浏览器")
	fmt.Println("  - 配置文件中可以自定义识别规则和字典")
	fmt.Println("  - 支持OCR验证码识别和多种验证码类型检测")
	fmt.Println()
	fmt.Println("🌐 项目地址: https://github.com/cyberspacesec/chrome_auto_login")
	fmt.Println("📧 联系作者: zhizhuo@cyberspacesec.com")
	fmt.Println("🔒 安全提醒: 仅限授权渗透测试使用")
	fmt.Println()
}

func printAnalysisResult(analysis *detector.PageAnalysis) {
	util.LogInfo("=== 页面分析结果 ===")
	util.LogInfo(fmt.Sprintf("页面标题: %s", analysis.Title))
	util.LogInfo(fmt.Sprintf("页面URL: %s", analysis.URL))
	util.LogInfo(fmt.Sprintf("是否为登录页面: %t (置信度: %.2f)", analysis.IsLogin, analysis.Confidence))
	util.LogInfo(fmt.Sprintf("页面编码: %s", analysis.Encoding))
	util.LogInfo(fmt.Sprintf("分析用时: %v", analysis.LoadTime))

	// 显示响应头信息
	if len(analysis.ResponseHeaders) > 0 {
		util.LogInfo("响应头信息:")
		for key, value := range analysis.ResponseHeaders {
			util.LogInfo(fmt.Sprintf("  %s: %s", key, value))
		}
	}

	// 显示检测到的特征
	if len(analysis.DetectedFeatures) > 0 {
		util.LogInfo("检测到的页面特征:")
		for _, feature := range analysis.DetectedFeatures {
			util.LogInfo(fmt.Sprintf("  • %s", feature))
		}
	}

	// 显示表单元素
	if analysis.FormElements != nil {
		util.LogInfo("检测到的表单元素:")
		if analysis.FormElements.UsernameSelector != "" {
			util.LogInfo(fmt.Sprintf("  用户名输入框: %s", analysis.FormElements.UsernameSelector))
		}
		if analysis.FormElements.PasswordSelector != "" {
			util.LogInfo(fmt.Sprintf("  密码输入框: %s", analysis.FormElements.PasswordSelector))
		}
		if analysis.FormElements.CaptchaSelector != "" {
			util.LogInfo(fmt.Sprintf("  验证码输入框: %s", analysis.FormElements.CaptchaSelector))
		}
		if analysis.FormElements.CheckboxSelector != "" {
			util.LogInfo(fmt.Sprintf("  用户协议复选框: %s", analysis.FormElements.CheckboxSelector))
		}
		if analysis.FormElements.SubmitSelector != "" {
			util.LogInfo(fmt.Sprintf("  提交按钮: %s", analysis.FormElements.SubmitSelector))
		}

		// 显示验证码检测结果
		if analysis.FormElements.HasCaptcha && analysis.FormElements.CaptchaInfo != nil {
			captcha := analysis.FormElements.CaptchaInfo
			util.LogInfo("验证码检测结果:")
			util.LogInfo(fmt.Sprintf("  🎯 类型: %s", captcha.GetTypeName()))
			util.LogInfo(fmt.Sprintf("  📊 置信度: %.2f", captcha.Confidence))
			util.LogInfo(fmt.Sprintf("  📋 处理策略: %s", captcha.GetHandlingStrategy()))
			if captcha.Selector != "" {
				util.LogInfo(fmt.Sprintf("  🎯 选择器: %s", captcha.Selector))
			}
			if captcha.ImageURL != "" {
				util.LogInfo(fmt.Sprintf("  🖼️ 图片URL: %s", captcha.ImageURL))
			}
		}
	}

	// 显示页面源码（如果需要的话）
	if analysis.PageSource != "" {
		util.LogInfo("\n=== 页面源码 ===")
		fmt.Println(analysis.PageSource)
		util.LogInfo("=== 页面源码结束 ===")
	}

	// 显示错误信息
	if analysis.ErrorMessage != "" {
		util.LogError(fmt.Sprintf("分析过程中出现错误: %s", analysis.ErrorMessage))
	}
}

func printBruteForceResult(result *bruteforce.BruteForceResult) {
	util.LogInfo("=== 爆破结果 ===")

	if result.Success {
		util.LogSuccess("🎉 爆破成功!")
		util.LogInfo(fmt.Sprintf("用户名: %s", result.Username))
		util.LogInfo(fmt.Sprintf("密码: %s", result.Password))
		util.LogInfo(fmt.Sprintf("目标URL: %s", result.URL))

		// 保存截图
		if len(result.Screenshot) > 0 {
			if err := os.WriteFile("success_screenshot.png", result.Screenshot, 0644); err == nil {
				util.LogInfo("成功截图已保存: success_screenshot.png")
			}
		}
	} else {
		util.LogFailure("❌ 爆破失败")
		util.LogWarn(fmt.Sprintf("失败原因: %s", result.ErrorMessage))
		util.LogInfo(fmt.Sprintf("目标URL: %s", result.URL))
	}
}
