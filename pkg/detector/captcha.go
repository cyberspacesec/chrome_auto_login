package detector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"

	"github.com/cyberspacesec/chrome_auto_login/pkg/browser"
	"github.com/cyberspacesec/chrome_auto_login/pkg/config"
)

// CaptchaType 验证码类型
type CaptchaType int

const (
	CaptchaTypeNone      CaptchaType = iota // 无验证码
	CaptchaTypeText                         // 文字验证码
	CaptchaTypeImage                        // 图片验证码
	CaptchaTypeSlider                       // 滑块验证码
	CaptchaTypeClick                        // 点击验证码
	CaptchaTypeRecaptcha                    // Google reCAPTCHA
	CaptchaTypeHCaptcha                     // hCaptcha
	CaptchaTypeBehavior                     // 行为验证
	CaptchaTypeUnknown                      // 未知类型
)

// CaptchaInfo 验证码信息
type CaptchaInfo struct {
	Type        CaptchaType `json:"type"`
	Selector    string      `json:"selector"`
	ImageURL    string      `json:"image_url"`
	Description string      `json:"description"`
	Element     *cdp.Node   `json:"-"`
	Confidence  float64     `json:"confidence"` // 检测置信度 0-1
}

// CaptchaDetector 验证码检测器
type CaptchaDetector struct {
	browser *browser.Browser
	config  *config.Config
	logger  *logrus.Logger
}

// NewCaptchaDetector 创建验证码检测器
func NewCaptchaDetector(browser *browser.Browser, cfg *config.Config, logger *logrus.Logger) *CaptchaDetector {
	return &CaptchaDetector{
		browser: browser,
		config:  cfg,
		logger:  logger,
	}
}

// DetectCaptcha 检测页面中的验证码
func (cd *CaptchaDetector) DetectCaptcha() (*CaptchaInfo, error) {
	// 创建10秒超时上下文
	detectCtx, detectCancel := context.WithTimeout(cd.browser.GetContext(), 10*time.Second)
	defer detectCancel()

	if cd.config.Captcha.Detection.VerboseOutput {
		cd.logger.Info("🔍 开始检测页面验证码...")
	} else {
		cd.logger.Debug("🔍 开始检测验证码...")
	}

	// 第一步：快速预检测是否存在验证码相关元素
	hasCaptha, err := cd.quickPreDetect(detectCtx)
	if err != nil {
		cd.logger.Warn("验证码预检测失败:", err)
		return &CaptchaInfo{Type: CaptchaTypeNone}, nil
	}

	if !hasCaptha {
		if cd.config.Captcha.Detection.VerboseOutput {
			cd.logger.Info("✅ 未检测到验证码相关元素，页面不包含验证码")
		} else {
			cd.logger.Debug("✅ 未检测到验证码")
		}
		return &CaptchaInfo{Type: CaptchaTypeNone}, nil
	}

	if cd.config.Captcha.Detection.VerboseOutput {
		cd.logger.Info("🎯 检测到验证码相关元素，开始分析验证码类型...")
	}

	// 第二步：详细分析验证码类型（剩余时间内）
	captcha, err := cd.analyzeCaptchaType(detectCtx)
	if err != nil {
		cd.logger.Warn("验证码类型分析失败:", err)
		return &CaptchaInfo{Type: CaptchaTypeUnknown}, nil
	}

	if captcha != nil && captcha.Type != CaptchaTypeNone {
		// 输出检测结果
		typeMsg := fmt.Sprintf("🎯 验证码类型分析完成: %s", captcha.GetTypeName())
		if cd.config.Captcha.Detection.VerboseOutput {
			typeMsg += fmt.Sprintf(" (置信度: %.2f)", captcha.Confidence)
			if captcha.Selector != "" {
				typeMsg += fmt.Sprintf(" (选择器: %s)", captcha.Selector)
			}
			cd.logger.Info(typeMsg)
			cd.logger.Info(fmt.Sprintf("📋 验证码描述: %s", captcha.Description))
			cd.logger.Info(fmt.Sprintf("🛠️ 处理策略: %s", captcha.GetHandlingStrategy()))
		} else {
			cd.logger.Info(typeMsg)
		}
		return captcha, nil
	}

	if cd.config.Captcha.Detection.VerboseOutput {
		cd.logger.Info("⚠️ 检测到验证码元素但无法确定具体类型")
	}
	return &CaptchaInfo{Type: CaptchaTypeUnknown}, nil
}

// quickPreDetect 快速预检测是否存在验证码相关元素
func (cd *CaptchaDetector) quickPreDetect(ctx context.Context) (bool, error) {
	// 创建3秒超时的预检测上下文
	preCtx, preCancel := context.WithTimeout(ctx, 3*time.Second)
	defer preCancel()

	// 所有可能的验证码相关选择器（快速检测）
	quickSelectors := []string{
		// 验证码输入框
		"input[name*='captcha']",
		"input[name*='verify']",
		"input[name*='code']",
		"input[placeholder*='验证码']",
		"input[placeholder*='captcha']",
		"input[placeholder*='verify']",
		"#captcha", "#verify", "#code",
		".captcha", ".verify", ".code",

		// 验证码图片
		"img[src*='captcha']",
		"img[src*='verify']",
		"img[src*='vcode']",
		"img[alt*='验证码']",
		"img[alt*='captcha']",

		// 第三方验证码
		".g-recaptcha",
		".h-captcha",
		"iframe[src*='recaptcha']",
		"iframe[src*='hcaptcha']",

		// 滑块验证码
		".slider-captcha",
		".slide-captcha",
		".geetest_slider",
		".nc_iconfont",
		".yidun_slider",

		// 点击验证码
		"[class*='click'][class*='captcha']",
		"[class*='click'][class*='verify']",
	}

	// 快速检查是否存在任何验证码相关元素
	for _, selector := range quickSelectors {
		select {
		case <-preCtx.Done():
			return false, preCtx.Err()
		default:
			if found, err := cd.quickCheckElement(preCtx, selector); err == nil && found {
				return true, nil
			}
		}
	}

	// 通过页面文本快速检测验证码关键词
	return cd.quickTextDetect(preCtx)
}

// quickCheckElement 快速检查元素是否存在
func (cd *CaptchaDetector) quickCheckElement(ctx context.Context, selector string) (bool, error) {
	var nodes []*cdp.Node
	err := chromedp.Run(ctx, chromedp.Nodes(selector, &nodes, chromedp.ByQuery))
	return len(nodes) > 0, err
}

// quickTextDetect 通过页面文本快速检测验证码
func (cd *CaptchaDetector) quickTextDetect(ctx context.Context) (bool, error) {
	var pageText string
	err := chromedp.Run(ctx, chromedp.Text("body", &pageText, chromedp.ByQuery))
	if err != nil {
		return false, err
	}

	// 验证码关键词（快速检测）
	keywords := []string{
		"验证码", "captcha", "verify", "verification",
		"滑动", "slide", "slider", "拖拽", "drag",
		"点击", "click", "recaptcha", "hcaptcha",
	}

	pageTextLower := strings.ToLower(pageText)
	for _, keyword := range keywords {
		if strings.Contains(pageTextLower, strings.ToLower(keyword)) {
			return true, nil
		}
	}

	return false, nil
}

// analyzeCaptchaType 分析验证码类型
func (cd *CaptchaDetector) analyzeCaptchaType(ctx context.Context) (*CaptchaInfo, error) {
	// 按优先级检测各种类型的验证码
	detectors := []struct {
		name string
		fn   func(context.Context) (*CaptchaInfo, error)
	}{
		{"reCAPTCHA", cd.detectRecaptcha},
		{"hCaptcha", cd.detectHCaptcha},
		{"滑块验证码", cd.detectSliderCaptcha},
		{"图片验证码", cd.detectImageCaptcha},
		{"文字验证码", cd.detectTextCaptcha},
		{"点击验证码", cd.detectClickCaptcha},
		{"行为验证码", cd.detectBehaviorCaptcha},
	}

	for _, detector := range detectors {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// 为每个检测器分配最多1.5秒
			detectorCtx, detectorCancel := context.WithTimeout(ctx, 1500*time.Millisecond)
			captcha, err := detector.fn(detectorCtx)
			detectorCancel()

			if err == nil && captcha != nil && captcha.Type != CaptchaTypeNone {
				return captcha, nil
			}
		}
	}

	return &CaptchaInfo{Type: CaptchaTypeNone}, nil
}

// detectRecaptcha 检测Google reCAPTCHA
func (cd *CaptchaDetector) detectRecaptcha(ctx context.Context) (*CaptchaInfo, error) {
	selectors := []string{
		".g-recaptcha",
		"#recaptcha",
		"[data-sitekey]",
		".recaptcha-checkbox",
		"iframe[src*='recaptcha']",
		"iframe[title*='reCAPTCHA']",
		"#g-recaptcha-response",
	}

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(ctx, selector); err == nil && found {
			return &CaptchaInfo{
				Type:        CaptchaTypeRecaptcha,
				Selector:    selector,
				Description: "Google reCAPTCHA 验证码 - 第三方智能验证服务",
				Confidence:  0.95,
			}, nil
		}
	}

	return nil, nil
}

// detectHCaptcha 检测hCaptcha
func (cd *CaptchaDetector) detectHCaptcha(ctx context.Context) (*CaptchaInfo, error) {
	selectors := []string{
		".h-captcha",
		"[data-hcaptcha-sitekey]",
		"iframe[src*='hcaptcha']",
		"iframe[title*='hCaptcha']",
		"#h-captcha-response",
	}

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(ctx, selector); err == nil && found {
			return &CaptchaInfo{
				Type:        CaptchaTypeHCaptcha,
				Selector:    selector,
				Description: "hCaptcha 验证码 - 第三方智能验证服务",
				Confidence:  0.95,
			}, nil
		}
	}

	return nil, nil
}

// detectSliderCaptcha 检测滑块验证码
func (cd *CaptchaDetector) detectSliderCaptcha(ctx context.Context) (*CaptchaInfo, error) {
	// 滑块验证码的常见特征
	selectors := []string{
		".slider-captcha",
		".slide-captcha",
		".captcha-slider",
		"[class*='slider'][class*='captcha']",
		"[class*='slide'][class*='verify']",
		".geetest_slider",
		".nc_iconfont",
		".yidun_slider",
		".captcha-drag",
		"[class*='drag'][class*='verify']",
	}

	keywords := []string{
		"滑动", "slide", "slider", "拖拽", "drag",
		"向右滑动", "slide to verify", "拖动完成验证",
		"请完成滑动验证", "请拖动滑块",
	}

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(ctx, selector); err == nil && found {
			return &CaptchaInfo{
				Type:        CaptchaTypeSlider,
				Selector:    selector,
				Description: "滑块验证码 - 需要拖动滑块完成验证",
				Confidence:  0.9,
			}, nil
		}
	}

	// 通过文本内容检测
	if confidence := cd.detectByKeywords(ctx, keywords); confidence > 0.7 {
		return &CaptchaInfo{
			Type:        CaptchaTypeSlider,
			Description: "滑块验证码 - 通过页面文本内容检测到",
			Confidence:  confidence,
		}, nil
	}

	return nil, nil
}

// detectImageCaptcha 检测图片验证码
func (cd *CaptchaDetector) detectImageCaptcha(ctx context.Context) (*CaptchaInfo, error) {
	selectors := []string{
		"img[src*='captcha']",
		"img[src*='verify']",
		"img[src*='vcode']",
		"img[alt*='验证码']",
		"img[alt*='captcha']",
		".captcha-image",
		".verify-image",
		"#captcha_img",
		"#verify_img",
		".code-img",
		"[class*='captcha'][class*='img']",
	}

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(ctx, selector); err == nil && found {
			imageURL, _ := cd.getImageURL(ctx, selector)
			return &CaptchaInfo{
				Type:        CaptchaTypeImage,
				Selector:    selector,
				ImageURL:    imageURL,
				Description: "图片验证码 - 需要识别图片中的字符",
				Confidence:  0.85,
			}, nil
		}
	}

	return nil, nil
}

// detectTextCaptcha 检测文字验证码
func (cd *CaptchaDetector) detectTextCaptcha(ctx context.Context) (*CaptchaInfo, error) {
	// 文字验证码通常通过输入框检测
	selectors := cd.config.GetCaptchaSelectors()

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(ctx, selector); err == nil && found {
			return &CaptchaInfo{
				Type:        CaptchaTypeText,
				Selector:    selector,
				Description: "文字验证码 - 需要输入验证码字符",
				Confidence:  0.8,
			}, nil
		}
	}

	// 检测验证码关键词
	keywords := []string{
		"验证码", "captcha", "verify code", "verification code",
		"图形验证码", "图片验证码", "security code",
	}

	if confidence := cd.detectByKeywords(ctx, keywords); confidence > 0.6 {
		return &CaptchaInfo{
			Type:        CaptchaTypeText,
			Description: "文字验证码 - 通过页面文本内容检测到",
			Confidence:  confidence,
		}, nil
	}

	return nil, nil
}

// detectClickCaptcha 检测点击验证码
func (cd *CaptchaDetector) detectClickCaptcha(ctx context.Context) (*CaptchaInfo, error) {
	selectors := []string{
		"[class*='click'][class*='captcha']",
		"[class*='click'][class*='verify']",
		".captcha-click",
		".click-verify",
		"[id*='click'][id*='captcha']",
		"[data-click*='verify']",
	}

	keywords := []string{
		"点击", "click", "请点击", "按顺序点击",
		"点击验证", "click to verify", "click captcha",
		"请按顺序点击", "点击文字",
	}

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(ctx, selector); err == nil && found {
			return &CaptchaInfo{
				Type:        CaptchaTypeClick,
				Selector:    selector,
				Description: "点击验证码 - 需要按要求点击指定位置",
				Confidence:  0.85,
			}, nil
		}
	}

	// 通过文本内容检测
	if confidence := cd.detectByKeywords(ctx, keywords); confidence > 0.7 {
		return &CaptchaInfo{
			Type:        CaptchaTypeClick,
			Description: "点击验证码 - 通过页面文本内容检测到",
			Confidence:  confidence,
		}, nil
	}

	return nil, nil
}

// detectBehaviorCaptcha 检测行为验证码
func (cd *CaptchaDetector) detectBehaviorCaptcha(ctx context.Context) (*CaptchaInfo, error) {
	selectors := []string{
		"[class*='behavior'][class*='captcha']",
		"[class*='behavior'][class*='verify']",
		".captcha-behavior",
		".behavior-verify",
		"[data-behavior*='verify']",
	}

	keywords := []string{
		"行为验证", "behavior", "智能验证", "无感验证",
		"人机验证", "bot detection", "智能识别",
	}

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(ctx, selector); err == nil && found {
			return &CaptchaInfo{
				Type:        CaptchaTypeBehavior,
				Selector:    selector,
				Description: "行为验证码 - 基于用户行为模式的智能验证",
				Confidence:  0.8,
			}, nil
		}
	}

	// 通过文本内容检测
	if confidence := cd.detectByKeywords(ctx, keywords); confidence > 0.6 {
		return &CaptchaInfo{
			Type:        CaptchaTypeBehavior,
			Description: "行为验证码 - 通过页面文本内容检测到",
			Confidence:  confidence,
		}, nil
	}

	return nil, nil
}

// checkElementExists 检查元素是否存在
func (cd *CaptchaDetector) checkElementExists(ctx context.Context, selector string) (bool, error) {
	checkCtx, checkCancel := context.WithTimeout(ctx, 1*time.Second)
	defer checkCancel()

	var nodes []*cdp.Node
	err := chromedp.Run(checkCtx, chromedp.Nodes(selector, &nodes))
	return len(nodes) > 0, err
}

// detectByKeywords 通过关键词检测验证码
func (cd *CaptchaDetector) detectByKeywords(ctx context.Context, keywords []string) float64 {
	keywordCtx, keywordCancel := context.WithTimeout(ctx, 1*time.Second)
	defer keywordCancel()

	var pageText string
	err := chromedp.Run(keywordCtx, chromedp.Text("body", &pageText))
	if err != nil {
		return 0
	}

	pageText = strings.ToLower(pageText)
	matchCount := 0
	for _, keyword := range keywords {
		if strings.Contains(pageText, strings.ToLower(keyword)) {
			matchCount++
		}
	}

	if len(keywords) == 0 {
		return 0
	}

	confidence := float64(matchCount) / float64(len(keywords))
	if matchCount > 0 {
		confidence = confidence*0.8 + 0.2 // 最少0.2的置信度
	}

	return confidence
}

// getImageURL 获取图片URL
func (cd *CaptchaDetector) getImageURL(ctx context.Context, selector string) (string, error) {
	urlCtx, urlCancel := context.WithTimeout(ctx, 1*time.Second)
	defer urlCancel()

	var imageURL string
	err := chromedp.Run(urlCtx, chromedp.AttributeValue(selector, "src", &imageURL, nil))
	return imageURL, err
}

// GetTypeName 获取验证码类型名称
func (ci *CaptchaInfo) GetTypeName() string {
	switch ci.Type {
	case CaptchaTypeNone:
		return "无验证码"
	case CaptchaTypeText:
		return "文字验证码"
	case CaptchaTypeImage:
		return "图片验证码"
	case CaptchaTypeSlider:
		return "滑块验证码"
	case CaptchaTypeClick:
		return "点击验证码"
	case CaptchaTypeRecaptcha:
		return "Google reCAPTCHA"
	case CaptchaTypeHCaptcha:
		return "hCaptcha"
	case CaptchaTypeBehavior:
		return "行为验证码"
	case CaptchaTypeUnknown:
		return "未知验证码"
	default:
		return "未知类型"
	}
}

// IsInteractive 检查验证码是否需要交互处理
func (ci *CaptchaInfo) IsInteractive() bool {
	switch ci.Type {
	case CaptchaTypeRecaptcha, CaptchaTypeHCaptcha:
		return true // 第三方验证码通常需要人工交互
	case CaptchaTypeSlider, CaptchaTypeClick:
		return true // 滑块和点击验证码需要模拟交互
	case CaptchaTypeBehavior:
		return true // 行为验证码需要特殊处理
	case CaptchaTypeText, CaptchaTypeImage:
		return false // 文字和图片验证码可以通过OCR处理
	default:
		return true // 未知类型默认需要交互
	}
}

// GetHandlingStrategy 获取处理策略说明
func (ci *CaptchaInfo) GetHandlingStrategy() string {
	switch ci.Type {
	case CaptchaTypeNone:
		return "无需处理"
	case CaptchaTypeText:
		return "可通过OCR识别自动处理"
	case CaptchaTypeImage:
		return "可通过OCR识别自动处理"
	case CaptchaTypeSlider:
		return "可通过模拟拖拽操作自动处理"
	case CaptchaTypeClick:
		return "可通过模拟点击操作自动处理"
	case CaptchaTypeRecaptcha:
		return "需要人工交互或第三方服务处理"
	case CaptchaTypeHCaptcha:
		return "需要人工交互或第三方服务处理"
	case CaptchaTypeBehavior:
		return "需要分析行为模式进行特殊处理"
	case CaptchaTypeUnknown:
		return "需要进一步分析或人工处理"
	default:
		return "未知处理策略"
	}
}
