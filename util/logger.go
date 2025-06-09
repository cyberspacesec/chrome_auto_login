package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	Logger  *logrus.Logger
	logFile *os.File
)

// LogConfig 日志配置
type LogConfig struct {
	Level          string
	SaveToFile     bool
	LogDir         string
	FilenameFormat string
	MaxFiles       int
	MaxSize        int
	RotateByDate   bool
}

// SimpleFormatter 简洁的日志格式器
type SimpleFormatter struct{}

// Format 实现自定义格式
func (f *SimpleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	level := strings.ToUpper(entry.Level.String())

	// 格式: LEVEL[timestamp] message
	return []byte(fmt.Sprintf("%s[%s] %s\n", level, timestamp, entry.Message)), nil
}

// InitLogger 初始化日志器
func InitLogger(config LogConfig) error {
	Logger = logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// 设置输出格式 - 使用简洁的格式
	Logger.SetFormatter(&SimpleFormatter{})

	// 如果需要保存到文件
	if config.SaveToFile {
		err := setupFileLogging(config)
		if err != nil {
			return fmt.Errorf("设置文件日志失败: %v", err)
		}
	}

	return nil
}

// setupFileLogging 设置文件日志
func setupFileLogging(config LogConfig) error {
	// 创建日志目录
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 生成日志文件名
	filename := time.Now().Format(config.FilenameFormat)
	logPath := filepath.Join(config.LogDir, filename)

	// 打开日志文件
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	logFile = file

	// 设置多重输出（同时输出到控制台和文件）
	multiWriter := io.MultiWriter(os.Stdout, file)
	Logger.SetOutput(multiWriter)

	// 清理旧日志文件
	if config.MaxFiles > 0 {
		_ = cleanOldLogFiles(config.LogDir, config.MaxFiles)
	}

	return nil
}

// cleanOldLogFiles 清理旧日志文件
func cleanOldLogFiles(logDir string, maxFiles int) error {
	files, err := filepath.Glob(filepath.Join(logDir, "*.log"))
	if err != nil {
		return err
	}

	if len(files) <= maxFiles {
		return nil
	}

	// 按文件修改时间排序
	sort.Slice(files, func(i, j int) bool {
		stat1, _ := os.Stat(files[i])
		stat2, _ := os.Stat(files[j])
		return stat1.ModTime().Before(stat2.ModTime())
	})

	// 删除最旧的文件
	for i := 0; i < len(files)-maxFiles; i++ {
		_ = os.Remove(files[i])
	}

	return nil
}

// CloseLogger 关闭日志器
func CloseLogger() {
	if logFile != nil {
		_ = logFile.Close()
	}
}

// LogInfo 记录信息日志
func LogInfo(message string) {
	if Logger != nil {
		Logger.Info(message)
	}
}

// LogError 记录错误日志
func LogError(message string) {
	if Logger != nil {
		Logger.Error(message)
	}
}

// LogWarn 记录警告日志
func LogWarn(message string) {
	if Logger != nil {
		Logger.Warn(message)
	}
}

// LogDebug 记录调试日志
func LogDebug(message string) {
	if Logger != nil {
		Logger.Debug(message)
	}
}

// LogSuccess 记录成功日志
func LogSuccess(message string) {
	if Logger != nil {
		Logger.WithField("type", "success").Info(message)
	}
}

// LogFailure 记录失败日志
func LogFailure(message string) {
	if Logger != nil {
		Logger.WithField("type", "failure").Warn(message)
	}
}

// ResultLogger 结果记录器
type ResultLogger struct {
	saveDir               string
	successFilenameFormat string
	failureFilenameFormat string
	format                string
	realtimeSave          bool
}

// NewResultLogger 创建结果记录器
func NewResultLogger(saveDir, successFormat, failureFormat, format string, realtime bool) *ResultLogger {
	// 创建结果目录
	_ = os.MkdirAll(saveDir, 0755)

	return &ResultLogger{
		saveDir:               saveDir,
		successFilenameFormat: successFormat,
		failureFilenameFormat: failureFormat,
		format:                format,
		realtimeSave:          realtime,
	}
}

// LogSuccess 记录成功结果
func (rl *ResultLogger) LogSuccess(url, username, password string) error {
	if !rl.realtimeSave || rl.successFilenameFormat == "" {
		return nil
	}

	filename := time.Now().Format(rl.successFilenameFormat)
	filePath := filepath.Join(rl.saveDir, filename)

	var content string
	switch rl.format {
	case "url:username:password":
		content = fmt.Sprintf("%s:%s:%s\n", url, username, password)
	default:
		content = fmt.Sprintf("%s:%s:%s\n", url, username, password)
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	_, err = file.WriteString(content)
	return err
}

// LogFailure 记录失败结果
func (rl *ResultLogger) LogFailure(url, username, password string) error {
	if !rl.realtimeSave || rl.failureFilenameFormat == "" {
		return nil
	}

	filename := time.Now().Format(rl.failureFilenameFormat)
	filePath := filepath.Join(rl.saveDir, filename)

	var content string
	switch rl.format {
	case "url:username:password":
		content = fmt.Sprintf("%s:%s:%s\n", url, username, password)
	default:
		content = fmt.Sprintf("%s:%s:%s\n", url, username, password)
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	_, err = file.WriteString(content)
	return err
}

// ProgressAwareLogger 支持进度条的日志记录器
type ProgressAwareLogger struct {
	statusDisplay *StatusDisplay
}

// NewProgressAwareLogger 创建支持进度条的日志记录器
func NewProgressAwareLogger(statusDisplay *StatusDisplay) *ProgressAwareLogger {
	return &ProgressAwareLogger{
		statusDisplay: statusDisplay,
	}
}

// Info 输出信息日志
func (pal *ProgressAwareLogger) Info(message string) {
	pal.logWithProgressManagement(func() {
		LogInfo(message)
	})
}

// Error 输出错误日志
func (pal *ProgressAwareLogger) Error(message string) {
	pal.logWithProgressManagement(func() {
		LogError(message)
	})
}

// Warn 输出警告日志
func (pal *ProgressAwareLogger) Warn(message string) {
	pal.logWithProgressManagement(func() {
		LogWarn(message)
	})
}

// Debug 输出调试日志
func (pal *ProgressAwareLogger) Debug(message string) {
	pal.logWithProgressManagement(func() {
		LogDebug(message)
	})
}

// logWithProgressManagement 统一的进度条管理日志输出
func (pal *ProgressAwareLogger) logWithProgressManagement(logFunc func()) {
	// 如果有进度条并且是浮动模式，需要特殊处理
	if pal.statusDisplay != nil && pal.statusDisplay.progressBar != nil && pal.statusDisplay.progressBar.isFloating {
		// 清除底部进度条
		pal.statusDisplay.progressBar.Clear()

		// 执行日志输出
		logFunc()

		// 延迟重绘进度条
		pal.statusDisplay.progressBar.ForceRedraw()
	} else {
		// 普通模式直接输出
		logFunc()
	}
}
