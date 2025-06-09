package main

import (
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
║    │  🎯 版本: %-10s    👨‍💻 作者: %-20s              │    ║
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
		configFile = flag.String("config", "config/config.yaml", "配置文件路径")
		targetURL  = flag.String("url", "", "目标登录页面URL")
		analyze    = flag.Bool("analyze", false, "仅分析页面，不执行爆破")
		debug      = flag.Bool("debug", false, "调试模式，显示浏览器窗口和详细操作过程")
		help       = flag.Bool("help", false, "显示帮助信息")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *targetURL == "" {
		fmt.Println("错误: 必须指定目标URL")
		fmt.Println("使用 -help 查看帮助信息")
		os.Exit(1)
	}

	// 加载配置
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
		os.Exit(1)
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

	// 导航到目标URL
	if err := browserInstance.NavigateTo(*targetURL); err != nil {
		util.LogError(fmt.Sprintf("导航到目标URL失败: %v", err))
		os.Exit(1)
	}

	// 如果只是分析模式
	if *analyze {
		util.LogInfo("=== 页面分析模式 ===")

		// 分析页面
		analysis, err := pageDetector.AnalyzePage()
		if err != nil {
			util.LogError(fmt.Sprintf("页面分析失败: %v", err))
			os.Exit(1)
		}

		// 输出分析结果
		printAnalysisResult(analysis)
		return
	}

	// 创建状态显示器和进度感知日志器
	statusDisplay := util.NewStatusDisplay()
	progressLogger := util.NewProgressAwareLogger(statusDisplay)

	// 创建爆破引擎
	bruteforceEngine := bruteforce.NewBruteForceEngine(browserInstance, pageDetector, cfg, progressLogger)

	// 显示爆破信息
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🎯 自动化登录爆破即将开始")
	fmt.Println(strings.Repeat("=", 70))

	// 执行爆破
	result, err := bruteforceEngine.ExecuteBruteForce(*targetURL)
	if err != nil {
		util.LogError(fmt.Sprintf("爆破执行失败: %v", err))
		os.Exit(1)
	}

	// 输出结果
	printBruteForceResult(result)
}

func showHelp() {
	// 打印带有版本信息的标题
	fmt.Printf("\n╔═══════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                  🔐 Chrome自动化登录爆破工具                           ║\n")
	fmt.Printf("║                     版本: %-10s | 作者: %-15s       ║\n", Version, Author)
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Println("用法:")
	fmt.Println("  ./chrome_auto_login -url <目标URL> [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -url string      目标登录页面URL (必需)")
	fmt.Println("  -config string   配置文件路径 (默认: config/config.yaml)")
	fmt.Println("  -analyze         仅分析页面，不执行爆破")
	fmt.Println("  -debug           调试模式，显示浏览器窗口和详细操作过程")
	fmt.Println("  -help            显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 基本用法")
	fmt.Println("  ./chrome_auto_login -url \"http://example.com/login\"")
	fmt.Println()
	fmt.Println("  # 调试模式（显示浏览器窗口）")
	fmt.Println("  ./chrome_auto_login -url \"http://example.com/login\" -debug")
	fmt.Println()
	fmt.Println("  # 使用自定义配置文件")
	fmt.Println("  ./chrome_auto_login -url \"http://example.com/login\" -config \"my_config.yaml\"")
	fmt.Println()
	fmt.Println("  # 仅分析页面")
	fmt.Println("  ./chrome_auto_login -url \"http://example.com/login\" -analyze")
	fmt.Println()
	fmt.Println("注意:")
	fmt.Println("  - 此工具仅用于授权测试，请勿用于非法用途")
	fmt.Println("  - 确保已安装Chrome浏览器")
	fmt.Println("  - 配置文件中可以自定义识别规则和字典")
	fmt.Println()
	fmt.Println("🌐 项目地址: https://github.com/cyberspacesec/chrome_auto_login")
	fmt.Println("📧 联系作者: zhizhuo@cyberspacesec.com")
	fmt.Println("🔒 安全提醒: 仅限授权渗透测试使用")
	fmt.Println()
}

func printAnalysisResult(analysis map[string]interface{}) {
	util.LogInfo("=== 页面分析结果 ===")
	util.LogInfo(fmt.Sprintf("页面标题: %v", analysis["title"]))
	util.LogInfo(fmt.Sprintf("页面URL: %v", analysis["url"]))
	util.LogInfo(fmt.Sprintf("是否为登录页面: %v", analysis["is_login"]))

	if formElements, ok := analysis["form_elements"].(map[string]interface{}); ok {
		util.LogInfo("检测到的表单元素:")
		for elementType, selector := range formElements {
			util.LogInfo(fmt.Sprintf("  %s: %v", elementType, selector))
		}
	}

	// 显示验证码检测结果
	if captchaInfo, ok := analysis["captcha_info"].(map[string]interface{}); ok {
		util.LogInfo("验证码检测结果:")
		if captchaType, exists := captchaInfo["type"]; exists {
			util.LogInfo(fmt.Sprintf("  🎯 类型: %s", captchaType))
		}
		if confidence, exists := captchaInfo["confidence"]; exists {
			util.LogInfo(fmt.Sprintf("  📊 置信度: %.2f", confidence))
		}
		if strategy, exists := captchaInfo["strategy"]; exists {
			util.LogInfo(fmt.Sprintf("  📋 处理策略: %s", strategy))
		}
		if selector, exists := captchaInfo["selector"]; exists && selector != "" {
			util.LogInfo(fmt.Sprintf("  🎯 选择器: %s", selector))
		}
	}

	if features, ok := analysis["page_features"].([]string); ok {
		util.LogInfo("页面特征:")
		for _, feature := range features {
			util.LogInfo(fmt.Sprintf("  • %s", feature))
		}
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
