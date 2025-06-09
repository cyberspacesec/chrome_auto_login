package browser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"

	"github.com/cyberspacesec/chrome_auto_login/pkg/config"
)

// Browser 浏览器实例
type Browser struct {
	ctx    context.Context
	cancel context.CancelFunc
	config *config.Config
	logger *logrus.Logger
}

// NewBrowser 创建新的浏览器实例
func NewBrowser(cfg *config.Config, logger *logrus.Logger) *Browser {
	return &Browser{
		config: cfg,
		logger: logger,
	}
}

// Start 启动浏览器
func (b *Browser) Start() error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"),
		chromedp.Flag("headless", b.config.Browser.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-logging", true),
		chromedp.Flag("silent", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.WindowSize(b.config.Browser.Width, b.config.Browser.Height),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	
	// 创建自定义日志函数，过滤Chrome内部错误
	customLogf := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		// 过滤掉Chrome内部的错误信息
		if strings.Contains(msg, "could not unmarshal event") ||
		   strings.Contains(msg, "cookiePart") ||
		   strings.Contains(msg, "unknown ClientNavigationReason") ||
		   strings.Contains(msg, "parse error") {
			return // 忽略这些内部错误
		}
		// 只在debug模式下输出其他Chrome日志
		if !b.config.Browser.Headless {
			b.logger.Debug("Chrome: " + msg)
		}
	}

	// 统一使用自定义日志函数
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(customLogf))
	
	if !b.config.Browser.Headless {
		b.logger.Debug("🔍 调试模式：Chrome窗口可见，已屏蔽内部错误日志")
	}

	b.ctx = ctx
	b.cancel = func() {
		cancel()
		allocCancel()
	}

	// 启动浏览器（不设置超时，因为这只是启动浏览器进程）
	return chromedp.Run(b.ctx)
}

// Close 关闭浏览器
func (b *Browser) Close() {
	if b.cancel != nil {
		b.cancel()
	}
}

// NavigateTo 导航到指定URL
func (b *Browser) NavigateTo(url string) error {
	b.logger.Infof("导航到: %s", url)
	
	timeoutCtx, cancel := context.WithTimeout(b.ctx, time.Duration(b.config.Browser.Timeout)*time.Second)
	defer cancel()

	return chromedp.Run(timeoutCtx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second), // 给页面一些时间加载
	)
}

// GetPageInfo 获取页面信息
func (b *Browser) GetPageInfo() (title, url, content string, err error) {
	timeoutCtx, cancel := context.WithTimeout(b.ctx, 10*time.Second)
	defer cancel()

	err = chromedp.Run(timeoutCtx,
		chromedp.Title(&title),
		chromedp.Location(&url),
		chromedp.Text("body", &content, chromedp.ByQuery),
	)
	
	return title, url, content, err
}

// GetPageContent 获取页面内容
func (b *Browser) GetPageContent() (string, error) {
	var content string
	timeoutCtx, cancel := context.WithTimeout(b.ctx, 10*time.Second)
	defer cancel()

	err := chromedp.Run(timeoutCtx,
		chromedp.Text("body", &content, chromedp.ByQuery),
	)
	
	return content, err
}

// FindElement 查找页面元素
func (b *Browser) FindElement(selectors []string) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(b.ctx, 5*time.Second)
	defer cancel()

	for _, selector := range selectors {
		var nodes []*cdp.Node
		err := chromedp.Run(timeoutCtx,
			chromedp.Nodes(selector, &nodes, chromedp.AtLeast(0)),
		)
		
		if err == nil && len(nodes) > 0 {
			b.logger.Debugf("找到元素: %s", selector)
			return selector, nil
		}
	}
	
	return "", nil
}

// FillInput 填充输入框
func (b *Browser) FillInput(selector, value string) error {
	b.logger.Debugf("🖊️  填充输入框 %s: %s", selector, value)
	
	timeoutCtx, cancel := context.WithTimeout(b.ctx, 10*time.Second)
	defer cancel()

	err := chromedp.Run(timeoutCtx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		// 多重清空确保输入框完全清空
		chromedp.Focus(selector),
		chromedp.Clear(selector),
		// 使用JavaScript清空，更可靠
		chromedp.Evaluate(fmt.Sprintf(`
			try {
				const el = document.querySelector('%s');
				if (el) {
					el.value = '';
					el.focus();
				}
			} catch(e) { console.log('Clear failed:', e); }
		`, escapeJSString(selector)), nil),
		chromedp.Sleep(200*time.Millisecond), // 短暂延迟
		chromedp.SendKeys(selector, value),
		chromedp.Sleep(500*time.Millisecond), // 确保输入完成
	)
	
	if err == nil {
		// 验证输入是否成功
		if err := b.verifyInput(selector, value); err != nil {
			b.logger.Warnf("⚠️  输入验证失败: %v", err)
		} else {
			b.logger.Debugf("✅ 成功填充输入框: %s", selector)
		}
	}
	
	return err
}

// verifyInput 验证输入框的值是否正确
func (b *Browser) verifyInput(selector, expectedValue string) error {
	var actualValue string
	timeoutCtx, cancel := context.WithTimeout(b.ctx, 5*time.Second)
	defer cancel()

	err := chromedp.Run(timeoutCtx,
		chromedp.Value(selector, &actualValue, chromedp.ByQuery),
	)
	
	if err != nil {
		return fmt.Errorf("获取输入框值失败: %v", err)
	}
	
	if actualValue != expectedValue {
		b.logger.Debugf("输入验证: 期望='%s', 实际='%s'", expectedValue, actualValue)
		return fmt.Errorf("输入值不匹配: 期望='%s', 实际='%s'", expectedValue, actualValue)
	}
	
	return nil
}

// ClickElement 点击元素
func (b *Browser) ClickElement(selector string) error {
	b.logger.Debugf("🖱️  点击元素: %s", selector)
	
	timeoutCtx, cancel := context.WithTimeout(b.ctx, 10*time.Second)
	defer cancel()

	err := chromedp.Run(timeoutCtx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // 等待元素完全可见
		// 先尝试普通点击
		chromedp.Click(selector, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
	)
	
	if err != nil {
		// 如果普通点击失败，尝试JavaScript点击
		b.logger.Debugf("普通点击失败，尝试JavaScript点击...")
		err = chromedp.Run(timeoutCtx,
			chromedp.Evaluate(fmt.Sprintf(`
				try {
					const el = document.querySelector('%s');
					if (el) {
						el.click();
					}
				} catch(e) { console.log('Click failed:', e); }
			`, escapeJSString(selector)), nil),
		)
	}
	
	if err == nil {
		b.logger.Debugf("✅ 成功点击元素: %s", selector)
	}
	
	return err
}

// GetCurrentURL 获取当前URL
func (b *Browser) GetCurrentURL() (string, error) {
	var url string
	timeoutCtx, cancel := context.WithTimeout(b.ctx, 5*time.Second)
	defer cancel()

	err := chromedp.Run(timeoutCtx, chromedp.Location(&url))
	return url, err
}

// Screenshot 截图
func (b *Browser) Screenshot() ([]byte, error) {
	var buf []byte
	timeoutCtx, cancel := context.WithTimeout(b.ctx, 10*time.Second)
	defer cancel()

	err := chromedp.Run(timeoutCtx, chromedp.CaptureScreenshot(&buf))
	return buf, err
}

// escapeJSString 转义JavaScript字符串中的特殊字符
func escapeJSString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
} 