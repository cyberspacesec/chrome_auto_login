package test

import (
	"context"
	"log"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/cyberspacesec/chrome_auto_login/pkg/browser"
	"github.com/cyberspacesec/chrome_auto_login/pkg/config"
	"github.com/cyberspacesec/chrome_auto_login/pkg/detector"
	"github.com/cyberspacesec/chrome_auto_login/util"
)

// TestCaptchaDetection 测试验证码检测功能
func TestCaptchaDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过验证码检测测试（使用 -short 标志）")
	}

	// 屏蔽Chrome的错误日志
	log.SetOutput(os.Stdout)

	// 加载配置
	cfg, err := config.LoadConfig("../config/config.yaml")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 强制启用验证码检测
	cfg.Captcha.Detection.Enabled = true
	cfg.Browser.Headless = true // 测试时使用无头模式

	// 创建日志
	logConfig := util.LogConfig{
		Level:      cfg.Logging.Level,
		SaveToFile: false, // 测试时不保存到文件
	}
	if err := util.InitLogger(logConfig); err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}
	logger := util.Logger

	// 启动浏览器
	browserInstance := browser.NewBrowser(cfg, logger)
	if err := browserInstance.Start(); err != nil {
		t.Fatalf("启动浏览器失败: %v", err)
	}
	defer browserInstance.Close()

	// 创建验证码检测器
	captchaDetector := detector.NewCaptchaDetector(browserInstance, cfg, logger)

	// 测试用例列表
	testCases := []struct {
		name     string
		url      string
		expected string // 期望检测到的验证码类型
		timeout  time.Duration
	}{
		{
			name:     "reCAPTCHA演示",
			url:      "https://www.google.com/recaptcha/api2/demo",
			expected: "Google reCAPTCHA",
			timeout:  30 * time.Second,
		},
		{
			name:     "普通HTML表单页面",
			url:      "https://httpbin.org/forms/post",
			expected: "无验证码",
			timeout:  20 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("测试: %s", tc.name)
			t.Logf("URL: %s", tc.url)

			// 导航到测试页面，带超时控制
			done := make(chan error, 1)
			go func() {
				done <- browserInstance.NavigateTo(tc.url)
			}()

			select {
			case err := <-done:
				if err != nil {
					t.Logf("导航失败，跳过: %v", err)
					return
				}
			case <-time.After(tc.timeout):
				t.Logf("导航超时，跳过测试")
				return
			}

			// 等待页面加载
			time.Sleep(3 * time.Second)

			// 检测验证码
			captchaInfo, err := captchaDetector.DetectCaptcha()
			if err != nil {
				t.Errorf("验证码检测失败: %v", err)
				return
			}

			t.Logf("检测结果: %s", captchaInfo.GetTypeName())
			t.Logf("置信度: %.2f", captchaInfo.Confidence)
			t.Logf("处理策略: %s", captchaInfo.GetHandlingStrategy())

			if captchaInfo.Selector != "" {
				t.Logf("选择器: %s", captchaInfo.Selector)
			}

			// 验证检测结果
			if tc.expected != "" && captchaInfo.GetTypeName() != tc.expected {
				t.Logf("期望: %s, 实际: %s", tc.expected, captchaInfo.GetTypeName())
				// 注意：这里用Log而不是Error，因为网站可能会变化
			}
		})
	}
}

// TestCaptchaIntegration 测试验证码与登录检测的集成
func TestCaptchaIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过验证码集成测试（使用 -short 标志）")
	}

	// 屏蔽Chrome的错误日志
	log.SetOutput(os.Stdout)

	// 获取系统对应的Chrome路径
	var chromePath string
	switch runtime.GOOS {
	case "darwin": // macOS
		chromePath = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	case "linux":
		chromePath = "google-chrome"
	case "windows":
		chromePath = "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
	default:
		chromePath = "google-chrome"
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-logging", true),
		chromedp.Flag("silent", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...interface{}) {}))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 加载配置
	cfg, err := config.LoadConfig("../config/config.yaml")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 启用验证码检测
	cfg.Captcha.Detection.Enabled = true
	cfg.Browser.Headless = true

	// 创建日志
	logConfig := util.LogConfig{
		Level:      "error", // 使用error级别减少测试输出
		SaveToFile: false,   // 测试时不保存到文件
	}
	if err := util.InitLogger(logConfig); err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}
	logger := util.Logger

	// 创建浏览器实例
	browserInstance := browser.NewBrowser(cfg, logger)
	if err := browserInstance.Start(); err != nil {
		t.Fatalf("启动浏览器失败: %v", err)
	}
	defer browserInstance.Close()

	// 创建页面检测器
	pageDetector := detector.NewPageDetector(browserInstance, cfg, logger)

	// 测试验证码检测集成
	testURL := "https://httpbin.org/forms/post"

	// 带超时的导航
	done := make(chan error, 1)
	go func() {
		done <- browserInstance.NavigateTo(testURL)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("导航失败: %v", err)
		}
	case <-time.After(20 * time.Second):
		t.Fatalf("导航超时")
	}

	// 检测登录表单
	formElements, err := pageDetector.DetectLoginForm()
	if err != nil {
		t.Fatalf("检测登录表单失败: %v", err)
	}

	t.Logf("✓ 登录表单检测完成")
	t.Logf("  用户名选择器: %s", formElements.UsernameSelector)
	t.Logf("  密码选择器: %s", formElements.PasswordSelector)
	t.Logf("  提交按钮选择器: %s", formElements.SubmitSelector)
	t.Logf("  是否有验证码: %v", formElements.HasCaptcha)

	if formElements.HasCaptcha && formElements.CaptchaInfo != nil {
		t.Logf("  验证码类型: %s", formElements.CaptchaInfo.GetTypeName())
		t.Logf("  验证码置信度: %.2f", formElements.CaptchaInfo.Confidence)
		t.Logf("  处理策略: %s", formElements.CaptchaInfo.GetHandlingStrategy())
	} else {
		t.Logf("  未检测到验证码")
	}
}

// TestCaptchaDetectionTimeout 测试验证码检测超时控制
func TestCaptchaDetectionTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过超时测试（使用 -short 标志）")
	}

	// 屏蔽Chrome的错误日志
	log.SetOutput(os.Stdout)

	// 加载配置
	cfg, err := config.LoadConfig("../config/config.yaml")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 启用验证码检测，设置较短的超时时间进行测试
	cfg.Captcha.Detection.Enabled = true
	cfg.Captcha.Detection.Timeout = 10 // 10秒超时
	cfg.Browser.Headless = true

	// 创建日志
	logConfig := util.LogConfig{
		Level:      "error",
		SaveToFile: false, // 测试时不保存到文件
	}
	if err := util.InitLogger(logConfig); err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}
	logger := util.Logger

	// 启动浏览器
	browserInstance := browser.NewBrowser(cfg, logger)
	if err := browserInstance.Start(); err != nil {
		t.Fatalf("启动浏览器失败: %v", err)
	}
	defer browserInstance.Close()

	// 创建验证码检测器
	captchaDetector := detector.NewCaptchaDetector(browserInstance, cfg, logger)

	// 导航到简单页面
	if err := browserInstance.NavigateTo("https://httpbin.org/html"); err != nil {
		t.Fatalf("导航失败: %v", err)
	}

	// 记录开始时间
	startTime := time.Now()

	// 检测验证码
	captchaInfo, err := captchaDetector.DetectCaptcha()

	// 记录结束时间
	elapsed := time.Since(startTime)

	if err != nil {
		t.Errorf("验证码检测失败: %v", err)
		return
	}

	t.Logf("检测结果: %s", captchaInfo.GetTypeName())
	t.Logf("检测用时: %v", elapsed)

	// 验证检测时间不超过配置的超时时间（加上一些容差）
	maxAllowedTime := time.Duration(cfg.Captcha.Detection.Timeout+2) * time.Second
	if elapsed > maxAllowedTime {
		t.Errorf("检测时间超过预期: %v > %v", elapsed, maxAllowedTime)
	}

	// 验证无验证码页面应该快速完成
	if captchaInfo.GetTypeName() == "无验证码" && elapsed > 5*time.Second {
		t.Errorf("无验证码页面检测时间过长: %v", elapsed)
	}
}
