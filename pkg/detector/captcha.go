package detector

import (
	"context"
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
	CaptchaTypeNone        CaptchaType = iota // 无验证码
	CaptchaTypeText                           // 文字验证码
	CaptchaTypeImage                          // 图片验证码
	CaptchaTypeSlider                         // 滑块验证码
	CaptchaTypeClick                          // 点击验证码
	CaptchaTypeRecaptcha                      // Google reCAPTCHA
	CaptchaTypeHCaptcha                       // hCaptcha
	CaptchaTypeBehavior                       // 行为验证
	CaptchaTypeUnknown                        // 未知类型
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
	cd.logger.Debug("🔍 开始检测验证码...")

	// 检测各种类型的验证码
	detectors := []func() (*CaptchaInfo, error){
		cd.detectRecaptcha,
		cd.detectHCaptcha,
		cd.detectSliderCaptcha,
		cd.detectImageCaptcha,
		cd.detectTextCaptcha,
		cd.detectClickCaptcha,
		cd.detectBehaviorCaptcha,
	}

	for _, detector := range detectors {
		if captcha, err := detector(); err == nil && captcha != nil && captcha.Type != CaptchaTypeNone {
			cd.logger.Infof("🎯 检测到验证码: %s (置信度: %.2f)", captcha.GetTypeName(), captcha.Confidence)
			return captcha, nil
		}
	}

	cd.logger.Debug("✅ 未检测到验证码")
	return &CaptchaInfo{Type: CaptchaTypeNone}, nil
}

// detectRecaptcha 检测Google reCAPTCHA
func (cd *CaptchaDetector) detectRecaptcha() (*CaptchaInfo, error) {
	selectors := []string{
		".g-recaptcha",
		"#recaptcha",
		"[data-sitekey]",
		".recaptcha-checkbox",
		"iframe[src*='recaptcha']",
	}

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(selector); err == nil && found {
			return &CaptchaInfo{
				Type:        CaptchaTypeRecaptcha,
				Selector:    selector,
				Description: "Google reCAPTCHA 验证码",
				Confidence:  0.95,
			}, nil
		}
	}

	return nil, nil
}

// detectHCaptcha 检测hCaptcha
func (cd *CaptchaDetector) detectHCaptcha() (*CaptchaInfo, error) {
	selectors := []string{
		".h-captcha",
		"[data-hcaptcha-sitekey]",
		"iframe[src*='hcaptcha']",
	}

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(selector); err == nil && found {
			return &CaptchaInfo{
				Type:        CaptchaTypeHCaptcha,
				Selector:    selector,
				Description: "hCaptcha 验证码",
				Confidence:  0.95,
			}, nil
		}
	}

	return nil, nil
}

// detectSliderCaptcha 检测滑块验证码
func (cd *CaptchaDetector) detectSliderCaptcha() (*CaptchaInfo, error) {
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
	}

	keywords := []string{
		"滑动", "slide", "slider", "拖拽", "drag",
		"向右滑动", "slide to verify", "拖动完成验证",
	}

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(selector); err == nil && found {
			return &CaptchaInfo{
				Type:        CaptchaTypeSlider,
				Selector:    selector,
				Description: "滑块验证码",
				Confidence:  0.9,
			}, nil
		}
	}

	// 通过文本内容检测
	if confidence := cd.detectByKeywords(keywords); confidence > 0.7 {
		return &CaptchaInfo{
			Type:        CaptchaTypeSlider,
			Description: "滑块验证码 (通过文本检测)",
			Confidence:  confidence,
		}, nil
	}

	return nil, nil
}

// detectImageCaptcha 检测图片验证码
func (cd *CaptchaDetector) detectImageCaptcha() (*CaptchaInfo, error) {
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
	}

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(selector); err == nil && found {
			imageURL, _ := cd.getImageURL(selector)
			return &CaptchaInfo{
				Type:        CaptchaTypeImage,
				Selector:    selector,
				ImageURL:    imageURL,
				Description: "图片验证码",
				Confidence:  0.85,
			}, nil
		}
	}

	return nil, nil
}

// detectTextCaptcha 检测文字验证码
func (cd *CaptchaDetector) detectTextCaptcha() (*CaptchaInfo, error) {
	// 检测验证码输入框附近的文字
	selectors := []string{
		"input[placeholder*='验证码']",
		"input[placeholder*='captcha']",
		"input[placeholder*='verify']",
		"input[name*='captcha']",
		"input[name*='verify']",
		"input[name*='vcode']",
		"input[id*='captcha']",
		"input[id*='verify']",
	}

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(selector); err == nil && found {
			return &CaptchaInfo{
				Type:        CaptchaTypeText,
				Selector:    selector,
				Description: "文字验证码输入框",
				Confidence:  0.8,
			}, nil
		}
	}

	return nil, nil
}

// detectClickCaptcha 检测点击验证码
func (cd *CaptchaDetector) detectClickCaptcha() (*CaptchaInfo, error) {
	keywords := []string{
		"点击", "click", "按顺序点击", "请点击",
		"点击图片", "click on", "select images",
	}

	selectors := []string{
		".click-captcha",
		".image-select",
		"[class*='click'][class*='captcha']",
	}

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(selector); err == nil && found {
			return &CaptchaInfo{
				Type:        CaptchaTypeClick,
				Selector:    selector,
				Description: "点击验证码",
				Confidence:  0.8,
			}, nil
		}
	}

	if confidence := cd.detectByKeywords(keywords); confidence > 0.6 {
		return &CaptchaInfo{
			Type:        CaptchaTypeClick,
			Description: "点击验证码 (通过文本检测)",
			Confidence:  confidence,
		}, nil
	}

	return nil, nil
}

// detectBehaviorCaptcha 检测行为验证
func (cd *CaptchaDetector) detectBehaviorCaptcha() (*CaptchaInfo, error) {
	selectors := []string{
		".behavior-captcha",
		".intelligent-captcha",
		"[class*='behavior'][class*='verify']",
	}

	keywords := []string{
		"行为验证", "behavior", "intelligent",
		"智能验证", "无感验证",
	}

	for _, selector := range selectors {
		if found, err := cd.checkElementExists(selector); err == nil && found {
			return &CaptchaInfo{
				Type:        CaptchaTypeBehavior,
				Selector:    selector,
				Description: "行为验证码",
				Confidence:  0.7,
			}, nil
		}
	}

	if confidence := cd.detectByKeywords(keywords); confidence > 0.5 {
		return &CaptchaInfo{
			Type:        CaptchaTypeBehavior,
			Description: "行为验证码 (通过文本检测)",
			Confidence:  confidence,
		}, nil
	}

	return nil, nil
}

// checkElementExists 检查元素是否存在
func (cd *CaptchaDetector) checkElementExists(selector string) (bool, error) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var nodes []*cdp.Node
	err := chromedp.Run(timeoutCtx,
		chromedp.Nodes(selector, &nodes, chromedp.AtLeast(0)),
	)

	return err == nil && len(nodes) > 0, err
}

// detectByKeywords 通过关键词检测
func (cd *CaptchaDetector) detectByKeywords(keywords []string) float64 {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var content string
	err := chromedp.Run(timeoutCtx,
		chromedp.Text("body", &content, chromedp.ByQuery),
	)

	if err != nil {
		return 0
	}

	content = strings.ToLower(content)
	matchCount := 0
	
	for _, keyword := range keywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			matchCount++
		}
	}

	confidence := float64(matchCount) / float64(len(keywords))
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// getImageURL 获取图片URL
func (cd *CaptchaDetector) getImageURL(selector string) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var src string
	err := chromedp.Run(timeoutCtx,
		chromedp.AttributeValue(selector, "src", &src, nil),
	)

	return src, err
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
		return "行为验证"
	default:
		return "未知验证码"
	}
}

// IsInteractive 判断是否需要人工交互
func (ci *CaptchaInfo) IsInteractive() bool {
	switch ci.Type {
	case CaptchaTypeNone:
		return false
	case CaptchaTypeText:
		return true // 需要OCR或人工输入
	case CaptchaTypeImage:
		return true // 需要OCR或人工输入
	case CaptchaTypeSlider:
		return true // 需要模拟滑动
	case CaptchaTypeClick:
		return true // 需要图像识别和点击
	case CaptchaTypeRecaptcha, CaptchaTypeHCaptcha:
		return true // 需要特殊处理
	case CaptchaTypeBehavior:
		return false // 可能自动通过
	default:
		return true
	}
}

// GetHandlingStrategy 获取处理策略描述
func (ci *CaptchaInfo) GetHandlingStrategy() string {
	switch ci.Type {
	case CaptchaTypeNone:
		return "无需处理"
	case CaptchaTypeText, CaptchaTypeImage:
		return "需要OCR识别或人工输入验证码"
	case CaptchaTypeSlider:
		return "需要模拟滑块拖拽操作"
	case CaptchaTypeClick:
		return "需要图像识别和模拟点击操作"
	case CaptchaTypeRecaptcha:
		return "需要集成reCAPTCHA解决方案"
	case CaptchaTypeHCaptcha:
		return "需要集成hCaptcha解决方案"
	case CaptchaTypeBehavior:
		return "尝试模拟正常用户行为"
	default:
		return "需要人工分析和处理"
	}
} 