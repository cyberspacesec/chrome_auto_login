package test

import (
	"context"
	"log"
	"os"
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

	// 创建日志
	logger := util.SetupLogger(cfg.Logging.Level, "")

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
	}{
		{
			name:     "reCAPTCHA演示",
			url:      "https://www.google.com/recaptcha/api2/demo",
			expected: "Google reCAPTCHA",
		},
		{
			name:     "普通登录页面",
			url:      "https://httpbin.org/forms/post",
			expected: "无验证码",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("测试: %s", tc.name)
			t.Logf("URL: %s", tc.url)

			// 导航到测试页面
			if err := browserInstance.NavigateTo(tc.url); err != nil {
				t.Logf("导航失败，跳过: %v", err)
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

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"),
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

	// 创建日志
	logger := util.SetupLogger("debug", "")

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
	if err := browserInstance.NavigateTo(testURL); err != nil {
		t.Fatalf("导航失败: %v", err)
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
	}
} 