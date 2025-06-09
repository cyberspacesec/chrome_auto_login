package detector

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/cyberspacesec/chrome_auto_login/pkg/browser"
	"github.com/cyberspacesec/chrome_auto_login/pkg/config"
)

// LoginFormElements 登录表单元素
type LoginFormElements struct {
	UsernameSelector string
	PasswordSelector string
	CaptchaSelector  string
	SubmitSelector   string
	HasCaptcha       bool
	CaptchaInfo      *CaptchaInfo
}

// PageDetector 页面检测器
type PageDetector struct {
	browser         *browser.Browser
	config          *config.Config
	logger          *logrus.Logger
	captchaDetector *CaptchaDetector
}

// NewPageDetector 创建页面检测器
func NewPageDetector(browser *browser.Browser, cfg *config.Config, logger *logrus.Logger) *PageDetector {
	detector := &PageDetector{
		browser: browser,
		config:  cfg,
		logger:  logger,
	}
	// 创建验证码检测器
	detector.captchaDetector = NewCaptchaDetector(browser, cfg, logger)
	return detector
}

// IsLoginPage 检测当前页面是否为登录页面
func (d *PageDetector) IsLoginPage() (bool, error) {
	title, url, content, err := d.browser.GetPageInfo()
	if err != nil {
		return false, fmt.Errorf("获取页面信息失败: %v", err)
	}

	d.logger.Infof("检测页面: 标题='%s', URL='%s'", title, url)

	isLogin := d.config.IsLoginPage(title, url, content)

	if isLogin {
		d.logger.Info("✓ 检测到登录页面")
	} else {
		d.logger.Info("✗ 非登录页面")
	}

	return isLogin, nil
}

// DetectLoginForm 检测登录表单元素
func (d *PageDetector) DetectLoginForm() (*LoginFormElements, error) {
	d.logger.Info("开始检测登录表单元素...")

	elements := &LoginFormElements{}

	// 检测用户名输入框
	usernameSelector, err := d.browser.FindElement(d.config.GetUsernameSelectors())
	if err != nil {
		return nil, fmt.Errorf("检测用户名输入框失败: %v", err)
	}
	if usernameSelector == "" {
		d.logger.Warn("未找到用户名输入框")
	} else {
		elements.UsernameSelector = usernameSelector
		d.logger.Infof("✓ 找到用户名输入框: %s", usernameSelector)
	}

	// 检测密码输入框
	passwordSelector, err := d.browser.FindElement(d.config.GetPasswordSelectors())
	if err != nil {
		return nil, fmt.Errorf("检测密码输入框失败: %v", err)
	}
	if passwordSelector == "" {
		d.logger.Warn("未找到密码输入框")
	} else {
		elements.PasswordSelector = passwordSelector
		d.logger.Infof("✓ 找到密码输入框: %s", passwordSelector)
	}

	// 检测验证码
	if d.config.Captcha.Detection.Enabled {
		d.logger.Debug("🔍 开始智能验证码检测...")
		detectedCaptcha, err := d.captchaDetector.DetectCaptcha()
		if err == nil && detectedCaptcha != nil && detectedCaptcha.Type != CaptchaTypeNone {
			elements.CaptchaInfo = detectedCaptcha
			elements.CaptchaSelector = detectedCaptcha.Selector
			elements.HasCaptcha = true
			d.logger.Infof("🎯 检测到验证码: %s", detectedCaptcha.GetTypeName())
			d.logger.Infof("📋 处理策略: %s", detectedCaptcha.GetHandlingStrategy())
		}
	}

	// 如果智能检测没有找到，回退到传统方法
	if !elements.HasCaptcha {
		captchaSelector, err := d.browser.FindElement(d.config.GetCaptchaSelectors())
		if err != nil {
			d.logger.Warnf("检测验证码输入框时出错: %v", err)
		}
		if captchaSelector != "" {
			elements.CaptchaSelector = captchaSelector
			elements.HasCaptcha = true
			d.logger.Infof("✓ 通过传统方法找到验证码输入框: %s", captchaSelector)
			// 创建简单的验证码信息
			elements.CaptchaInfo = &CaptchaInfo{
				Type:        CaptchaTypeText,
				Selector:    captchaSelector,
				Description: "传统验证码输入框",
				Confidence:  0.6,
			}
		} else {
			d.logger.Info("未检测到验证码")
		}
	}

	// 检测提交按钮
	submitSelector, err := d.browser.FindElement(d.config.GetSubmitSelectors())
	if err != nil {
		return nil, fmt.Errorf("检测提交按钮失败: %v", err)
	}
	if submitSelector == "" {
		d.logger.Warn("未找到提交按钮")
	} else {
		elements.SubmitSelector = submitSelector
		d.logger.Infof("✓ 找到提交按钮: %s", submitSelector)
	}

	// 验证必要元素
	if elements.UsernameSelector == "" || elements.PasswordSelector == "" || elements.SubmitSelector == "" {
		return nil, fmt.Errorf("缺少关键登录元素")
	}

	d.logger.Info("登录表单元素检测完成")
	return elements, nil
}

// AnalyzePage 分析页面并给出详细报告
func (d *PageDetector) AnalyzePage() (map[string]interface{}, error) {
	title, url, content, err := d.browser.GetPageInfo()
	if err != nil {
		return nil, fmt.Errorf("获取页面信息失败: %v", err)
	}

	analysis := map[string]interface{}{
		"title":    title,
		"url":      url,
		"is_login": d.config.IsLoginPage(title, url, content),
	}

	// 检测各种表单元素
	formElements := map[string]interface{}{}

	// 用户名输入框
	if usernameSelector, _ := d.browser.FindElement(d.config.GetUsernameSelectors()); usernameSelector != "" {
		formElements["username"] = usernameSelector
	}

	// 密码输入框
	if passwordSelector, _ := d.browser.FindElement(d.config.GetPasswordSelectors()); passwordSelector != "" {
		formElements["password"] = passwordSelector
	}

	// 验证码输入框
	if captchaSelector, _ := d.browser.FindElement(d.config.GetCaptchaSelectors()); captchaSelector != "" {
		formElements["captcha"] = captchaSelector
	}

	// 提交按钮
	if submitSelector, _ := d.browser.FindElement(d.config.GetSubmitSelectors()); submitSelector != "" {
		formElements["submit"] = submitSelector
	}

	analysis["form_elements"] = formElements

	// 检测页面特征
	var features []string
	if strings.Contains(strings.ToLower(content), "username") || strings.Contains(content, "用户名") {
		features = append(features, "包含用户名字段")
	}
	if strings.Contains(strings.ToLower(content), "password") || strings.Contains(content, "密码") {
		features = append(features, "包含密码字段")
	}
	if strings.Contains(strings.ToLower(content), "captcha") || strings.Contains(content, "验证码") {
		features = append(features, "包含验证码")
	}
	if strings.Contains(strings.ToLower(content), "login") || strings.Contains(content, "登录") {
		features = append(features, "包含登录文本")
	}

	analysis["page_features"] = features

	// 检测验证码
	if d.config.Captcha.Detection.Enabled {
		captchaInfo, err := d.captchaDetector.DetectCaptcha()
		if err == nil && captchaInfo != nil {
			captchaAnalysis := map[string]interface{}{
				"type":       captchaInfo.GetTypeName(),
				"confidence": captchaInfo.Confidence,
				"strategy":   captchaInfo.GetHandlingStrategy(),
				"selector":   captchaInfo.Selector,
			}
			analysis["captcha_info"] = captchaAnalysis
		}
	}

	return analysis, nil
}
