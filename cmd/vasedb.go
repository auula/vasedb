package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/auula/vasedb/clog"
	"github.com/auula/vasedb/conf"
	"github.com/auula/vasedb/disk"
	"github.com/auula/vasedb/server"
	"github.com/fatih/color"
)

const (
	version = "v0.1.1"
	website = "https://vasedb.github.io"
)

var (
	//go:embed banner.txt
	banner    string
	greenFont = color.New(color.FgMagenta)
	logo      = greenFont.Sprintf(banner, version, website)
	daemon    = false
)

// 初始化全局需要使用的组件
func init() {
	// 打印 Banner 信息
	fmt.Println(logo)

	// 解析命令行输入的参数，默认命令行参数优先级最高，但是相对于能设置参数比较少
	fl := parseFlags()

	// 根据命令行传入的配置文件地址，覆盖掉默认的配置
	if fl.config != conf.DefaultPath {
		if err := conf.Load(fl.config, conf.Settings); err != nil {
			clog.Failed(err)
		}
	}

	if fl.debug {
		conf.Settings.Debug = fl.debug
		// 覆盖掉默认 DEBUG 选项
		clog.IsDebug = conf.Settings.Debug
	}

	if fl.auth != conf.DefaultConfig.Password {
		conf.Settings.Password = fl.auth
	}

	if fl.path != conf.DefaultConfig.Path {
		conf.Settings.Path = fl.path
	}

	if int32(fl.port) != conf.DefaultConfig.Port {
		conf.Settings.Port = int32(fl.port)
	}

	clog.Debug(fmt.Sprintf("%+v\n", conf.Settings))

	// 设置一下运行过程中日志输出文件的路径
	clog.SetPath(conf.Settings.Logging)

	if err := disk.InitFS(conf.Settings.Path); err != nil {
		clog.Failed(err)
	}

	// 是否启用守护进程
	daemon = fl.daemon
}

// flags 优先级别最高的参数，从命令行传入
type flags struct {
	auth   string
	port   int
	path   string
	config string
	debug  bool
	daemon bool
}

// parseFlags 解析从命令行启动输入的主要参数
func parseFlags() (fl *flags) {
	fl = new(flags)
	flag.StringVar(&fl.auth, "auth", conf.DefaultConfig.Password, "--auth specify the server authentication password.")
	flag.StringVar(&fl.config, "config", conf.DefaultPath, "--config specify the configuration file path.")
	flag.StringVar(&fl.path, "path", conf.DefaultConfig.Path, "--path specify the data storage directory.")
	flag.IntVar(&fl.port, "port", int(conf.DefaultConfig.Port), "--port specify the HTTP server port.")
	flag.BoolVar(&fl.debug, "debug", conf.DefaultConfig.Debug, "--debug whether to enable debug mode.")
	flag.BoolVar(&fl.daemon, "daemon", false, "--daemon whether to run with a daemon.")
	flag.Parse()
	return
}

// trimDaemon 从 os.Args 中移除 "-daemon" 参数
func trimDaemon(args []string) []string {
	var newArgs []string

	// 遍历 args 切片
	for i := 1; i < len(args); i++ {
		// 巨坑无比，无法知道用户输入的参数是 -- 还是 - 开头
		if args[i] == "-daemon" || args[i] == "--daemon" {
			// 当发现 "-daemon" 参数时，跳过当前参数
			continue
		}
		newArgs = append(newArgs, args[i])
	}

	return newArgs
}

func splitArgs(args []string) []string {
	var newArgs []string
	// i = 1 避免处理 ./vasedb 本身
	for i := 1; i < len(args); i++ {
		// 巨坑无比，如果没有这段代码 newArgs 元素中会出现 --port=2468 元素
		// 而不是分割好的 [--port , 2468] 这样的元素，否则命令行解析错误
		if strings.Contains(args[i], "=") && strings.Count(args[i], "=") == 1 {
			newArgs = append(newArgs, strings.Split(args[i], "=")...)
		} else {
			if strings.Count(args[i], "=") > 1 {
				// 如果是 --port==8080 ，过滤掉 == 不合法过滤掉
				continue
			}
			newArgs = append(newArgs, strings.Split(args[i], "=")...)
		}
	}

	return newArgs
}

func main() {
	// 检查是否启用了守护进程模式
	if daemon {
		// 后台守护进程模式启动，创建一个与当前程序相同的命令
		cmd := exec.Command(os.Args[0], splitArgs(trimDaemon(os.Args))...)

		// 如果需要传递环境变量信息
		cmd.Env = os.Environ()

		// 从当前进程启动守护进程
		if err := cmd.Start(); err != nil {
			clog.Failed(err)
		}

		clog.Info(fmt.Sprintf("Daemon launched PID: %d", cmd.Process.Pid))
	} else {
		// 开始执行正常的 vasedb 逻辑，这里会启动 HTTP 服务器让客户端连接
		hs := server.New(conf.Settings)
		hs.Startup()
	}
}
