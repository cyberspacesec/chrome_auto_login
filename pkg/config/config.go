package config

import (
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Config 主配置结构
type Config struct {
	Browser           BrowserConfig           `yaml:"browser"`
	LoginPageDetection LoginPageDetectionConfig `yaml:"login_page_detection"`
	FormElements      FormElementsConfig      `yaml:"form_elements"`
	Bruteforce        BruteforceConfig        `yaml:"bruteforce"`
	Logging           LoggingConfig           `yaml:"logging"`
	Results           ResultsConfig           `yaml:"results"`
	Captcha           CaptchaConfig           `yaml:"captcha"`
}

// BrowserConfig 浏览器配置
type BrowserConfig struct {
	Headless bool `yaml:"headless"`
	Timeout  int  `yaml:"timeout"`
	Width    int  `yaml:"width"`
	Height   int  `yaml:"height"`
}

// LoginPageDetectionConfig 登录页面检测配置
type LoginPageDetectionConfig struct {
	TitlePatterns   []string `yaml:"title_patterns"`
	URLPatterns     []string `yaml:"url_patterns"`
	ContentKeywords []string `yaml:"content_keywords"`
}

// FormElementsConfig 表单元素配置
type FormElementsConfig struct {
	UsernameSelectors []string `yaml:"username_selectors"`
	PasswordSelectors []string `yaml:"password_selectors"`
	CaptchaSelectors  []string `yaml:"captcha_selectors"`
	SubmitSelectors   []string `yaml:"submit_selectors"`
}

// BruteforceConfig 爆破配置
type BruteforceConfig struct {
	Usernames   []string `yaml:"usernames"`
	Passwords   []string `yaml:"passwords"`
	Delay       int      `yaml:"delay"`
	MaxRetries  int      `yaml:"max_retries"`
	Concurrent  int      `yaml:"concurrent"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level          string              `yaml:"level"`
	File           string              `yaml:"file"`
	Format         string              `yaml:"format"`
	FileManagement LogFileManagement   `yaml:"file_management"`
}

// LogFileManagement 日志文件管理配置
type LogFileManagement struct {
	SaveToFile       bool   `yaml:"save_to_file"`
	LogDir           string `yaml:"log_dir"`
	FilenameFormat   string `yaml:"filename_format"`
	MaxFiles         int    `yaml:"max_files"`
	MaxSize          int    `yaml:"max_size"`
	RotateByDate     bool   `yaml:"rotate_by_date"`
}

// ResultsConfig 结果存储配置
type ResultsConfig struct {
	SaveDir                  string `yaml:"save_dir"`
	SuccessFilenameFormat    string `yaml:"success_filename_format"`
	FailureFilenameFormat    string `yaml:"failure_filename_format"`
	Format                   string `yaml:"format"`
	RealtimeSave             bool   `yaml:"realtime_save"`
}

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	Detection       CaptchaDetectionConfig       `yaml:"detection"`
	Handling        CaptchaHandlingConfig        `yaml:"handling"`
	OCR             CaptchaOCRConfig             `yaml:"ocr"`
	Slider          CaptchaSliderConfig          `yaml:"slider"`
	CustomSelectors CaptchaCustomSelectorsConfig `yaml:"custom_selectors"`
}

// CaptchaDetectionConfig 验证码检测配置
type CaptchaDetectionConfig struct {
	Enabled             bool    `yaml:"enabled"`
	Timeout             int     `yaml:"timeout"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
}

// CaptchaHandlingConfig 验证码处理配置
type CaptchaHandlingConfig struct {
	SkipOnDetection bool `yaml:"skip_on_detection"`
	ManualInput     bool `yaml:"manual_input"`
	OCREnabled      bool `yaml:"ocr_enabled"`
}

// CaptchaOCRConfig OCR配置
type CaptchaOCRConfig struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key"`
	Timeout  int    `yaml:"timeout"`
}

// CaptchaSliderConfig 滑块验证码配置
type CaptchaSliderConfig struct {
	Enabled       bool `yaml:"enabled"`
	SimulateHuman bool `yaml:"simulate_human"`
	DragSpeed     int  `yaml:"drag_speed"`
}

// CaptchaCustomSelectorsConfig 自定义验证码选择器配置
type CaptchaCustomSelectorsConfig struct {
	ImageSelectors  []string `yaml:"image_selectors"`
	InputSelectors  []string `yaml:"input_selectors"`
	SliderSelectors []string `yaml:"slider_selectors"`
}

var globalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}

	// 读取配置文件
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	// 解析YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	globalConfig = &config
	return &config, nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return globalConfig
}

// IsLoginPage 检查是否为登录页面
func (c *Config) IsLoginPage(title, url, content string) bool {
	// 检查标题
	for _, pattern := range c.LoginPageDetection.TitlePatterns {
		if matched, _ := regexp.MatchString(pattern, title); matched {
			return true
		}
	}

	// 检查URL
	for _, pattern := range c.LoginPageDetection.URLPatterns {
		if matched, _ := regexp.MatchString(pattern, url); matched {
			return true
		}
	}

	// 检查内容关键词
	for _, keyword := range c.LoginPageDetection.ContentKeywords {
		if matched, _ := regexp.MatchString("(?i).*"+keyword+".*", content); matched {
			return true
		}
	}

	return false
}

// GetUsernameSelectors 获取用户名选择器
func (c *Config) GetUsernameSelectors() []string {
	return c.FormElements.UsernameSelectors
}

// GetPasswordSelectors 获取密码选择器
func (c *Config) GetPasswordSelectors() []string {
	return c.FormElements.PasswordSelectors
}

// GetCaptchaSelectors 获取验证码选择器
func (c *Config) GetCaptchaSelectors() []string {
	return c.FormElements.CaptchaSelectors
}

// GetSubmitSelectors 获取提交按钮选择器
func (c *Config) GetSubmitSelectors() []string {
	return c.FormElements.SubmitSelectors
}

// GetCredentials 获取用户名密码组合
func (c *Config) GetCredentials() []Credential {
	var credentials []Credential
	for _, username := range c.Bruteforce.Usernames {
		for _, password := range c.Bruteforce.Passwords {
			credentials = append(credentials, Credential{
				Username: username,
				Password: password,
			})
		}
	}
	return credentials
}

// Credential 认证凭据
type Credential struct {
	Username string
	Password string
} 