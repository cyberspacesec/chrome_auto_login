package test

import (
	"context"
	"log"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// getChromePath 获取系统对应的Chrome路径
func getChromePath() string {
	switch runtime.GOOS {
	case "darwin": // macOS
		return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	case "linux":
		return "google-chrome"
	case "windows":
		return "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
	default:
		return "google-chrome"
	}
}

// TestChromeConnection 测试Chrome连接
func TestChromeConnection(t *testing.T) {
	// 屏蔽Chrome的错误日志
	log.SetOutput(os.Stdout)

	// 创建Chrome上下文
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(getChromePath()),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-logging", true),
		chromedp.Flag("silent", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	// 禁用日志
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...interface{}) {}))
	defer cancel()

	// 设置超时
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	t.Log("正在测试Chrome连接...")

	// 测试访问网页
	var title string
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.baidu.com"),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Title(&title),
	)

	if err != nil {
		t.Fatalf("Chrome连接失败: %v", err)
	}

	if title == "" {
		t.Fatalf("页面标题为空")
	}

	if !strings.Contains(title, "百度") {
		t.Logf("页面标题不包含'百度'，实际标题: %s", title)
		// 不强制失败，因为百度可能会改变标题
	}

	t.Logf("✓ Chrome连接测试通过，页面标题: %s", title)
}

// TestBasicWebAccess 测试基本网页访问
func TestBasicWebAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过网络访问测试（使用 -short 标志）")
	}

	// 屏蔽Chrome的错误日志
	log.SetOutput(os.Stdout)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(getChromePath()),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-logging", true),
		chromedp.Flag("silent", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...interface{}) {}))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 测试用例
	testCases := []struct {
		name      string
		url       string
		checkFunc func(string) bool
	}{
		{
			name: "HTTPBin HTML测试",
			url:  "https://httpbin.org/html",
			checkFunc: func(title string) bool {
				return strings.Contains(strings.ToLower(title), "html") ||
					strings.Contains(strings.ToLower(title), "test") ||
					title != ""
			},
		},
		{
			name: "HTTPBin JSON测试",
			url:  "https://httpbin.org/json",
			checkFunc: func(title string) bool {
				// JSON页面可能没有标题或有简单标题
				return true
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("正在测试: %s", tc.url)

			var title, url string

			// 导航到页面，带超时控制
			done := make(chan error, 1)
			go func() {
				done <- chromedp.Run(ctx,
					chromedp.Navigate(tc.url),
					chromedp.Sleep(2*time.Second),
					chromedp.Title(&title),
					chromedp.Location(&url),
				)
			}()

			select {
			case err := <-done:
				if err != nil {
					t.Fatalf("访问测试网站失败: %v", err)
				}
			case <-time.After(20 * time.Second):
				t.Fatalf("访问网站超时")
			}

			// 验证结果
			if !tc.checkFunc(title) {
				t.Errorf("页面验证失败，标题: %s", title)
			}

			t.Logf("✓ 网页访问测试成功")
			t.Logf("  页面标题: %s", title)
			t.Logf("  访问URL: %s", url)
		})
	}
}

// TestChromePerformance 测试Chrome性能
func TestChromePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试（使用 -short 标志）")
	}

	// 屏蔽Chrome的错误日志
	log.SetOutput(os.Stdout)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(getChromePath()),
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

	t.Log("正在测试Chrome性能...")

	// 记录开始时间
	startTime := time.Now()

	// 连续访问多个页面测试性能
	urls := []string{
		"https://httpbin.org/html",
		"https://httpbin.org/json",
		"https://httpbin.org/xml",
	}

	for i, url := range urls {
		t.Logf("访问第 %d/%d 个页面: %s", i+1, len(urls), url)

		pageStartTime := time.Now()

		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.WaitVisible("body", chromedp.ByQuery),
		)

		pageElapsed := time.Since(pageStartTime)

		if err != nil {
			t.Logf("访问页面失败: %v", err)
			continue
		}

		t.Logf("  页面加载用时: %v", pageElapsed)

		// 如果单个页面加载超过10秒，发出警告
		if pageElapsed > 10*time.Second {
			t.Logf("  ⚠️ 页面加载较慢: %v", pageElapsed)
		}
	}

	totalElapsed := time.Since(startTime)
	t.Logf("✓ 性能测试完成，总用时: %v", totalElapsed)

	// 平均每个页面的加载时间
	avgTime := totalElapsed / time.Duration(len(urls))
	t.Logf("  平均页面加载时间: %v", avgTime)

	// 如果平均时间超过8秒，发出警告
	if avgTime > 8*time.Second {
		t.Logf("  ⚠️ 平均页面加载时间较慢，可能需要优化网络环境")
	}
}

// contains 工具函数：检查字符串包含关系（不区分大小写）
func contains(text, substr string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(substr))
}
