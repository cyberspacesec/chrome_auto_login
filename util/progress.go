package util

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"strings"
	"sync"
	"time"
)

// ProgressBar 进度条结构
type ProgressBar struct {
	current    int
	total      int
	width      int
	startTime  time.Time
	prefix     string
	mu         sync.Mutex
	finished   bool
	lastLine   string // 保存上次显示的内容
	isFloating bool   // 是否浮动显示
	termHeight int    // 终端高度
	termWidth  int    // 终端宽度
}

// NewProgressBar 创建新的进度条
func NewProgressBar(total int, prefix string) *ProgressBar {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 80, 24 // 默认尺寸
	}

	return &ProgressBar{
		current:    0,
		total:      total,
		width:      50,
		startTime:  time.Now(),
		prefix:     prefix,
		finished:   false,
		isFloating: true, // 默认启用浮动模式
		termWidth:  width,
		termHeight: height,
	}
}

// Update 更新进度
func (pb *ProgressBar) Update(current int, message string) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.finished {
		return
	}

	pb.current = current
	if pb.isFloating {
		pb.renderFloatingProgress(message)
	} else {
		pb.render(message)
	}
}

// Finish 完成进度条
func (pb *ProgressBar) Finish(message string) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.finished = true
	pb.current = pb.total

	if pb.isFloating {
		// 清除浮动进度条并显示完成消息
		pb.clearBottomLine()
		fmt.Printf("✅ %s\n", message)
	} else {
		// 完成时清除进度条并换行
		fmt.Print("\r\033[K")
		fmt.Printf("✅ %s\n", message)
	}
}

// Clear 清除进度条
func (pb *ProgressBar) Clear() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.finished {
		return
	}

	if pb.isFloating {
		// 浮动模式下不清除进度条，保持在底部固定位置
		// 让日志在上方正常输出，互不干扰
		return
	} else {
		// 只清除当前行
		fmt.Print("\r\033[K")
	}
}

// Redraw 重新绘制进度条
func (pb *ProgressBar) Redraw() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.finished {
		return
	}

	if pb.isFloating {
		pb.renderFloatingProgress("继续爆破中...")
	} else {
		pb.render("继续爆破中...")
	}
}

// ForceRedraw 强制重新绘制进度条（用于日志输出后）
func (pb *ProgressBar) ForceRedraw() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.finished {
		return
	}

	if pb.isFloating {
		// 立即重新渲染底部进度条，不需要延迟
		if pb.current >= 0 {
			pb.renderFloatingProgress("继续爆破中...")
		}
	} else {
		// 简单重绘
		if pb.lastLine != "" {
			fmt.Print(pb.lastLine)
		}
	}
}

// renderFloatingProgress 渲染浮动进度条（固定在底部，与日志完全隔离）
func (pb *ProgressBar) renderFloatingProgress(message string) {
	if pb.total <= 0 {
		return
	}

	// 重新获取终端尺寸（可能会变化）
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = pb.termWidth, pb.termHeight // 使用缓存的尺寸
	} else {
		pb.termWidth = width
		pb.termHeight = height
	}

	// 计算百分比
	percentage := float64(pb.current) / float64(pb.total) * 100

	// 计算进度条长度（固定长度，确保格式一致）
	barWidth := 50 // 固定长度，与示例保持一致

	// 构建进度条 - 始终显示为空状态
	bar := strings.Repeat("░", barWidth)

	// 计算耗时和预估剩余时间
	elapsed := time.Since(pb.startTime)
	var eta string
	if pb.current > 0 {
		avgTime := elapsed / time.Duration(pb.current)
		remaining := time.Duration(pb.total-pb.current) * avgTime
		eta = formatDuration(remaining)
	} else {
		eta = "计算中..."
	}

	// 构建进度条文本，按照期望格式
	progressText := fmt.Sprintf("🔓 爆破进度 [%s] %d/%d (%.1f%%) 剩余: %s - %s",
		bar, pb.current, pb.total, percentage, eta, message)

	// 保存当前行内容
	pb.lastLine = progressText

	// 直接定位到底部最后一行，完全独立显示
	fmt.Printf("\033[%d;1H\033[K%s\033[1G", height, progressText)
}

// clearBottomLine 清除底部行
func (pb *ProgressBar) clearBottomLine() {
	if pb.termHeight > 0 {
		fmt.Printf("\033[%d;1H\033[K", pb.termHeight)
	}
}

// render 渲染进度条（普通模式）
func (pb *ProgressBar) render(message string) {
	if pb.total <= 0 {
		return
	}

	// 计算百分比
	percentage := float64(pb.current) / float64(pb.total) * 100

	// 构建进度条 - 始终显示为空状态（固定长度）
	bar := strings.Repeat("░", 50)

	// 计算耗时和预估剩余时间
	elapsed := time.Since(pb.startTime)
	var eta string
	if pb.current > 0 {
		avgTime := elapsed / time.Duration(pb.current)
		remaining := time.Duration(pb.total-pb.current) * avgTime
		eta = formatDuration(remaining)
	} else {
		eta = "计算中..."
	}

	// 构建进度条文本，按照期望格式
	progressText := fmt.Sprintf("\r🔓 爆破进度 [%s] %d/%d (%.1f%%) 剩余: %s - %s",
		bar, pb.current, pb.total, percentage, eta, message)

	// 保存当前行内容
	pb.lastLine = progressText

	// 输出进度条（使用\r回到行首，不换行）
	fmt.Print(progressText)
}

// formatDuration 格式化时间
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "0s"
	}

	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
}

// StatusDisplay 状态显示器
type StatusDisplay struct {
	startTime     time.Time
	attempts      int
	successful    int
	failed        int
	currentTarget string
	progressBar   *ProgressBar
}

// NewStatusDisplay 创建状态显示器
func NewStatusDisplay() *StatusDisplay {
	return &StatusDisplay{
		startTime: time.Now(),
	}
}

// SetProgressBar 设置进度条
func (sd *StatusDisplay) SetProgressBar(pb *ProgressBar) {
	sd.progressBar = pb
}

// UpdateAttempt 更新尝试状态
func (sd *StatusDisplay) UpdateAttempt(username, password string, success bool) {
	sd.attempts++
	sd.currentTarget = fmt.Sprintf("%s:%s", username, password)

	// 这里不输出成功/失败消息，让爆破引擎通过logger输出
	if success {
		sd.successful++
	} else {
		sd.failed++
	}
}

// LogMessage 输出日志消息（与进度条完全隔离）
func (sd *StatusDisplay) LogMessage(message string) {
	// 在浮动模式下，确保日志输出在进度条上方，完全不干扰
	if sd.progressBar != nil && sd.progressBar.isFloating {
		// 日志输出到标准输出，让它自然滚动
		if message != "" {
			fmt.Println(message)
		}

		// 立即重绘进度条到底部（确保位置固定）
		sd.progressBar.ForceRedraw()
	} else {
		// 非浮动模式直接输出
		if message != "" {
			fmt.Println(message)
		}
	}
}

// ShowSummary 显示摘要信息
func (sd *StatusDisplay) ShowSummary() {
	// 清除浮动进度条
	if sd.progressBar != nil && sd.progressBar.isFloating {
		sd.progressBar.clearBottomLine()
	}

	elapsed := time.Since(sd.startTime)
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("📊 爆破摘要报告\n")
	fmt.Printf("总耗时: %s\n", formatDuration(elapsed))
	fmt.Printf("总尝试: %d 次\n", sd.attempts)
	fmt.Printf("成功: %d 次\n", sd.successful)
	fmt.Printf("失败: %d 次\n", sd.failed)
	if sd.attempts > 0 {
		fmt.Printf("成功率: %.1f%%\n", float64(sd.successful)/float64(sd.attempts)*100)
		fmt.Printf("平均速度: %.1f 次/秒\n", float64(sd.attempts)/elapsed.Seconds())
	}
	fmt.Println(strings.Repeat("=", 60))
}

// ProgressManager 进度条管理器
type ProgressManager struct {
	progressBar *ProgressBar
	mu          sync.Mutex
}

// LogWithProgress 输出日志并正确处理进度条（完全隔离）
func (pm *ProgressManager) LogWithProgress(logFunc func()) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 浮动模式下，确保日志和进度条完全分离
	if pm.progressBar != nil && pm.progressBar.isFloating {
		// 执行日志输出，自动在进度条上方显示
		logFunc()

		// 立即重新绘制进度条到底部，保持固定位置
		pm.progressBar.ForceRedraw()
	} else {
		// 非浮动模式直接执行
		logFunc()
	}
}

// LiveStatus 实时状态显示
type LiveStatus struct {
	isRunning bool
}

// Start 开始显示状态
func (ls *LiveStatus) Start() {
	ls.isRunning = true
	go ls.animate()
}

// Stop 停止显示状态
func (ls *LiveStatus) Stop() {
	ls.isRunning = false
}

// animate 动画效果
func (ls *LiveStatus) animate() {
	chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0

	for ls.isRunning {
		fmt.Printf("\r%s 正在执行中...", chars[i%len(chars)])
		i++
		time.Sleep(100 * time.Millisecond)
	}
}
