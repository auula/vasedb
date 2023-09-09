package clog

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fatih/color"
)

const (
	processName = "VASEDB"
)

var (
	// Logger colors and log message prefixes
	warnColor   = color.New(color.Bold, color.FgYellow)
	infoColor   = color.New(color.Bold, color.FgGreen)
	redColor    = color.New(color.Bold, color.FgRed)
	debugColor  = color.New(color.Bold, color.FgHiMagenta)
	errorPrefix = redColor.Sprintf("[ERROR] ")
	warnPrefix  = warnColor.Sprintf("[WARN] ")
	infoPrefix  = infoColor.Sprintf("[INFO] ")
	debugPrefix = debugColor.Sprintf("[DEBUG] ")

	IsDebug = false
	caw     = os.O_CREATE | os.O_APPEND | os.O_WRONLY
)

var (
	clog *log.Logger
	dlog *log.Logger
)

func init() {
	// 总共有两套日志记录器
	// [VASEDB:C] 为主进程记录器记录正常运行状态日志信息
	// [VASEDB:D] 为辅助记录器记录为 Debug 模式下的日志信息
	clog = NewLogger(os.Stdout, "["+processName+":C] ", log.Ldate|log.Ltime)
	// [VASEDB:D] 只能输出日志信息到标准输出中
	dlog = NewLogger(os.Stdout, "["+processName+":D] ", log.Ldate|log.Ltime|log.Lshortfile)

}

func NewLogger(out io.Writer, prefix string, flag int) *log.Logger {
	return log.New(out, prefix, flag)
}

func NewColorLogger(out io.Writer, prefix string, flag int) {
	clog = log.New(out, prefix, flag)
}

func SetPath(path string) {
	// 如果已经存在了就直接追加,不存在就创建
	file, err := os.OpenFile(path, caw, 0755)
	if err != nil {
		Error(err)
	}
	// 正常模式的日志记录需要输出到控制台和日志文件中
	NewColorLogger(io.MultiWriter(os.Stdout, file), "["+processName+":C] ", log.Ldate|log.Ltime)
	Info("Initial logger successful")
}

func Error(v ...interface{}) {
	clog.Output(2, errorPrefix+fmt.Sprint(v...))
}

func Failed(v ...interface{}) {
	clog.Output(2, errorPrefix+fmt.Sprint(v...))
	os.Exit(1)
}

func Warn(v ...interface{}) {
	clog.Output(2, warnPrefix+fmt.Sprint(v...))
}

func Info(v ...interface{}) {
	clog.Output(2, infoPrefix+fmt.Sprint(v...))
}

func Debug(v ...interface{}) {
	if IsDebug {
		dlog.Output(2, debugPrefix+fmt.Sprint(v...))
	}
}
