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
		// SSL证书相关配置 - 忽略证书错误
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("ignore-ssl-errors", true),
		chromedp.Flag("ignore-certificate-errors-spki-list", true),
		chromedp.Flag("ignore-certificate-errors-ssl-errors", true),
		chromedp.Flag("allow-running-insecure-content", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("allow-cross-origin-auth-prompt", true),
		// 网络相关配置
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.WindowSize(b.config.Browser.Width, b.config.Browser.Height),
	)

	// 如果指定了Chrome路径，使用自定义路径
	if b.config.Browser.ChromePath != "" {
		opts = append(opts, chromedp.ExecPath(b.config.Browser.ChromePath))
		b.logger.Info("使用指定的Chrome路径: %s", b.config.Browser.ChromePath)
	}

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

// GetContext 获取浏览器上下文
func (b *Browser) GetContext() context.Context {
	return b.ctx
}

// GetInputValue 获取输入框的当前值
func (b *Browser) GetInputValue(selector string) (string, error) {
	b.logger.Debugf("🔍 获取输入框值: %s", selector)

	timeoutCtx, cancel := context.WithTimeout(b.ctx, 5*time.Second)
	defer cancel()

	var value string
	err := chromedp.Run(timeoutCtx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Value(selector, &value, chromedp.ByQuery),
	)

	if err != nil {
		return "", fmt.Errorf("获取输入框值失败: %v", err)
	}

	return value, nil
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

	timeoutCtx, cancel := context.WithTimeout(b.ctx, 15*time.Second)
	defer cancel()

	err := chromedp.Run(timeoutCtx,
		// 等待元素可见
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond), // 等待元素完全加载

		// 先点击激活输入框
		chromedp.Click(selector, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		// 聚焦到输入框
		chromedp.Focus(selector, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		// 第一步：彻底清空输入框
		chromedp.Evaluate(fmt.Sprintf(`
			try {
				const el = document.querySelector('%s');
				if (el) {
					// 聚焦元素
					el.focus();
					
					// 全选内容
					el.select();
					if (el.setSelectionRange) {
						el.setSelectionRange(0, el.value.length);
					}
					
					// 使用execCommand删除
					document.execCommand('selectAll');
					document.execCommand('delete');
					
					// 强制设置为空
					el.value = '';
					el.textContent = '';
					if (el.innerHTML !== undefined) el.innerHTML = '';
					
					// 再次全选并删除以确保完全清空
					el.select();
					document.execCommand('delete');
					el.value = '';
					
					console.log('Step 1 - Input cleared, value now: "' + el.value + '"');
				}
			} catch(e) { console.log('Step 1 clear failed:', e); }
		`, escapeJSString(selector)), nil),

		chromedp.Sleep(800*time.Millisecond), // 更长的等待时间确保清空完成

		// 第二步：验证清空结果
		chromedp.Evaluate(fmt.Sprintf(`
			try {
				const el = document.querySelector('%s');
				if (el && el.value !== '') {
					console.log('WARNING: Input not fully cleared, value: "' + el.value + '"');
					// 如果还有内容，再次强制清空
					el.value = '';
					el.focus();
					el.select();
					document.execCommand('delete');
					el.value = '';
				}
				console.log('Step 2 - Final clear check, value: "' + el.value + '"');
			} catch(e) { console.log('Step 2 verify failed:', e); }
		`, escapeJSString(selector)), nil),

		chromedp.Sleep(400*time.Millisecond),

		// 第三步：设置新值
		chromedp.Evaluate(fmt.Sprintf(`
			try {
				const el = document.querySelector('%s');
				if (el) {
					// 确保元素处于聚焦状态
					el.focus();
					
					// 最后一次确保清空
					el.value = '';
					
					// 设置新值
					el.value = '%s';
					
					// 触发所有相关事件
					el.dispatchEvent(new Event('input', { bubbles: true, cancelable: true }));
					el.dispatchEvent(new Event('change', { bubbles: true, cancelable: true }));
					el.dispatchEvent(new Event('keyup', { bubbles: true, cancelable: true }));
					el.dispatchEvent(new Event('blur', { bubbles: true, cancelable: true }));
					
					console.log('Step 3 - Input filled with: "' + el.value + '"');
				}
			} catch(e) { console.log('Step 3 fill failed:', e); }
		`, escapeJSString(selector), escapeJSString(value)), nil),

		chromedp.Sleep(400*time.Millisecond), // 等待设置完成
		chromedp.Sleep(500*time.Millisecond), // 确保输入完成
	)

	if err == nil {
		// 验证输入是否成功
		if err := b.verifyInput(selector, value); err != nil {
			b.logger.Warnf("⚠️  输入验证失败，尝试重新输入: %v", err)
			// 重试一次
			err = b.retryFillInput(selector, value)
		} else {
			b.logger.Debugf("✅ 成功填充输入框: %s", selector)
		}
	}

	return err
}

// retryFillInput 重试填充输入框
func (b *Browser) retryFillInput(selector, value string) error {
	b.logger.Debugf("🔄 重试填充输入框: %s", selector)

	timeoutCtx, cancel := context.WithTimeout(b.ctx, 10*time.Second)
	defer cancel()

	// 更激进的清空和输入策略
	err := chromedp.Run(timeoutCtx,
		// 点击并聚焦
		chromedp.Click(selector, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		// 使用更强力的清空和设置方法
		chromedp.Evaluate(fmt.Sprintf(`
			try {
				const el = document.querySelector('%s');
				if (el) {
					// 更激进的清空方法
					el.focus();
					
					// 连续多次全选删除
					for (let i = 0; i < 3; i++) {
						el.select();
						if (el.setSelectionRange) {
							el.setSelectionRange(0, el.value.length);
						}
						document.execCommand('selectAll');
						document.execCommand('delete');
						el.value = '';
					}
					
					// 最后确保完全为空
					el.value = '';
					el.textContent = '';
					if (el.innerHTML !== undefined) el.innerHTML = '';
					
					console.log('Retry clear completed, value: "' + el.value + '"');
				}
			} catch(e) { console.log('Retry clear failed:', e); }
		`, escapeJSString(selector)), nil),

		chromedp.Sleep(600*time.Millisecond), // 更长等待时间

		// 设置新值
		chromedp.Evaluate(fmt.Sprintf(`
			try {
				const el = document.querySelector('%s');
				if (el) {
					el.focus();
					el.value = '%s';
					
					// 触发事件
					el.dispatchEvent(new Event('input', { bubbles: true, cancelable: true }));
					el.dispatchEvent(new Event('change', { bubbles: true, cancelable: true }));
					el.dispatchEvent(new Event('keyup', { bubbles: true, cancelable: true }));
					
					console.log('Retry fill completed, value: "' + el.value + '"');
				}
			} catch(e) { console.log('Retry fill failed:', e); }
		`, escapeJSString(selector), escapeJSString(value)), nil),
		chromedp.Sleep(500*time.Millisecond),
	)

	if err == nil {
		// 最终验证
		if verifyErr := b.verifyInput(selector, value); verifyErr != nil {
			b.logger.Warnf("⚠️  重试后仍然验证失败: %v", verifyErr)
			return verifyErr
		} else {
			b.logger.Debugf("✅ 重试填充成功: %s", selector)
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

// ClickCheckbox 点击复选框
func (b *Browser) ClickCheckbox(selector string) error {
	b.logger.Debugf("☑️  点击复选框: %s", selector)

	timeoutCtx, cancel := context.WithTimeout(b.ctx, 10*time.Second)
	defer cancel()

	// 首先检查复选框是否已经被选中
	var checkedAttr string
	var isChecked bool
	err := chromedp.Run(timeoutCtx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.AttributeValue(selector, "checked", &checkedAttr, &isChecked),
	)

	if err != nil {
		b.logger.Warnf("获取复选框状态失败，继续尝试点击: %v", err)
	}

	if isChecked && checkedAttr != "" {
		b.logger.Debug("✅ 复选框已经选中，无需点击")
		return nil
	}

	// 点击复选框
	err = chromedp.Run(timeoutCtx,
		chromedp.Click(selector, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // 等待状态更新
	)

	if err != nil {
		// 如果普通点击失败，尝试用JavaScript点击
		b.logger.Debugf("普通点击失败，尝试JavaScript点击")
		err = chromedp.Run(timeoutCtx,
			chromedp.Evaluate(fmt.Sprintf(`
				try {
					const checkbox = document.querySelector('%s');
					if (checkbox && !checkbox.checked) {
						checkbox.click();
					}
				} catch(e) { console.log('Checkbox click failed:', e); }
			`, escapeJSString(selector)), nil),
		)
	}

	if err == nil {
		b.logger.Debugf("✅ 成功点击复选框: %s", selector)
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
