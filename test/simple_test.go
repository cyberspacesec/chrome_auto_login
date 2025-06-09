package test

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestChromeConnection 测试Chrome连接
func TestChromeConnection(t *testing.T) {
	// 屏蔽Chrome的错误日志
	log.SetOutput(os.Stdout)

	// 创建Chrome上下文
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

	// 禁用日志
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...interface{}) {}))
	defer cancel()

	// 设置超时
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

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

	if !strings.Contains(title, "百度") {
		t.Fatalf("页面标题不正确: %s", title)
	}

	fmt.Printf("✓ Chrome连接测试通过，页面标题: %s\n", title)
}

// TestBasicWebAccess 测试基本网页访问
func TestBasicWebAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过网络访问测试（使用 -short 标志）")
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

	var title, url string
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://httpbin.org/html"),
		chromedp.Sleep(2*time.Second),
		chromedp.Title(&title),
		chromedp.Location(&url),
	)

	if err != nil {
		t.Fatalf("访问测试网站失败: %v", err)
	}

	fmt.Printf("✓ 网页访问测试成功\n")
	fmt.Printf("  页面标题: %s\n", title)
	fmt.Printf("  访问URL: %s\n", url)
}

func contains(text, substr string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(substr))
} 