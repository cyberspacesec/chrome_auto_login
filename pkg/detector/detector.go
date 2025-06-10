package detector

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"

	"github.com/cyberspacesec/chrome_auto_login/pkg/browser"
	"github.com/cyberspacesec/chrome_auto_login/pkg/config"
)

// LoginFormElements 登录表单元素
type LoginFormElements struct {
	UsernameSelector string       `json:"username_selector"`
	PasswordSelector string       `json:"password_selector"`
	CaptchaSelector  string       `json:"captcha_selector"`
	SubmitSelector   string       `json:"submit_selector"`
	HasCaptcha       bool         `json:"has_captcha"`
	CaptchaInfo      *CaptchaInfo `json:"captcha_info"`
}

// PageAnalysis 页面分析结果
type PageAnalysis struct {
	Title            string             `json:"title"`
	URL              string             `json:"url"`
	IsLogin          bool               `json:"is_login"`
	Confidence       float64            `json:"confidence"`
	DetectedFeatures []string           `json:"detected_features"`
	FormElements     *LoginFormElements `json:"form_elements"`
	PageSource       string             `json:"page_source"`
	Encoding         string             `json:"encoding"`
	ResponseHeaders  map[string]string  `json:"response_headers"`
	LoadTime         time.Duration      `json:"load_time"`
	ErrorMessage     string             `json:"error_message"`
}

// PageDetector 页面检测器
type PageDetector struct {
	browser         *browser.Browser
	config          *config.Config
	logger          *logrus.Logger
	captchaDetector *CaptchaDetector

	// 检测阶段超时配置
	pageLoadTimeout      time.Duration // 页面加载超时
	elementDetectTimeout time.Duration // 元素检测超时
	analysisTimeout      time.Duration // 分析超时
}

// NewPageDetector 创建页面检测器
func NewPageDetector(browser *browser.Browser, cfg *config.Config, logger *logrus.Logger) *PageDetector {
	captchaDetector := NewCaptchaDetector(browser, cfg, logger)

	return &PageDetector{
		browser:              browser,
		config:               cfg,
		logger:               logger,
		captchaDetector:      captchaDetector,
		pageLoadTimeout:      60 * time.Second, // 页面加载最长60秒
		elementDetectTimeout: 10 * time.Second, // 元素检测10秒
		analysisTimeout:      15 * time.Second, // 页面分析15秒
	}
}

// IsLoginPage 检查是否为登录页面
func (pd *PageDetector) IsLoginPage() (bool, error) {
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(pd.browser.GetContext(), pd.analysisTimeout)
	defer cancel()

	// 获取页面基本信息
	var title, url, content string
	err := chromedp.Run(ctx,
		chromedp.Title(&title),
		chromedp.Location(&url),
		chromedp.Text("body", &content),
	)

	if err != nil {
		pd.logger.Warnf("⚠️ 获取页面信息失败: %v", err)
		return false, err
	}

	// 使用配置规则检测
	isLogin := pd.config.IsLoginPage(title, url, content)

	// 增强检测逻辑
	confidence := pd.calculateLoginConfidence(title, url, content, ctx)

	loadTime := time.Since(startTime)
	pd.logger.Debugf("页面检测完成，用时: %v, 置信度: %.2f", loadTime, confidence)

	// 置信度阈值判断
	if confidence >= 0.6 {
		isLogin = true
	}

	if isLogin {
		pd.logger.Info("✅ 确认为登录页面")
	} else {
		pd.logger.Info("❌ 非登录页面，将跳过处理")
	}

	return isLogin, nil
}

// calculateLoginConfidence 计算登录页面置信度
func (pd *PageDetector) calculateLoginConfidence(title, url, content string, ctx context.Context) float64 {
	var confidence float64

	// 标题权重: 30%
	titleScore := pd.checkTitleFeatures(title)
	confidence += titleScore * 0.3

	// URL权重: 20%
	urlScore := pd.checkURLFeatures(url)
	confidence += urlScore * 0.2

	// 内容权重: 25%
	contentScore := pd.checkContentFeatures(content)
	confidence += contentScore * 0.25

	// 表单元素权重: 25%
	formScore := pd.checkFormFeatures(ctx)
	confidence += formScore * 0.25

	return confidence
}

// checkTitleFeatures 检查标题特征
func (pd *PageDetector) checkTitleFeatures(title string) float64 {
	loginKeywords := []string{
		"登录", "登陆", "login", "sign in", "log in", "signin",
		"用户登录", "管理员登录", "后台登录", "系统登录",
		"admin", "administration", "后台管理", "管理系统",
		"auth", "authentication", "portal", "gateway",
	}

	title = strings.ToLower(title)
	score := 0.0

	for _, keyword := range loginKeywords {
		if strings.Contains(title, strings.ToLower(keyword)) {
			if strings.Contains(keyword, "login") || strings.Contains(keyword, "登录") {
				score += 0.4 // 核心关键词更高权重
			} else {
				score += 0.2
			}
		}
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}

// checkURLFeatures 检查URL特征
func (pd *PageDetector) checkURLFeatures(url string) float64 {
	urlPatterns := []string{
		`(?i).*/login.*`,
		`(?i).*/signin.*`,
		`(?i).*/auth.*`,
		`(?i).*/admin.*`,
		`(?i).*/user.*`,
		`(?i).*/portal.*`,
		`(?i).*/sso.*`,
		`(?i).*/oauth.*`,
	}

	score := 0.0
	url = strings.ToLower(url)

	for _, pattern := range urlPatterns {
		if matched, _ := regexp.MatchString(pattern, url); matched {
			score += 0.3
		}
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}

// checkContentFeatures 检查内容特征
func (pd *PageDetector) checkContentFeatures(content string) float64 {
	contentKeywords := map[string]float64{
		"用户名":      0.15,
		"密码":       0.15,
		"username": 0.15,
		"password": 0.15,
		"登录":       0.1,
		"login":    0.1,
		"账号":       0.1,
		"邮箱":       0.05,
		"手机号":      0.05,
		"验证码":      0.05,
		"captcha":  0.05,
		"记住我":      0.03,
		"忘记密码":     0.03,
		"注册":       -0.05, // 注册页面降低分数
		"register": -0.05,
	}

	content = strings.ToLower(content)
	score := 0.0

	for keyword, weight := range contentKeywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			score += weight
		}
	}

	if score > 1.0 {
		score = 1.0
	} else if score < 0 {
		score = 0.0
	}

	return score
}

// checkFormFeatures 检查表单特征
func (pd *PageDetector) checkFormFeatures(ctx context.Context) float64 {
	score := 0.0

	// 检查用户名输入框
	usernameFound := pd.checkElementsExist(ctx, pd.config.GetUsernameSelectors())
	if usernameFound {
		score += 0.4
	}

	// 检查密码输入框
	passwordFound := pd.checkElementsExist(ctx, pd.config.GetPasswordSelectors())
	if passwordFound {
		score += 0.4
	}

	// 检查提交按钮
	submitFound := pd.checkElementsExist(ctx, pd.config.GetSubmitSelectors())
	if submitFound {
		score += 0.2
	}

	return score
}

// checkElementsExist 检查元素是否存在
func (pd *PageDetector) checkElementsExist(ctx context.Context, selectors []string) bool {
	for _, selector := range selectors {
		var nodes []*cdp.Node
		err := chromedp.Run(ctx, chromedp.Nodes(selector, &nodes))
		if err == nil && len(nodes) > 0 {
			return true
		}
	}
	return false
}

// DetectLoginForm 检测登录表单元素
func (pd *PageDetector) DetectLoginForm() (*LoginFormElements, error) {
	startTime := time.Now()

	elements := &LoginFormElements{}

	// 检测用户名输入框
	usernameSelector, err := pd.browser.FindElement(pd.config.GetUsernameSelectors())
	if err == nil && usernameSelector != "" {
		elements.UsernameSelector = usernameSelector
		pd.logger.Debugf("✅ 发现用户名输入框: %s", usernameSelector)
	} else {
		pd.logger.Warn("⚠️ 未找到用户名输入框")
	}

	// 检测密码输入框
	passwordSelector, err := pd.browser.FindElement(pd.config.GetPasswordSelectors())
	if err == nil && passwordSelector != "" {
		elements.PasswordSelector = passwordSelector
		pd.logger.Debugf("✅ 发现密码输入框: %s", passwordSelector)
	} else {
		pd.logger.Warn("⚠️ 未找到密码输入框")
	}

	// 检测验证码
	captchaInfo, err := pd.captchaDetector.DetectCaptcha()
	if err == nil && captchaInfo != nil && captchaInfo.Type != CaptchaTypeNone {
		elements.HasCaptcha = true
		elements.CaptchaInfo = captchaInfo
		elements.CaptchaSelector = captchaInfo.Selector

		// 如果有验证码输入框，也尝试找到它
		if captchaSelector, err := pd.browser.FindElement(pd.config.GetCaptchaSelectors()); err == nil && captchaSelector != "" {
			elements.CaptchaSelector = captchaSelector
		}
	}

	// 检测提交按钮
	submitSelector, err := pd.browser.FindElement(pd.config.GetSubmitSelectors())
	if err == nil && submitSelector != "" {
		elements.SubmitSelector = submitSelector
		pd.logger.Debugf("✅ 发现提交按钮: %s", submitSelector)
	} else {
		pd.logger.Warn("⚠️ 未找到提交按钮")
	}

	detectTime := time.Since(startTime)
	pd.logger.Debugf("表单元素检测完成，用时: %v", detectTime)

	return elements, nil
}

// AnalyzePage 分析页面（增强版，包含源码）
func (pd *PageDetector) AnalyzePage() (*PageAnalysis, error) {
	startTime := time.Now()

	analysis := &PageAnalysis{
		DetectedFeatures: []string{},
		ResponseHeaders:  make(map[string]string),
	}

	// 设置网络监听来获取响应头
	ctx := pd.browser.GetContext()

	// 启用网络域
	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		pd.logger.Warnf("启用网络监听失败: %v", err)
	}

	// 监听响应头
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventResponseReceived:
			if strings.Contains(ev.Response.URL, "login") || ev.Type == network.ResourceTypeDocument {
				for key, value := range ev.Response.Headers {
					if key == "content-type" || key == "Content-Type" {
						analysis.ResponseHeaders["content-type"] = fmt.Sprintf("%v", value)
					}
					if key == "content-encoding" || key == "Content-Encoding" {
						analysis.ResponseHeaders["content-encoding"] = fmt.Sprintf("%v", value)
					}
				}
			}
		}
	})

	// 等待页面完全加载
	analyzeCtx, cancel := context.WithTimeout(ctx, pd.analysisTimeout)
	defer cancel()

	var title, url, content, pageSource string
	err := chromedp.Run(analyzeCtx,
		// 等待页面加载完成
		chromedp.Sleep(2*time.Second),
		// 获取基本信息
		chromedp.Title(&title),
		chromedp.Location(&url),
		chromedp.Text("body", &content),
		// 获取完整页面源码
		chromedp.OuterHTML("html", &pageSource),
	)

	if err != nil {
		analysis.ErrorMessage = fmt.Sprintf("获取页面信息失败: %v", err)
		pd.logger.Errorf("页面分析失败: %v", err)
		return analysis, err
	}

	// 检测页面编码
	encoding := pd.detectEncoding(pageSource, analysis.ResponseHeaders)
	analysis.Encoding = encoding

	// 如果需要，转换编码
	if encoding != "utf-8" && encoding != "UTF-8" {
		if convertedSource, err := pd.convertEncoding(pageSource, encoding); err == nil {
			pageSource = convertedSource
			pd.logger.Debugf("页面编码已从 %s 转换为 UTF-8", encoding)
		}
	}

	// 基本信息
	analysis.Title = title
	analysis.URL = url
	analysis.PageSource = pageSource
	analysis.LoadTime = time.Since(startTime)

	// 登录页面检测
	confidence := pd.calculateLoginConfidence(title, url, content, analyzeCtx)
	analysis.Confidence = confidence
	analysis.IsLogin = confidence >= 0.6

	// 特征检测
	if analysis.IsLogin {
		analysis.DetectedFeatures = append(analysis.DetectedFeatures, "登录页面")
	}

	// 检测表单元素
	if formElements, err := pd.DetectLoginForm(); err == nil {
		analysis.FormElements = formElements

		if formElements.UsernameSelector != "" {
			analysis.DetectedFeatures = append(analysis.DetectedFeatures, "用户名输入框")
		}
		if formElements.PasswordSelector != "" {
			analysis.DetectedFeatures = append(analysis.DetectedFeatures, "密码输入框")
		}
		if formElements.HasCaptcha {
			analysis.DetectedFeatures = append(analysis.DetectedFeatures, "验证码")
		}
		if formElements.SubmitSelector != "" {
			analysis.DetectedFeatures = append(analysis.DetectedFeatures, "提交按钮")
		}
	}

	pd.logger.Infof("✅ 页面分析完成，用时: %v, 置信度: %.2f", analysis.LoadTime, analysis.Confidence)

	return analysis, nil
}

// detectEncoding 检测页面编码
func (pd *PageDetector) detectEncoding(pageSource string, headers map[string]string) string {
	// 1. 从HTTP响应头检测
	if contentType, exists := headers["content-type"]; exists {
		if matched := regexp.MustCompile(`charset=([^;]+)`).FindStringSubmatch(contentType); len(matched) > 1 {
			encoding := strings.TrimSpace(matched[1])
			pd.logger.Debugf("从HTTP头检测到编码: %s", encoding)
			return encoding
		}
	}

	// 2. 从HTML meta标签检测
	metaPatterns := []string{
		`<meta\s+charset=["']?([^"'>\s]+)["']?`,
		`<meta\s+http-equiv=["']?content-type["']?\s+content=["']?[^"']*charset=([^"'>\s]+)["']?`,
		`<meta\s+content=["']?[^"']*charset=([^"'>\s]+)["']?\s+http-equiv=["']?content-type["']?`,
	}

	pageSourceLower := strings.ToLower(pageSource)
	for _, pattern := range metaPatterns {
		if matched := regexp.MustCompile(pattern).FindStringSubmatch(pageSourceLower); len(matched) > 1 {
			encoding := strings.TrimSpace(matched[1])
			pd.logger.Debugf("从HTML meta标签检测到编码: %s", encoding)
			return encoding
		}
	}

	// 3. 根据内容特征推测编码
	if strings.Contains(pageSource, "中文") || strings.Contains(pageSource, "登录") {
		if isGBK(pageSource) {
			pd.logger.Debug("根据内容特征推测编码: GBK")
			return "GBK"
		}
	}

	// 4. 默认编码
	pd.logger.Debug("使用默认编码: UTF-8")
	return "UTF-8"
}

// isGBK 简单判断是否可能是GBK编码
func isGBK(content string) bool {
	// 通过字节模式简单判断
	bytes := []byte(content)
	gbkCount := 0
	for i := 0; i < len(bytes)-1; i++ {
		if bytes[i] >= 0x81 && bytes[i] <= 0xFE && bytes[i+1] >= 0x40 && bytes[i+1] <= 0xFE {
			gbkCount++
		}
	}
	return gbkCount > len(bytes)/20 // 如果GBK特征字节超过5%
}

// convertEncoding 转换编码
func (pd *PageDetector) convertEncoding(content, fromEncoding string) (string, error) {
	fromEncoding = strings.ToUpper(fromEncoding)

	switch fromEncoding {
	case "GBK", "GB2312", "GB18030":
		if decoder := simplifiedchinese.GBK.NewDecoder(); decoder != nil {
			result, err := decoder.String(content)
			if err == nil {
				return result, nil
			}
		}
	case "BIG5":
		if decoder := traditionalchinese.Big5.NewDecoder(); decoder != nil {
			result, err := decoder.String(content)
			if err == nil {
				return result, nil
			}
		}
	case "SHIFT_JIS", "SHIFT-JIS":
		if decoder := japanese.ShiftJIS.NewDecoder(); decoder != nil {
			result, err := decoder.String(content)
			if err == nil {
				return result, nil
			}
		}
	case "ISO-2022-JP":
		if decoder := japanese.ISO2022JP.NewDecoder(); decoder != nil {
			result, err := decoder.String(content)
			if err == nil {
				return result, nil
			}
		}
	case "EUC-KR":
		if decoder := korean.EUCKR.NewDecoder(); decoder != nil {
			result, err := decoder.String(content)
			if err == nil {
				return result, nil
			}
		}
	case "ISO-8859-1", "LATIN1":
		if decoder := charmap.ISO8859_1.NewDecoder(); decoder != nil {
			result, err := decoder.String(content)
			if err == nil {
				return result, nil
			}
		}
	default:
		return content, fmt.Errorf("不支持的编码: %s", fromEncoding)
	}

	// 如果转换失败，返回原内容
	return content, nil
}
